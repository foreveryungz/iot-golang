package sensor

import "time"

type CreateRequest struct {
	DeviceID   uint   `json:"device_id" binding:"required"`
	SensorCode string `json:"sensor_code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Unit       string `json:"unit"`
	Status     string `json:"status"`
}

type DeviceResponse struct {
	ID         uint   `json:"id"`
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	Status     string `json:"status"`
}

type DetailResponse struct {
	ID         uint           `json:"id"`
	DeviceID   uint           `json:"device_id"`
	SensorCode string         `json:"sensor_code"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Unit       string         `json:"unit"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Device     DeviceResponse `json:"device" gorm:"embedded;embeddedPrefix:device__"`
}

type UpdateRequest struct {
	DeviceID   uint   `json:"device_id" binding:"required"`
	SensorCode string `json:"sensor_code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Unit       string `json:"unit"`
	Status     string `json:"status"`
}
