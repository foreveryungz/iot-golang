package monitoring

import "gorm.io/gorm"

type Repository interface {
	Create(data *MonitoringData) error
	FindAll() ([]MonitoringData, error)
	DeviceExists(id uint) (bool, error)
	FindSensorByID(id uint) (*SensorLookup, error)
}

type repository struct {
	db *gorm.DB
}

type deviceLookup struct {
	ID uint
}

func (deviceLookup) TableName() string {
	return "devices"
}

type SensorLookup struct {
	ID       uint
	DeviceID uint
}

func (SensorLookup) TableName() string {
	return "sensors"
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(data *MonitoringData) error {
	return r.db.Create(data).Error
}

func (r *repository) FindAll() ([]MonitoringData, error) {
	var items []MonitoringData
	err := r.db.Order("id desc").Find(&items).Error
	return items, err
}

func (r *repository) DeviceExists(id uint) (bool, error) {
	var record deviceLookup
	err := r.db.Select("id").First(&record, id).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	return err == nil, err
}

func (r *repository) FindSensorByID(id uint) (*SensorLookup, error) {
	var sensor SensorLookup
	err := r.db.Select("id", "device_id").First(&sensor, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &sensor, nil
}
