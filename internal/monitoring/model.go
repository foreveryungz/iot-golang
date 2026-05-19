package monitoring

import (
	"time"

	"iot-golang/internal/device"
	"iot-golang/internal/sensor"
)

type MonitoringData struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	DeviceID uint `gorm:"not null;index" json:"device_id"`
	SensorID uint `gorm:"not null;index" json:"sensor_id"`

	Device device.Device `gorm:"foreignKey:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Sensor sensor.Sensor `gorm:"foreignKey:SensorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`

	Value float64 `gorm:"not null" json:"value"`
	Unit  string  `json:"unit"`

	Status string `gorm:"default:pending" json:"status"`
	// pending, sent, failed, dead_letter

	ScheduledAt *time.Time `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`

	RetryCount  int        `gorm:"default:0" json:"retry_count"`
	NextRetryAt *time.Time `json:"next_retry_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
