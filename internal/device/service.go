package device

import "strconv"

type Service interface {
	Create(input *CreateRequest) (*Device, error)
	FindAll() ([]Device, error)
	FindByID(id string) (*Device, error)
	Update(id string, input *UpdateRequest) (*Device, error)
	Delete(id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(input *CreateRequest) (*Device, error) {
	device := &Device{
		DeviceCode: input.DeviceCode,
		Name:       input.Name,
		Location:   input.Location,
		Status:     input.Status,
	}

	if device.Status == "" {
		device.Status = "inactive"
	}
	if !isValidStatus(device.Status) {
		return nil, ErrInvalidStatus
	}

	err := s.repo.Create(device)
	return device, err
}

func (s *service) FindAll() ([]Device, error) {
	return s.repo.FindAll()
}

func (s *service) FindByID(id string) (*Device, error) {
	if _, err := parseID(id); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *service) Update(id string, input *UpdateRequest) (*Device, error) {
	if _, err := parseID(id); err != nil {
		return nil, err
	}

	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	item.DeviceCode = input.DeviceCode
	item.Name = input.Name
	item.Location = input.Location
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
