package delivery

import (
	"iot-golang/internal/monitoring"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindPendingReady(now time.Time) ([]monitoring.MonitoringData, error)
	FindRetryable(now time.Time, maxRetry int) ([]monitoring.MonitoringData, error)
	FindAllLogs() ([]DeliveryLog, error)
	SaveDeliveryResult(result DeliveryResult) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindPendingReady(now time.Time) ([]monitoring.MonitoringData, error) {
	var items []monitoring.MonitoringData
	err := r.db.
		Where("status = ?", "pending").
		Where("scheduled_at IS NULL OR scheduled_at <= ?", now).
		Order("id asc").
		Find(&items).Error
	return items, err
}

func (r *repository) FindRetryable(now time.Time, maxRetry int) ([]monitoring.MonitoringData, error) {
	var items []monitoring.MonitoringData
	err := r.db.
		Where("status = ?", "failed").
		Where("retry_count < ?", maxRetry).
		Where("next_retry_at IS NOT NULL AND next_retry_at <= ?", now).
		Order("id asc").
		Find(&items).Error
	return items, err
}

func (r *repository) FindAllLogs() ([]DeliveryLog, error) {
	var logs []DeliveryLog
	err := r.db.Order("id desc").Find(&logs).Error
	return logs, err
}

func (r *repository) SaveDeliveryResult(result DeliveryResult) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&result.Log).Error; err != nil {
			return err
		}

		return tx.Model(&monitoring.MonitoringData{}).
			Where("id = ?", result.MonitoringDataID).
			Updates(result.Updates).Error
	})
}
