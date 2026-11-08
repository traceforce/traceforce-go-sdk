package traceforce

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.traceforce.co/api/v1"
)

type Client struct {
	httpClient   *http.Client
	baseURL      string
	apiKey       string // the token; read and written under mu (a client is shared by goroutines)
	extraHeaders map[string]string
	credentials  *Credentials // set by NewClientWithCredentials: log in again when the token expires
	mu           sync.RWMutex // guards apiKey
	loginMu      sync.Mutex   // one renewal at a time; guards the two fields below
	renewErr     error        // the last failed renewal ...
	renewErrAt   time.Time    // ... and when, so the calls waiting on it share it instead of each logging in again
}

// renewErrShared is how long the calls that waited on a failed renewal get its
// error instead of trying again (each try may take the 30s timeout).
const renewErrShared = 10 * time.Second

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
// HTTP 401), the cloud-connection calls log in again once and repeat the
// request; the other calls (datalakes, hosting environments, source apps,
// links) use the token as it is.
func NewClientWithCredentials(creds Credentials, url string, options *ClientOptions) (*Client, error) {
	c, err := NewClient("", url, options)
	if err != nil {
		return nil, err
	}
	c.credentials = &creds
	if err := c.renewToken(""); err != nil {
		return nil, err
	}
	return c, nil
}

// token is the current token ("" before the first login).
func (c *Client) token() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiKey
}

// renewToken logs in again with the client's credentials, unless another
// goroutine already replaced the token that failed (then its new token is
// used) or a renewal just failed (then its error is returned).
func (c *Client) renewToken(failed string) error {
	c.loginMu.Lock()
	defer c.loginMu.Unlock()
	if c.token() != failed {
		return nil
	}
	if c.renewErr != nil && time.Since(c.renewErrAt) < renewErrShared {
		return c.renewErr
	}
	if err := c.login(); err != nil {
		c.renewErr, c.renewErrAt = err, time.Now()
		return err
	}
	c.renewErr = nil
	return nil
}

// login gets a new token at POST /api-keys (sent without Authorization) and keeps it.
func (c *Client) login() error {
	var res struct {
		AccessToken string `json:"access_token"`
	}
	login := map[string]string{
		"client_id": c.credentials.ClientID, "client_secret": c.credentials.ClientSecret, "device_id": c.credentials.DeviceID,
	}
	if err := c.send(http.MethodPost, "/api-keys", "", login, &res); err != nil {
		return fmt.Errorf("login with the API client failed: %w", err)
	}
	if res.AccessToken == "" {
		return fmt.Errorf("login with the API client returned no token")
	}
	c.mu.Lock()
	c.apiKey = res.AccessToken
	c.mu.Unlock()
	return nil
}

// APIError is an answer with a status outside 2xx, from every call of this
// client. Its text is "HTTP <status>: <body>". Callers can read the status and
// the answer's "code" field, if any:
// errors.As(err, &apiErr) && apiErr.Code == CodeCollectorTooOld.
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

// isSuccess is the one success rule of this client: 2xx only.
func isSuccess(status int) bool { return status >= 200 && status < 300 }

// callAPI sends one request to the Traceforce API: body as JSON (if not nil),
// JSON answer into out (if not nil). An answer outside 2xx is an *APIError.
// With credentials, an expired token (401) is renewed once.
func (c *Client) callAPI(method, path string, body, out any) error {
	token := c.token()
	err := c.send(method, path, token, body, out)
	var apiErr *APIError
	if c.credentials != nil && errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized {
		if loginErr := c.renewToken(token); loginErr != nil {
			return loginErr
		}
		return c.send(method, path, c.token(), body, out)
	}
	return err
}

// send makes one request with this token ("" sends no Authorization).
func (c *Client) send(method, path, token string, body, out any) error {
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
	for k, v := range c.headers(token) {
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
	if err := validateResponse(resp); err != nil {
		return err
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
func (c *Client) buildHeaders() map[string]string { return c.headers(c.token()) }

func (c *Client) headers(token string) map[string]string {
	headers := map[string]string{}
	// No token yet while logging in (NewClientWithCredentials).
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	// Add extra headers
	for k, v := range c.extraHeaders {
		headers[k] = v
	}

	return headers
}
