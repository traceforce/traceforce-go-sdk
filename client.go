package traceforce

import (
	"bytes"
	"encoding/json"
	"errors"
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
	credentials  *Credentials // set by NewClientWithCredentials: log in again when the token expires
}

// Credentials of one of the org's API clients (Settings → API Clients).
type Credentials struct {
	ClientID     string
	ClientSecret string
	// DeviceID is who the token is for, e.g. CloudCollectorDeviceID("aws").
	DeviceID string
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

// NewClientWithCredentials logs in with one of the org's API clients and
// returns a client using that token. When the token expires (an answer with
// HTTP 401), the client logs in again once and repeats the request.
func NewClientWithCredentials(creds Credentials, url string, options *ClientOptions) (*Client, error) {
	c, err := NewClient("", url, options)
	if err != nil {
		return nil, err
	}
	c.credentials = &creds
	if err := c.login(); err != nil {
		return nil, err
	}
	return c, nil
}

// login gets a new token at POST /api-keys with the client's credentials.
func (c *Client) login() error {
	var res struct {
		AccessToken string `json:"access_token"`
	}
	login := map[string]string{
		"client_id": c.credentials.ClientID, "client_secret": c.credentials.ClientSecret, "device_id": c.credentials.DeviceID,
	}
	c.apiKey = ""
	if err := c.send(http.MethodPost, "/api-keys", login, &res); err != nil {
		return fmt.Errorf("login with the API client failed: %w", err)
	}
	if res.AccessToken == "" {
		return fmt.Errorf("login with the API client returned no token")
	}
	c.apiKey = res.AccessToken
	return nil
}

// APIError is an answer with a status outside 2xx. Its text is "HTTP <status>: <body>".
// Callers can read the status and the answer's "code" field, if any:
// errors.As(err, &apiErr) && apiErr.Code == "collector_too_old".
type APIError struct {
	StatusCode int
	Body       string
	Code       string // the "code" field of a JSON answer, e.g. "connection_not_found"; "" if none
}

func (e *APIError) Error() string { return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body) }

func newAPIError(status int, body []byte) *APIError {
	var answer struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(body, &answer) // not JSON: no code
	return &APIError{StatusCode: status, Body: string(body), Code: answer.Code}
}

// callAPI sends one request to the Traceforce API: body as JSON (if not nil),
// JSON answer into out (if not nil). An answer outside 2xx is an *APIError.
// With credentials, an expired token (401) is renewed once.
func (c *Client) callAPI(method, path string, body, out any) error {
	err := c.send(method, path, body, out)
	var apiErr *APIError
	if c.credentials != nil && errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized {
		if loginErr := c.login(); loginErr != nil {
			return loginErr
		}
		return c.send(method, path, body, out)
	}
	return err
}

func (c *Client) send(method, path string, body, out any) error {
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { // only 2xx is success (a 3xx or 1xx too is an error)
		body, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, body)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		// Read the unused answer to the end so the connection is reused; the call
		// already succeeded, so a failed read here is not an error.
		_, _ = io.Copy(io.Discard, resp.Body)
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
