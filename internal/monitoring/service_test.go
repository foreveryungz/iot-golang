package monitoring

import (
	"testing"

	"iot-golang/internal/device"
	"iot-golang/internal/sensor"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateMonitoringDataSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice, testSensor := seedTestDeviceAndSensor(t, db)
	svc := NewService(NewRepository(db))

	data, err := svc.Create(&CreateRequest{
		DeviceID: testDevice.ID,
		SensorID: testSensor.ID,
		Value:    27.5,
		Unit:     "C",
		Status:   "pending",
	})

	assert.NoError(t, err)
	assert.NotZero(t, data.ID)
	assert.Equal(t, testDevice.ID, data.DeviceID)
	assert.Equal(t, testSensor.ID, data.SensorID)
	assert.Equal(t, 27.5, data.Value)
	assert.Equal(t, "C", data.Unit)

	stored := fetchMonitoringDataByID(t, db, data.ID)
	assert.Equal(t, testDevice.ID, stored.DeviceID)
	assert.Equal(t, testSensor.ID, stored.SensorID)
	assert.Equal(t, 27.5, stored.Value)
	assert.Equal(t, "C", stored.Unit)
}

func TestCreateMonitoringDataDefaultStatusPending(t *testing.T) {
	db := setupTestDB(t)
	testDevice, testSensor := seedTestDeviceAndSensor(t, db)
	svc := NewService(NewRepository(db))

	data, err := svc.Create(&CreateRequest{
		DeviceID: testDevice.ID,
		SensorID: testSensor.ID,
		Value:    58.2,
		Unit:     "%",
	})

	assert.NoError(t, err)
	assert.Equal(t, "pending", data.Status)

	stored := fetchMonitoringDataByID(t, db, data.ID)
	assert.Equal(t, "pending", stored.Status)
}

func TestFindAllMonitoringDataSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice, testSensor := seedTestDeviceAndSensor(t, db)
	svc := NewService(NewRepository(db))

	seedMonitoringData(t, db, MonitoringData{
		DeviceID: testDevice.ID,
		SensorID: testSensor.ID,
		Value:    10.5,
		Unit:     "C",
		Status:   "pending",
	})
	seedMonitoringData(t, db, MonitoringData{
		DeviceID: testDevice.ID,
		SensorID: testSensor.ID,
		Value:    11.5,
		Unit:     "C",
		Status:   "sent",
	})

	items, err := svc.FindAll()

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 11.5, items[0].Value)
	assert.Equal(t, 10.5, items[1].Value)
}

func TestCreateMonitoringDataWithInvalidDeviceID(t *testing.T) {
	db := setupTestDB(t)
	_, testSensor := seedTestDeviceAndSensor(t, db)
	svc := NewService(NewRepository(db))

	data, err := svc.Create(&CreateRequest{
		DeviceID: 999999,
		SensorID: testSensor.ID,
		Value:    33.3,
		Unit:     "C",
	})

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.ErrorIs(t, err, ErrDeviceNotFound)

	var count int64
	db.Model(&MonitoringData{}).Where("sensor_id = ? AND value = ?", testSensor.ID, 33.3).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCreateMonitoringDataWithInvalidSensorID(t *testing.T) {
	db := setupTestDB(t)
	testDevice, _ := seedTestDeviceAndSensor(t, db)
	svc := NewService(NewRepository(db))

	data, err := svc.Create(&CreateRequest{
		DeviceID: testDevice.ID,
		SensorID: 999999,
		Value:    44.4,
		Unit:     "C",
	})

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.ErrorIs(t, err, ErrSensorNotFound)

	var count int64
	db.Model(&MonitoringData{}).Where("device_id = ? AND value = ?", testDevice.ID, 44.4).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCreateMonitoringDataWithSensorFromAnotherDeviceFails(t *testing.T) {
	db := setupTestDB(t)
	testDevice, _ := seedTestDeviceAndSensor(t, db)
	otherDevice := seedDevice(t, db, device.Device{
		DeviceCode: "DEV-MON-002",
		Name:       "Secondary Gateway",
		Location:   "Plant 2",
		Status:     "active",
	})
	otherSensor := seedSensor(t, db, sensor.Sensor{
		DeviceID:   otherDevice.ID,
		SensorCode: "SNS-MON-002",
		Name:       "Pressure Probe",
		Type:       "pressure",
		Unit:       "bar",
		Status:     "active",
	})

	svc := NewService(NewRepository(db))

	data, err := svc.Create(&CreateRequest{
		DeviceID: testDevice.ID,
		SensorID: otherSensor.ID,
		Value:    55.5,
		Unit:     "bar",
	})

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.ErrorIs(t, err, ErrSensorDeviceMismatch)

	var count int64
	db.Model(&MonitoringData{}).
		Where("device_id = ? AND sensor_id = ? AND value = ?", testDevice.ID, otherSensor.ID, 55.5).
		Count(&count)
	assert.Equal(t, int64(0), count)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&device.Device{}, &sensor.Sensor{}, &MonitoringData{}); err != nil {
		t.Fatalf("migrate test schema: %v", err)
	}

	return db
}

func seedTestDeviceAndSensor(t *testing.T, db *gorm.DB) (device.Device, sensor.Sensor) {
	t.Helper()

	testDevice := device.Device{
		DeviceCode: "DEV-MON-001",
		Name:       "Monitoring Gateway",
		Location:   "Plant 1",
		Status:     "active",
	}
	testDevice = seedDevice(t, db, testDevice)

	testSensor := sensor.Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-MON-001",
		Name:       "Temperature Probe",
		Type:       "temperature",
		Unit:       "C",
		Status:     "active",
	}
	testSensor = seedSensor(t, db, testSensor)

	return testDevice, testSensor
}

func seedDevice(t *testing.T, db *gorm.DB, item device.Device) device.Device {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed device: %v", err)
	}

	return item
}

func seedSensor(t *testing.T, db *gorm.DB, item sensor.Sensor) sensor.Sensor {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed sensor: %v", err)
	}

	return item
}

func seedMonitoringData(t *testing.T, db *gorm.DB, data MonitoringData) MonitoringData {
	t.Helper()

	if err := db.Create(&data).Error; err != nil {
		t.Fatalf("seed monitoring data: %v", err)
	}

	return data
}

func fetchMonitoringDataByID(t *testing.T, db *gorm.DB, id uint) MonitoringData {
	t.Helper()

	var data MonitoringData
	if err := db.First(&data, id).Error; err != nil {
		t.Fatalf("fetch monitoring data: %v", err)
	}

	return data
}
