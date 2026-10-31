package clients

import (
	"net/http"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() (*HTTPClient, error) {
	return &HTTPClient{client: &http.Client{}}, nil
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
