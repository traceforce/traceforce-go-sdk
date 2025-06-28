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

// NewClient creates a new Supabase client.
// url is the Supabase URL.
// key is the Supabase API key.
// options is the Supabase client options.
func NewClient(url, key string, options *ClientOptions) (*Client, error) {
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
