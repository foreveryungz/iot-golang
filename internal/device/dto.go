package device

type CreateRequest struct {
	DeviceCode string `json:"device_code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Location   string `json:"location"`
	Status     string `json:"status"`
}

type UpdateRequest struct {
	DeviceCode string `json:"device_code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Location   string `json:"location"`
	Status     string `json:"status"`
}
