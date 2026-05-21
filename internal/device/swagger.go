package device

type MessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DetailResponseEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Device `json:"data"`
}

type ListResponseEnvelope struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []Device `json:"data"`
}
