package monitoring

import "time"

type CreateRequest struct {
	DeviceID    uint       `json:"device_id" binding:"required"`
	SensorID    uint       `json:"sensor_id" binding:"required"`
	Value       float64    `json:"value" binding:"required"`
	Unit        string     `json:"unit"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}
