package httpclient

import (
	"bytes"
	"net/http"
	"time"
)

type Client interface {
	PostJSON(
		url string,
		payload []byte,
		headers map[string]string,
	) (*http.Response, error)
}

type HTTPClient struct {
	client *http.Client
}

func New() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *HTTPClient) PostJSON(
	url string,
	payload []byte,
	headers map[string]string,
) (*http.Response, error) {

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(payload),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.client.Do(req)
}
