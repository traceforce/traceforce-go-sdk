package traceforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type SourceAppStatus string

const (
	SourceAppStatusPending      SourceAppStatus = "pending"
	SourceAppStatusDeployed     SourceAppStatus = "deployed"
	SourceAppStatusDisconnected SourceAppStatus = "disconnected"
	SourceAppStatusConnected    SourceAppStatus = "connected"
)

type SourceAppType string

const (
	SourceAppTypeSalesforce SourceAppType = "salesforce"
)

// Request types
type CreateSourceAppRequest struct {
	HostingEnvironmentID string        `json:"hosting_environment_id"`
	Type                 SourceAppType `json:"type"`
	Name                 string        `json:"name"`
}

type UpdateSourceAppRequest struct {
	Name *string `json:"name,omitempty"`
}

// Response type
type SourceApp struct {
	ID                   string          `json:"id"`
	HostingEnvironmentID string          `json:"hosting_environment_id"`
	Type                 SourceAppType   `json:"type"`
	Name                 string          `json:"name"`
	Status               SourceAppStatus `json:"status"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (c *Client) CreateSourceApp(req CreateSourceAppRequest) (*SourceApp, error) {
	url := c.baseURL + "/source-apps"
	headers := c.buildHeaders()

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	for k, v := range headers { httpReq.Header.Set(k, v) }
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var createdSourceApp SourceApp
	err = json.NewDecoder(resp.Body).Decode(&createdSourceApp)
	if err != nil {
		return nil, err
	}

	return &createdSourceApp, nil
}

func (c *Client) GetSourceApps() ([]SourceApp, error) {
	url := c.baseURL + "/source-apps"
	headers := c.buildHeaders()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers { req.Header.Set(k, v) }
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var sourceApps []SourceApp
	err = json.NewDecoder(resp.Body).Decode(&sourceApps)
	if err != nil {
		return nil, err
	}

	return sourceApps, nil
}


func (c *Client) GetSourceAppsByHostingEnvironment(hostingEnvironmentID string) ([]SourceApp, error) {
	if hostingEnvironmentID == "" {
		return nil, fmt.Errorf("hosting environment ID cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(hostingEnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps?hosting_environment_id=" + url.QueryEscape(hostingEnvironmentID)
	headers := c.buildHeaders()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers { req.Header.Set(k, v) }
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var sourceApps []SourceApp
	err = json.NewDecoder(resp.Body).Decode(&sourceApps)
	if err != nil {
		return nil, err
	}

	return sourceApps, nil
}

func (c *Client) GetSourceApp(id string) (*SourceApp, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps/" + id
	headers := c.buildHeaders()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers { req.Header.Set(k, v) }
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var sourceApp SourceApp
	err = json.NewDecoder(resp.Body).Decode(&sourceApp)
	if err != nil {
		return nil, err
	}

	return &sourceApp, nil
}

func (c *Client) UpdateSourceApp(id string, req UpdateSourceAppRequest) (*SourceApp, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps/" + id
	headers := c.buildHeaders()

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	for k, v := range headers { httpReq.Header.Set(k, v) }
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var updatedSourceApp SourceApp
	err = json.NewDecoder(resp.Body).Decode(&updatedSourceApp)
	if err != nil {
		return nil, err
	}

	return &updatedSourceApp, nil
}

func (c *Client) DeleteSourceApp(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps/" + id
	headers := c.buildHeaders()

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers { req.Header.Set(k, v) }
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return err
	}

	return nil
}