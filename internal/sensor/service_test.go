package sensor

import (
	"strconv"
	"testing"

	"iot-golang/internal/device"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateSensorSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	item, err := svc.Create(&CreateRequest{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-001",
		Name:       "Temperature Sensor",
		Type:       "temperature",
		Unit:       "C",
	})

	assert.NoError(t, err)
	assert.NotZero(t, item.ID)
	assert.Equal(t, "SNS-001", item.SensorCode)
	assert.Equal(t, "Temperature Sensor", item.Name)
	assert.Equal(t, testDevice.ID, item.DeviceID)
	assert.Equal(t, "inactive", item.Status)

	stored := fetchSensorByID(t, db, item.ID)
	assert.Equal(t, item.SensorCode, stored.SensorCode)
	assert.Equal(t, item.Name, stored.Name)
	assert.Equal(t, testDevice.ID, stored.DeviceID)
	assert.Equal(t, "inactive", stored.Status)
}

func TestFindAllSensorsSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	seedSensor(t, db, Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-001",
		Name:       "Temperature Sensor",
		Type:       "temperature",
		Unit:       "C",
		Status:     "active",
	})
	seedSensor(t, db, Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-002",
		Name:       "Humidity Sensor",
		Type:       "humidity",
		Unit:       "%",
		Status:     "inactive",
	})

	items, err := svc.FindAll()

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "SNS-002", items[0].SensorCode)
	assert.Equal(t, "SNS-001", items[1].SensorCode)
	assert.Equal(t, testDevice.ID, items[0].Device.ID)
	assert.Equal(t, testDevice.DeviceCode, items[0].Device.DeviceCode)
	assert.Equal(t, testDevice.Name, items[0].Device.Name)
}

func TestFindSensorByIDSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	seeded := seedSensor(t, db, Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-003",
		Name:       "Pressure Sensor",
		Type:       "pressure",
		Unit:       "bar",
		Status:     "active",
	})

	item, err := svc.FindByID(strconv.FormatUint(uint64(seeded.ID), 10))

	assert.NoError(t, err)
	assert.Equal(t, seeded.ID, item.ID)
	assert.Equal(t, "SNS-003", item.SensorCode)
	assert.Equal(t, "Pressure Sensor", item.Name)
	assert.Equal(t, "pressure", item.Type)
	assert.Equal(t, "bar", item.Unit)
	assert.Equal(t, "active", item.Status)
	assert.Equal(t, testDevice.ID, item.Device.ID)
	assert.Equal(t, testDevice.DeviceCode, item.Device.DeviceCode)
	assert.Equal(t, testDevice.Name, item.Device.Name)
}

func TestUpdateSensorSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	seeded := seedSensor(t, db, Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-004",
		Name:       "Flow Sensor",
		Type:       "flow",
		Unit:       "lpm",
		Status:     "inactive",
	})

	updated, err := svc.Update(strconv.FormatUint(uint64(seeded.ID), 10), &UpdateRequest{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-004",
		Name:       "Flow Sensor Updated",
		Type:       "flow-rate",
		Unit:       "m3/h",
		Status:     "active",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Flow Sensor Updated", updated.Name)
	assert.Equal(t, "flow-rate", updated.Type)
	assert.Equal(t, "m3/h", updated.Unit)
	assert.Equal(t, "active", updated.Status)

	stored := fetchSensorByID(t, db, seeded.ID)
	assert.Equal(t, "Flow Sensor Updated", stored.Name)
	assert.Equal(t, "flow-rate", stored.Type)
	assert.Equal(t, "m3/h", stored.Unit)
	assert.Equal(t, "active", stored.Status)
}

func TestDeleteSensorSuccess(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	seeded := seedSensor(t, db, Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-005",
		Name:       "Level Sensor",
		Type:       "level",
		Unit:       "cm",
		Status:     "active",
	})

	err := svc.Delete(strconv.FormatUint(uint64(seeded.ID), 10))

	assert.NoError(t, err)

	_, findErr := svc.FindByID(strconv.FormatUint(uint64(seeded.ID), 10))
	assert.ErrorIs(t, findErr, ErrSensorNotFound)
}

func TestCreateSensorWithInvalidDeviceID(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	item, err := svc.Create(&CreateRequest{
		DeviceID:   999999,
		SensorCode: "SNS-ERR-001",
		Name:       "Orphan Sensor",
		Type:       "temperature",
		Unit:       "C",
		Status:     "active",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDeviceID)
	assert.Nil(t, item)

	var count int64
	db.Model(&Sensor{}).Where("sensor_code = ?", "SNS-ERR-001").Count(&count)

	assert.Equal(t, int64(0), count)
}

func TestFindSensorByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	item, err := svc.FindByID("999999")

	assert.Nil(t, item)
	assert.ErrorIs(t, err, ErrSensorNotFound)
}

func TestCreateSensorDuplicateCodeFails(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	seedSensor(t, db, Sensor{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-DUP-001",
		Name:       "Original Sensor",
		Type:       "temperature",
		Unit:       "C",
		Status:     "active",
	})

	item, err := svc.Create(&CreateRequest{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-DUP-001",
		Name:       "Duplicate Sensor",
		Type:       "temperature",
		Unit:       "C",
		Status:     "active",
	})

	assert.ErrorIs(t, err, ErrDuplicateSensorCode)
	assert.NotNil(t, item)
}

func TestCreateSensorInvalidStatusFails(t *testing.T) {
	db := setupTestDB(t)
	testDevice := seedTestDevice(t, db)
	svc := NewService(NewRepository(db))

	item, err := svc.Create(&CreateRequest{
		DeviceID:   testDevice.ID,
		SensorCode: "SNS-BAD-STATUS",
		Name:       "Bad Status Sensor",
		Type:       "temperature",
		Unit:       "C",
		Status:     "broken",
	})

	assert.Nil(t, item)
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&device.Device{}, &Sensor{}); err != nil {
		t.Fatalf("migrate test schema: %v", err)
	}

	return db
}

func seedTestDevice(t *testing.T, db *gorm.DB) device.Device {
	t.Helper()

	item := device.Device{
		DeviceCode: "DEV-SNS-001",
		Name:       "Sensor Gateway",
		Location:   "Factory Floor",
		Status:     "active",
	}

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed device: %v", err)
	}

	return item
}

func seedSensor(t *testing.T, db *gorm.DB, item Sensor) Sensor {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed sensor: %v", err)
	}

	return item
}

func fetchSensorByID(t *testing.T, db *gorm.DB, id uint) Sensor {
	t.Helper()

	var item Sensor
	if err := db.First(&item, id).Error; err != nil {
		t.Fatalf("fetch sensor: %v", err)
	}

	return item
}
