package monitoring

import "errors"

type Service interface {
	Create(input *CreateRequest) (*MonitoringData, error)
	FindAll() ([]MonitoringData, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

var (
	ErrDeviceNotFound       = errors.New("device not found")
	ErrSensorNotFound       = errors.New("sensor not found")
	ErrSensorDeviceMismatch = errors.New("sensor does not belong to the given device")
)

func (s *service) Create(input *CreateRequest) (*MonitoringData, error) {
	deviceExists, err := s.repo.DeviceExists(input.DeviceID)
	if err != nil {
		return nil, err
	}
	if !deviceExists {
		return nil, ErrDeviceNotFound
	}

	sensorRecord, err := s.repo.FindSensorByID(input.SensorID)
	if err != nil {
		return nil, err
	}
	if sensorRecord == nil {
		return nil, ErrSensorNotFound
	}
	if sensorRecord.DeviceID != input.DeviceID {
		return nil, ErrSensorDeviceMismatch
	}

	data := &MonitoringData{
		DeviceID:    input.DeviceID,
		SensorID:    input.SensorID,
		Value:       input.Value,
		Unit:        input.Unit,
		Status:      input.Status,
		ScheduledAt: input.ScheduledAt,
	}

	if data.Status == "" {
		data.Status = "pending"
	}

	if err := s.repo.Create(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *service) FindAll() ([]MonitoringData, error) {
	return s.repo.FindAll()
}
