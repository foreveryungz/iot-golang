package sensor

import "strconv"

type Service interface {
	Create(input *CreateRequest) (*Sensor, error)
	FindAll() ([]Sensor, error)
	FindByID(id string) (*Sensor, error)
	Update(id string, input *UpdateRequest) (*Sensor, error)
	Delete(id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(input *CreateRequest) (*Sensor, error) {
	deviceExists, err := s.repo.DeviceExists(input.DeviceID)
	if err != nil {
		return nil, err
	}
	if !deviceExists {
		return nil, ErrInvalidDeviceID
	}

	item := &Sensor{
		DeviceID:   input.DeviceID,
		SensorCode: input.SensorCode,
		Name:       input.Name,
		Type:       input.Type,
		Unit:       input.Unit,
		Status:     input.Status,
	}

	if item.Status == "" {
		item.Status = "inactive"
	}
	if !isValidStatus(item.Status) {
		return nil, ErrInvalidStatus
	}

	err = s.repo.Create(item)
	return item, err
}

func (s *service) FindAll() ([]Sensor, error) {
	return s.repo.FindAll()
}

func (s *service) FindByID(id string) (*Sensor, error) {
	if _, err := parseID(id); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *service) Update(id string, input *UpdateRequest) (*Sensor, error) {
	if _, err := parseID(id); err != nil {
		return nil, err
	}

	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	deviceExists, err := s.repo.DeviceExists(input.DeviceID)
	if err != nil {
		return nil, err
	}
	if !deviceExists {
		return nil, ErrInvalidDeviceID
	}

	item.DeviceID = input.DeviceID
	item.SensorCode = input.SensorCode
	item.Name = input.Name
	item.Type = input.Type
	item.Unit = input.Unit
	item.Status = input.Status
	if !isValidStatus(item.Status) {
		return nil, ErrInvalidStatus
	}

	err = s.repo.Update(item)
	return item, err
}

func (s *service) Delete(id string) error {
	if _, err := parseID(id); err != nil {
		return err
	}

	item, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(item)
}

func parseID(id string) (uint64, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil || parsed == 0 {
		return 0, ErrInvalidInput
	}

	return parsed, nil
}
