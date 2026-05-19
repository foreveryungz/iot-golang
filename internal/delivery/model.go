package delivery

import "time"

type DeliveryLog struct {
	ID               uint `gorm:"primaryKey" json:"id"`
	MonitoringDataID uint `gorm:"not null;index" json:"monitoring_data_id"`

	TargetURL string `gorm:"not null" json:"target_url"`
	Status    string `gorm:"not null" json:"status"`
	// sent, failed

	StatusCode   int    `json:"status_code"`
	ErrorMessage string `json:"error_message"`

	Attempt int `gorm:"default:1" json:"attempt"`

	CreatedAt time.Time `json:"created_at"`
}

type DeliveryResult struct {
	MonitoringDataID uint
	Log              DeliveryLog
	Updates          map[string]interface{}
}
