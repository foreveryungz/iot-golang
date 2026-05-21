package monitoring

type MessageResponse struct {
	Message string `json:"message"`
}

type DetailResponseEnvelope struct {
	Data MonitoringData `json:"data"`
}

type ListResponseEnvelope struct {
	Data []MonitoringData `json:"data"`
}
