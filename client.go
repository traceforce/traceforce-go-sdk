package traceforce

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://www.traceforce.co/api/v1"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

type ClientOptions struct {
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

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    url,
		apiKey:     key,
	}, nil
}
