package sensor

type MessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DetailResponseEnvelope struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    DetailResponse `json:"data"`
}

type ListResponseEnvelope struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []DetailResponse `json:"data"`
}
