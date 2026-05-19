package device

import (
	"strconv"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateDeviceSuccess(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	device, err := svc.Create(&CreateRequest{
		DeviceCode: "DEV-001",
		Name:       "Pump Sensor Gateway",
		Location:   "Plant A",
	})

	assert.NoError(t, err)
	assert.NotZero(t, device.ID)
	assert.Equal(t, "DEV-001", device.DeviceCode)
	assert.Equal(t, "Pump Sensor Gateway", device.Name)
	assert.Equal(t, "inactive", device.Status)

	stored := fetchDeviceByID(t, db, device.ID)
	assert.Equal(t, device.DeviceCode, stored.DeviceCode)
	assert.Equal(t, device.Name, stored.Name)
	assert.Equal(t, "inactive", stored.Status)
}

func TestFindAllDevicesSuccess(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	seedDevice(t, db, Device{
		DeviceCode: "DEV-001",
		Name:       "Boiler Monitor",
		Location:   "Zone 1",
		Status:     "active",
	})
	seedDevice(t, db, Device{
		DeviceCode: "DEV-002",
		Name:       "Cooling Monitor",
		Location:   "Zone 2",
		Status:     "inactive",
	})

	devices, err := svc.FindAll()

	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "DEV-002", devices[0].DeviceCode)
	assert.Equal(t, "DEV-001", devices[1].DeviceCode)
}

func TestFindDeviceByIDSuccess(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	seeded := seedDevice(t, db, Device{
		DeviceCode: "DEV-003",
		Name:       "Pressure Gateway",
		Location:   "Tank Farm",
		Status:     "active",
	})

	device, err := svc.FindByID(strconv.FormatUint(uint64(seeded.ID), 10))

	assert.NoError(t, err)
	assert.Equal(t, seeded.ID, device.ID)
	assert.Equal(t, "DEV-003", device.DeviceCode)
	assert.Equal(t, "Pressure Gateway", device.Name)
	assert.Equal(t, "Tank Farm", device.Location)
	assert.Equal(t, "active", device.Status)
}

func TestUpdateDeviceSuccess(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	seeded := seedDevice(t, db, Device{
		DeviceCode: "DEV-004",
		Name:       "Valve Controller",
		Location:   "Warehouse",
		Status:     "inactive",
	})

	updated, err := svc.Update(strconv.FormatUint(uint64(seeded.ID), 10), &UpdateRequest{
		DeviceCode: "DEV-004",
		Name:       "Valve Controller Updated",
		Location:   "Warehouse B",
		Status:     "active",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Valve Controller Updated", updated.Name)
	assert.Equal(t, "Warehouse B", updated.Location)
	assert.Equal(t, "active", updated.Status)

	stored := fetchDeviceByID(t, db, seeded.ID)
	assert.Equal(t, "Valve Controller Updated", stored.Name)
	assert.Equal(t, "Warehouse B", stored.Location)
	assert.Equal(t, "active", stored.Status)
}

func TestDeleteDeviceSuccess(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	seeded := seedDevice(t, db, Device{
		DeviceCode: "DEV-005",
		Name:       "Flow Meter Hub",
		Location:   "Line 3",
		Status:     "active",
	})

	err := svc.Delete(strconv.FormatUint(uint64(seeded.ID), 10))

	assert.NoError(t, err)

	_, findErr := svc.FindByID(strconv.FormatUint(uint64(seeded.ID), 10))
	assert.ErrorIs(t, findErr, ErrDeviceNotFound)
}

func TestFindDeviceByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	device, err := svc.FindByID("999999")

	assert.Nil(t, device)
	assert.ErrorIs(t, err, ErrDeviceNotFound)
}

func TestCreateDeviceDuplicateCodeFails(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	seedDevice(t, db, Device{
		DeviceCode: "DEV-DUP-001",
		Name:       "Primary Device",
		Location:   "Zone 1",
		Status:     "active",
	})

	device, err := svc.Create(&CreateRequest{
		DeviceCode: "DEV-DUP-001",
		Name:       "Duplicate Device",
		Location:   "Zone 2",
		Status:     "active",
	})

	assert.ErrorIs(t, err, ErrDuplicateDeviceCode)
	assert.NotNil(t, device)
}

func TestCreateDeviceInvalidStatusFails(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(NewRepository(db))

	device, err := svc.Create(&CreateRequest{
		DeviceCode: "DEV-BAD-STATUS",
		Name:       "Bad Status Device",
		Status:     "broken",
	})

	assert.Nil(t, device)
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&Device{}, &SensorRelation{}); err != nil {
		t.Fatalf("migrate test schema: %v", err)
	}

	return db
}

func seedDevice(t *testing.T, db *gorm.DB, device Device) Device {
	t.Helper()

	if err := db.Create(&device).Error; err != nil {
		t.Fatalf("seed device: %v", err)
	}

	return device
}

func fetchDeviceByID(t *testing.T, db *gorm.DB, id uint) Device {
	t.Helper()

	var device Device
	if err := db.First(&device, id).Error; err != nil {
		t.Fatalf("fetch device: %v", err)
	}

	return device
}
