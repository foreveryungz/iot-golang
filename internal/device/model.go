package device

import (
	"time"
)

type SensorRelation struct {
	ID         uint   `json:"id"`
	DeviceID   uint   `gorm:"column:device_id" json:"-"`
	SensorCode string `json:"sensor_code"`
	Name       string `json:"name"`
}

func (SensorRelation) TableName() string {
	return "sensors"
}

type Device struct {
	ID         uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceCode string           `gorm:"uniqueIndex;not null" json:"device_code"`
	Name       string           `gorm:"not null" json:"name"`
	Location   string           `json:"location"`
	Status     string           `gorm:"default:inactive" json:"status"`
	Sensors    []SensorRelation `gorm:"foreignKey:DeviceID;references:ID" json:"sensors"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
