package sensor

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(sensor *Sensor) error
	FindAll() ([]Sensor, error)
	FindByID(id string) (*Sensor, error)
	Update(sensor *Sensor) error
	Delete(sensor *Sensor) error
	DeviceExists(id uint) (bool, error)
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

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(device *Sensor) error {
	err := r.db.Create(device).Error
	if isDuplicateConstraintError(err) {
		return ErrDuplicateSensorCode
	}

	return err
}

func (r *repository) FindAll() ([]Sensor, error) {
	var sensors []Sensor
	err := r.db.Preload("Device").Order("id desc").Find(&sensors).Error
	return sensors, err
}

func (r *repository) FindByID(id string) (*Sensor, error) {
	var sensor Sensor
	err := r.db.Preload("Device").First(&sensor, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSensorNotFound
	}
	return &sensor, err
}

func (r *repository) Update(device *Sensor) error {
	err := r.db.Save(device).Error
	if isDuplicateConstraintError(err) {
		return ErrDuplicateSensorCode
	}

	return err
}

func (r *repository) Delete(device *Sensor) error {
	return r.db.Delete(device).Error
}

func (r *repository) DeviceExists(id uint) (bool, error) {
	var record deviceLookup
	err := r.db.Select("id").First(&record, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return err == nil, err
}
