package sensor

import (
	"time"
)

type DeviceRelation struct {
	ID         uint   `json:"id"`
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
}

func (DeviceRelation) TableName() string {
	return "devices"
}

type Sensor struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID   uint           `gorm:"not null;index" json:"device_id"`
	SensorCode string         `gorm:"uniqueIndex;not null" json:"sensor_code"`
	Name       string         `gorm:"not null" json:"name"`
	Type       string         `gorm:"not null" json:"type"`
	Unit       string         `json:"unit"`
	Status     string         `gorm:"default:inactive" json:"status"`
	Device     DeviceRelation `gorm:"foreignKey:DeviceID;references:ID" json:"device"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
