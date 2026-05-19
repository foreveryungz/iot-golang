package delivery

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"iot-golang/internal/monitoring"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockHTTPClient struct {
	statusCode int
	err        error
}

func (m *mockHTTPClient) PostJSON(url string, payload []byte, headers map[string]string) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &http.Response{
		Status:     fmt.Sprintf("%d %s", m.statusCode, http.StatusText(m.statusCode)),
		StatusCode: m.statusCode,
		Body:       http.NoBody,
	}, nil
}

func TestRetryableStatusCode(t *testing.T) {
	assert.True(t, isRetryableStatusCode(500))
	assert.True(t, isRetryableStatusCode(503))
	assert.True(t, isRetryableStatusCode(429))

	assert.False(t, isRetryableStatusCode(400))
	assert.False(t, isRetryableStatusCode(401))
	assert.False(t, isRetryableStatusCode(404))
}

func TestRetryableError(t *testing.T) {
	err := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: errors.New("connection refused"),
	}

	assert.True(t, isRetryableError(err))
	assert.False(t, isRetryableError(nil))
}

func TestSendMonitoringDataMarksSentOnHTTP200(t *testing.T) {
	// This validates the success path: a 2xx response marks the record as sent,
	// clears retry scheduling metadata, and writes a sent delivery log.
	db := newDeliveryTestDB(t)
	record := createMonitoringRecord(t, db, monitoring.MonitoringData{
		DeviceID:   1,
		SensorID:   1,
		Value:      42.5,
		Unit:       "C",
		Status:     "pending",
		RetryCount: 1,
	})

	svc := NewService(NewRepository(db), "http://example.com/webhook", &mockHTTPClient{
		statusCode: http.StatusOK,
	})

	err := svc.SendMonitoringData(record)

	assert.NoError(t, err)

	updated := fetchMonitoringRecord(t, db, record.ID)
	logs := fetchDeliveryLogs(t, db, record.ID)

	assert.Equal(t, "sent", updated.Status)
	assert.NotNil(t, updated.SentAt)
	assert.Nil(t, updated.NextRetryAt)
	assert.Equal(t, record.RetryCount, updated.RetryCount)

	assert.Len(t, logs, 1)
	assert.Equal(t, "sent", logs[0].Status)
	assert.Equal(t, http.StatusOK, logs[0].StatusCode)
	assert.Equal(t, record.RetryCount+1, logs[0].Attempt)
	assert.Empty(t, logs[0].ErrorMessage)
}

func TestSendMonitoringDataSchedulesRetryOnHTTP500(t *testing.T) {
	// This validates retryable failure handling: a 5xx response increments the
	// retry counter, leaves the record failed, schedules the next retry, and logs
	// the failed attempt for the retry loop to pick up later.
	db := newDeliveryTestDB(t)
	record := createMonitoringRecord(t, db, monitoring.MonitoringData{
		DeviceID:   1,
		SensorID:   1,
		Value:      10,
		Unit:       "C",
		Status:     "pending",
		RetryCount: 0,
	})

	svc := NewService(NewRepository(db), "http://example.com/webhook", &mockHTTPClient{
		statusCode: http.StatusInternalServerError,
	})

	before := time.Now()
	err := svc.SendMonitoringData(record)
	after := time.Now()

	assert.NoError(t, err)

	updated := fetchMonitoringRecord(t, db, record.ID)
	logs := fetchDeliveryLogs(t, db, record.ID)

	assert.Equal(t, "failed", updated.Status)
	assert.Nil(t, updated.SentAt)
	assert.Equal(t, record.RetryCount+1, updated.RetryCount)
	assert.NotNil(t, updated.NextRetryAt)
	assert.True(t, updated.NextRetryAt.After(before.Add(RetryDelay).Add(-2*time.Second)))
	assert.True(t, updated.NextRetryAt.Before(after.Add(RetryDelay).Add(2*time.Second)))

	assert.Len(t, logs, 1)
	assert.Equal(t, "failed", logs[0].Status)
	assert.Equal(t, http.StatusInternalServerError, logs[0].StatusCode)
	assert.Equal(t, record.RetryCount+1, logs[0].Attempt)
	assert.Contains(t, logs[0].ErrorMessage, "500")
}

func TestSendMonitoringDataStopsRetryLoopOnHTTP400(t *testing.T) {
	// This validates non-retryable failure handling: a 4xx response should stop
	// retry scheduling, clear next_retry_at, and move the record to its terminal
	// non-retryable state while still writing a failed delivery log.
	db := newDeliveryTestDB(t)
	record := createMonitoringRecord(t, db, monitoring.MonitoringData{
		DeviceID:   1,
		SensorID:   1,
		Value:      18,
		Unit:       "C",
		Status:     "pending",
		RetryCount: 1,
	})

	svc := NewService(NewRepository(db), "http://example.com/webhook", &mockHTTPClient{
		statusCode: http.StatusBadRequest,
	})

	err := svc.SendMonitoringData(record)

	assert.NoError(t, err)

	updated := fetchMonitoringRecord(t, db, record.ID)
	logs := fetchDeliveryLogs(t, db, record.ID)

	assert.Equal(t, "dead_letter", updated.Status)
	assert.Nil(t, updated.SentAt)
	assert.Nil(t, updated.NextRetryAt)
	assert.Equal(t, record.RetryCount+1, updated.RetryCount)

	assert.Len(t, logs, 1)
	assert.Equal(t, "failed", logs[0].Status)
	assert.Equal(t, http.StatusBadRequest, logs[0].StatusCode)
	assert.Equal(t, record.RetryCount+1, logs[0].Attempt)
	assert.Contains(t, logs[0].ErrorMessage, "400")
}

func TestSendMonitoringDataMovesToDeadLetterAtMaxRetryCount(t *testing.T) {
	// This validates the retry ceiling: once the next failed attempt reaches the
	// max retry count, the record becomes dead_letter and retry scheduling stops.
	db := newDeliveryTestDB(t)
	record := createMonitoringRecord(t, db, monitoring.MonitoringData{
		DeviceID:   1,
		SensorID:   1,
		Value:      99,
		Unit:       "C",
		Status:     "failed",
		RetryCount: MaxRetryCount - 1,
	})

	svc := NewService(NewRepository(db), "http://example.com/webhook", &mockHTTPClient{
		statusCode: http.StatusInternalServerError,
	})

	err := svc.SendMonitoringData(record)

	assert.NoError(t, err)

	updated := fetchMonitoringRecord(t, db, record.ID)
	logs := fetchDeliveryLogs(t, db, record.ID)

	assert.Equal(t, "dead_letter", updated.Status)
	assert.Nil(t, updated.SentAt)
	assert.Nil(t, updated.NextRetryAt)
	assert.Equal(t, MaxRetryCount, updated.RetryCount)

	assert.Len(t, logs, 1)
	assert.Equal(t, "failed", logs[0].Status)
	assert.Equal(t, http.StatusInternalServerError, logs[0].StatusCode)
	assert.Equal(t, MaxRetryCount, logs[0].Attempt)
}

func TestSaveDeliveryResultRollsBackWhenLogInsertFails(t *testing.T) {
	db := newDeliveryTestDB(t)
	record := createMonitoringRecord(t, db, monitoring.MonitoringData{
		DeviceID: 1,
		SensorID: 1,
		Value:    12,
		Status:   "pending",
	})

	repo := NewRepository(db)

	callbackName := "test:force_delivery_log_failure"
	err := db.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "DeliveryLog" {
			tx.AddError(errors.New("forced delivery log failure"))
		}
	})
	if err != nil {
		t.Fatalf("register create callback: %v", err)
	}
	defer db.Callback().Create().Remove(callbackName)

	now := time.Now()
	err = repo.SaveDeliveryResult(DeliveryResult{
		MonitoringDataID: record.ID,
		Log: DeliveryLog{
			MonitoringDataID: record.ID,
			TargetURL:        "http://example.com/webhook",
			Status:           "sent",
			StatusCode:       http.StatusOK,
			Attempt:          1,
		},
		Updates: map[string]interface{}{
			"status":        "sent",
			"sent_at":       &now,
			"next_retry_at": nil,
		},
	})

	assert.Error(t, err)

	updated := fetchMonitoringRecord(t, db, record.ID)
	logs := fetchDeliveryLogs(t, db, record.ID)

	assert.Equal(t, "pending", updated.Status)
	assert.Nil(t, updated.SentAt)
	assert.Len(t, logs, 0)
}

func TestSaveDeliveryResultRollsBackWhenMonitoringUpdateFails(t *testing.T) {
	db := newDeliveryTestDB(t)
	record := createMonitoringRecord(t, db, monitoring.MonitoringData{
		DeviceID: 1,
		SensorID: 1,
		Value:    12,
		Status:   "pending",
	})

	repo := NewRepository(db)

	callbackName := "test:force_monitoring_update_failure"
	err := db.Callback().Update().Before("gorm:update").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "MonitoringData" {
			tx.AddError(errors.New("forced monitoring update failure"))
		}
	})
	if err != nil {
		t.Fatalf("register update callback: %v", err)
	}
	defer db.Callback().Update().Remove(callbackName)

	now := time.Now()
	err = repo.SaveDeliveryResult(DeliveryResult{
		MonitoringDataID: record.ID,
		Log: DeliveryLog{
			MonitoringDataID: record.ID,
			TargetURL:        "http://example.com/webhook",
			Status:           "sent",
			StatusCode:       http.StatusOK,
			Attempt:          1,
		},
		Updates: map[string]interface{}{
			"status":        "sent",
			"sent_at":       &now,
			"next_retry_at": nil,
		},
	})

	assert.Error(t, err)

	updated := fetchMonitoringRecord(t, db, record.ID)
	logs := fetchDeliveryLogs(t, db, record.ID)

	assert.Equal(t, "pending", updated.Status)
	assert.Nil(t, updated.SentAt)
	assert.Len(t, logs, 0)
}

func newDeliveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	err = db.AutoMigrate(&monitoring.MonitoringData{}, &DeliveryLog{})
	if err != nil {
		t.Fatalf("migrate sqlite schema: %v", err)
	}

	return db
}

func createMonitoringRecord(t *testing.T, db *gorm.DB, data monitoring.MonitoringData) monitoring.MonitoringData {
	t.Helper()

	err := db.Create(&data).Error
	if err != nil {
		t.Fatalf("create monitoring record: %v", err)
	}

	return data
}

func fetchMonitoringRecord(t *testing.T, db *gorm.DB, id uint) monitoring.MonitoringData {
	t.Helper()

	var data monitoring.MonitoringData
	err := db.First(&data, id).Error
	if err != nil {
		t.Fatalf("fetch monitoring record: %v", err)
	}

	return data
}

func fetchDeliveryLogs(t *testing.T, db *gorm.DB, monitoringDataID uint) []DeliveryLog {
	t.Helper()

	var logs []DeliveryLog
	err := db.Where("monitoring_data_id = ?", monitoringDataID).Order("id asc").Find(&logs).Error
	if err != nil {
		t.Fatalf("fetch delivery logs: %v", err)
	}

	return logs
}
