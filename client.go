package traceforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// NewClientWithCredentials logs in with one of the org's API clients
// (Settings → API Clients) and returns a client using that token.
// deviceID is who the token is for, e.g. "cloud-collector-aws".
// The token lives 7 days; a Terraform run is much shorter.
func NewClientWithCredentials(clientID, clientSecret, deviceID, url string, options *ClientOptions) (*Client, error) {
	c, err := NewClient("", url, options)
	if err != nil {
		return nil, err
	}
	var res struct {
		AccessToken string `json:"access_token"`
	}
	login := map[string]string{"client_id": clientID, "client_secret": clientSecret, "device_id": deviceID}
	if err := c.callAPI(http.MethodPost, "/api-keys", login, &res); err != nil {
		return nil, fmt.Errorf("login with the API client failed: %w", err)
	}
	if res.AccessToken == "" {
		return nil, fmt.Errorf("login with the API client returned no token")
	}
	c.apiKey = res.AccessToken
	return c, nil
}

// APIError is an answer with status 400 or above (same text as validateResponse).
// Callers can read the status: errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string { return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body) }

// callAPI sends one request to the Traceforce API: body as JSON (if not nil), JSON answer into out (if not nil).
// A 400+ answer is an *APIError.
func (c *Client) callAPI(method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	for k, v := range c.buildHeaders() {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// buildHeaders creates a headers map with authorization and any extra headers
func (c *Client) buildHeaders() map[string]string {
	headers := map[string]string{}
	// No token yet while logging in (NewClientWithCredentials).
	if c.apiKey != "" {
		headers["Authorization"] = "Bearer " + c.apiKey
	}

	// Add extra headers
	for k, v := range c.extraHeaders {
		headers[k] = v
	}

	return headers
}
