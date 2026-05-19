package delivery

import (
	"encoding/json"
	"errors"
	"iot-golang/internal/monitoring"
	"iot-golang/pkg/httpclient"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Service interface {
	ProcessPending() (int, error)
	ProcessRetryable() (int, error)
	GetDeliveryLogs() ([]DeliveryLog, error)
	SendMonitoringData(data monitoring.MonitoringData) error
}

type service struct {
	repo       Repository
	targetURL  string
	httpClient httpclient.Client
}

func NewService(repo Repository, targetURL string, httpClient httpclient.Client) Service {
	return &service{
		repo:       repo,
		targetURL:  targetURL,
		httpClient: httpClient,
	}
}

const MaxRetryCount = 3
const RetryDelay = 1 * time.Minute

func (s *service) ProcessPending() (int, error) {
	items, err := s.repo.FindPendingReady(time.Now())
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if err := s.SendMonitoringData(item); err != nil {
			continue
		}
	}

	return len(items), nil
}

func (s *service) ProcessRetryable() (int, error) {
	items, err := s.repo.FindRetryable(time.Now(), MaxRetryCount)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if err := s.SendMonitoringData(item); err != nil {
			continue
		}
	}

	return len(items), nil
}

func (s *service) GetDeliveryLogs() ([]DeliveryLog, error) {
	return s.repo.FindAllLogs()
}

func (s *service) SendMonitoringData(data monitoring.MonitoringData) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.PostJSON(
		s.targetURL,
		payload,
		map[string]string{},
	)

	attempt := data.RetryCount + 1

	if err != nil {
		if handleErr := s.handleDeliveryError(data.ID, attempt, 0, err.Error(), isRetryableError(err)); handleErr != nil {
			return handleErr
		}
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if handleErr := s.handleDeliveryError(data.ID, attempt, resp.StatusCode, resp.Status, isRetryableStatusCode(resp.StatusCode)); handleErr != nil {
			return handleErr
		}
		return nil
	}

	now := time.Now()
	return s.repo.SaveDeliveryResult(DeliveryResult{
		MonitoringDataID: data.ID,
		Log: DeliveryLog{
			MonitoringDataID: data.ID,
			TargetURL:        s.targetURL,
			Status:           "sent",
			StatusCode:       resp.StatusCode,
			ErrorMessage:     "",
			Attempt:          attempt,
		},
		Updates: map[string]interface{}{
			"status":        "sent",
			"sent_at":       &now,
			"next_retry_at": nil,
			"retry_count":   data.RetryCount,
		},
	})
}

func isRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

func isRetryableError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	return errors.As(err, &urlErr)
}

func (s *service) handleDeliveryError(
	monitoringDataID uint,
	attempt int,
	statusCode int,
	errorMessage string,
	retryable bool,
) error {
	if retryable {
		return s.markAsFailed(monitoringDataID, attempt, statusCode, errorMessage)
	}

	return s.markAsDeadLetter(monitoringDataID, attempt, statusCode, errorMessage)
}

func (s *service) markAsFailed(
	monitoringDataID uint,
	attempt int,
	statusCode int,
	errorMessage string,
) error {
	if attempt >= MaxRetryCount {
		return s.markAsDeadLetter(monitoringDataID, attempt, statusCode, errorMessage)
	}

	nextRetryAt := time.Now().Add(RetryDelay)

	return s.repo.SaveDeliveryResult(DeliveryResult{
		MonitoringDataID: monitoringDataID,
		Log: DeliveryLog{
			MonitoringDataID: monitoringDataID,
			TargetURL:        s.targetURL,
			Status:           "failed",
			StatusCode:       statusCode,
			ErrorMessage:     errorMessage,
			Attempt:          attempt,
		},
		Updates: map[string]interface{}{
			"status":        "failed",
			"sent_at":       nil,
			"retry_count":   attempt,
			"next_retry_at": &nextRetryAt,
		},
	})
}

func (s *service) markAsDeadLetter(
	monitoringDataID uint,
	attempt int,
	statusCode int,
	errorMessage string,
) error {
	return s.repo.SaveDeliveryResult(DeliveryResult{
		MonitoringDataID: monitoringDataID,
		Log: DeliveryLog{
			MonitoringDataID: monitoringDataID,
			TargetURL:        s.targetURL,
			Status:           "failed",
			StatusCode:       statusCode,
			ErrorMessage:     errorMessage,
			Attempt:          attempt,
		},
		Updates: map[string]interface{}{
			"status":        "dead_letter",
			"sent_at":       nil,
			"retry_count":   attempt,
			"next_retry_at": nil,
		},
	})
}
