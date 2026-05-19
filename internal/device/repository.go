package device

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(device *Device) error
	FindAll() ([]Device, error)
	FindByID(id string) (*Device, error)
	Update(device *Device) error
	Delete(device *Device) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(device *Device) error {
	err := r.db.Create(device).Error
	if isDuplicateConstraintError(err) {
		return ErrDuplicateDeviceCode
	}

	return err
}

func (r *repository) FindAll() ([]Device, error) {
	var devices []Device
	err := r.db.Preload("Sensors").Order("id desc").Find(&devices).Error
	return devices, err
}

func (r *repository) FindByID(id string) (*Device, error) {
	var device Device
	err := r.db.Preload("Sensors").First(&device, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDeviceNotFound
	}
	return &device, err
}

func (r *repository) Update(device *Device) error {
	err := r.db.Save(device).Error
	if isDuplicateConstraintError(err) {
		return ErrDuplicateDeviceCode
	}

	return err
}

func (r *repository) Delete(device *Device) error {
	return r.db.Delete(device).Error
}
