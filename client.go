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
	auth         *tokenSource
	extraHeaders map[string]string
}

// tokenSource holds the client's token and renews it (a client is shared by goroutines).
type tokenSource struct {
	login   func() (string, error) // nil for a client made with a static key: it never renews
	mu      sync.RWMutex           // guards the three fields below
	token   string
	gen     int        // how many renewals have finished
	err     error      // the result of the last renewal
	loginMu sync.Mutex // one renewal at a time
}

// Token is the current token ("" before the first login) and the number of
// the renewal it came from (pass it to Renew).
func (s *tokenSource) Token() (string, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token, s.gen
}

// Renew logs in again, unless a renewal finished after gen was read: then the
// call gets that renewal's result (its new token or its error). So the calls
// in flight during one renewal share it, and a call started after a failed
// renewal tries again.
func (s *tokenSource) Renew(gen int) error {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	s.mu.RLock()
	done, err := s.gen != gen, s.err
	s.mu.RUnlock()
	if done {
		return err
	}
	token, err := s.login()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gen++
	s.err = err
	if err == nil {
		s.token = token
	}
	return err
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
		auth:         &tokenSource{token: key},
		extraHeaders: extraHeaders,
	}, nil
}

// NewClientWithCredentials logs in with one of the org's API clients and
// returns a client using that token. When the token expires (an answer with
// HTTP 401), the cloud-connection calls log in again once and repeat the
// request; the other calls (datalakes, hosting environments, source apps,
// links) use the token as it is. An "Authorization" in ExtraHeaders is not
// sent: it would replace the token.
func NewClientWithCredentials(creds Credentials, url string, options *ClientOptions) (*Client, error) {
	c, err := NewClient("", url, options)
	if err != nil {
		return nil, err
	}
	c.auth.login = func() (string, error) { return c.login(creds) }
	if err := c.auth.Renew(0); err != nil {
		return nil, err
	}
	return c, nil
}

// login gets a new token at POST /api-keys (sent without Authorization).
func (c *Client) login(creds Credentials) (string, error) {
	var res struct {
		AccessToken string `json:"access_token"`
	}
	login := map[string]string{
		"client_id": creds.ClientID, "client_secret": creds.ClientSecret, "device_id": creds.DeviceID,
	}
	if err := c.send(http.MethodPost, "/api-keys", "", login, &res); err != nil {
		return "", fmt.Errorf("login with the API client failed: %w", err)
	}
	if res.AccessToken == "" {
		return "", fmt.Errorf("login with the API client returned no token")
	}
	return res.AccessToken, nil
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
	token, gen := c.auth.Token()
	err := c.send(method, path, token, body, out)
	var apiErr *APIError
	if c.auth.login != nil && errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized {
		if loginErr := c.auth.Renew(gen); loginErr != nil {
			return loginErr
		}
		token, _ = c.auth.Token()
		return c.send(method, path, token, body, out)
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
	for k, v := range c.headersWithToken(token) {
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

// buildHeaders creates a headers map with the current token and any extra
// headers (for the older calls, which do not renew the token).
func (c *Client) buildHeaders() map[string]string {
	token, _ := c.auth.Token()
	return c.headersWithToken(token)
}

// headersWithToken: token is "" for the login request and for a client made
// without a key; then no Authorization is set.
func (c *Client) headersWithToken(token string) map[string]string {
	headers := map[string]string{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	// Add extra headers. A client with credentials keeps its own Authorization
	// (an extra one would replace the token, and be sent with the login).
	for k, v := range c.extraHeaders {
		if c.auth.login != nil && http.CanonicalHeaderKey(k) == "Authorization" {
			continue
		}
		headers[k] = v
	}

	return headers
}
