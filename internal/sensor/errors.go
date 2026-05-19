package sensor

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrSensorNotFound      = errors.New("Sensor not found")
	ErrDuplicateSensorCode = errors.New("duplicate sensor_code")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInvalidDeviceID     = errors.New("invalid device_id")
)

func isDuplicateConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}

func isValidStatus(status string) bool {
	return status == "active" || status == "inactive"
}
