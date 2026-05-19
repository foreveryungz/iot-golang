package device

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDeviceNotFound      = errors.New("Device not found")
	ErrDuplicateDeviceCode = errors.New("duplicate device_code")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidStatus       = errors.New("invalid status")
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
