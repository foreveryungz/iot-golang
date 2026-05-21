package delivery

type MessageResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

type LogsResponseEnvelope struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []DeliveryLog `json:"data"`
}

type ProcessPendingResponse struct {
	Message string `json:"message"`
	Total   int    `json:"total"`
}
