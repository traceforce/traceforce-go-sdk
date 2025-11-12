package traceforce

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.traceforce.co/api/v1"
)

type Client struct {
	httpClient   *http.Client
	baseURL      string
	apiKey       string
	extraHeaders map[string]string
}

type ClientOptions struct {
	// ExtraHeaders allows adding additional headers to all API requests.
	ExtraHeaders map[string]string `json:"extra_headers,omitempty"`
}

// NewClient creates a new Traceforce client.
// key is the Traceforce API key.
// url is the Traceforce URL.
// options is the Traceforce client options.
func NewClient(key, url string, options *ClientOptions) (*Client, error) {
	if url == "" {
		url = defaultBaseURL
	}

	if options == nil {
		options = &ClientOptions{}
	}

	extraHeaders := make(map[string]string)

	// Copy user-provided headers
	for k, v := range options.ExtraHeaders {
		extraHeaders[k] = v
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		httpClient:   httpClient,
		baseURL:      url,
		apiKey:       key,
		extraHeaders: extraHeaders,
	}, nil
}

// buildHeaders creates a headers map with authorization and any extra headers
func (c *Client) buildHeaders() map[string]string {
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	// Add extra headers
	for k, v := range c.extraHeaders {
		headers[k] = v
	}

	return headers
}
