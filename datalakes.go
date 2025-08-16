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

type DatalakeStatus string

const (
	DatalakeStatusPending             DatalakeStatus = "pending"
	DatalakeStatusDeployed            DatalakeStatus = "deployed"
	DatalakeStatusReady               DatalakeStatus = "ready"
	DatalakeStatusFailed              DatalakeStatus = "failed"
)

type DatalakeType string

const (
	DatalakeTypeBigQuery DatalakeType = "bigquery"
)

// Request types
type CreateDatalakeRequest struct {
	HostingEnvironmentID string       `json:"hosting_environment_id"`
	Type                 DatalakeType `json:"type"`
	Name                 string       `json:"name"`
	EnvironmentNativeID  string       `json:"environment_native_id"`
	Region               string       `json:"region"`
}

type UpdateDatalakeRequest struct {
	Name *string `json:"name,omitempty"`
}

// Response type
type Datalake struct {
	ID                   string         `json:"id"`
	HostingEnvironmentID string         `json:"hosting_environment_id"`
	Type                 DatalakeType   `json:"type"`
	Name                 string         `json:"name"`
	Status               DatalakeStatus `json:"status"`
	EnvironmentNativeID  string         `json:"environment_native_id"`
	Region               string         `json:"region"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

func (c *Client) CreateDatalake(req CreateDatalakeRequest) (*Datalake, error) {
	url := c.baseURL + "/datalakes"
	headers := c.buildHeaders()

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", headers["Authorization"])
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var createdDatalake Datalake
	err = json.NewDecoder(resp.Body).Decode(&createdDatalake)
	if err != nil {
		return nil, err
	}

	return &createdDatalake, nil
}

func (c *Client) GetDatalakes() ([]Datalake, error) {
	url := c.baseURL + "/datalakes"
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

	var datalakes []Datalake
	err = json.NewDecoder(resp.Body).Decode(&datalakes)
	if err != nil {
		return nil, err
	}

	return datalakes, nil
}

func (c *Client) GetDatalakesByHostingEnvironment(hostingEnvironmentID string) ([]Datalake, error) {
	if hostingEnvironmentID == "" {
		return nil, fmt.Errorf("hosting environment ID cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(hostingEnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/datalakes?hosting_environment_id=" + url.QueryEscape(hostingEnvironmentID)
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

	var datalakes []Datalake
	err = json.NewDecoder(resp.Body).Decode(&datalakes)
	if err != nil {
		return nil, err
	}

	return datalakes, nil
}

func (c *Client) GetDatalake(id string) (*Datalake, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/datalakes/" + id
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

	var datalake Datalake
	err = json.NewDecoder(resp.Body).Decode(&datalake)
	if err != nil {
		return nil, err
	}

	return &datalake, nil
}

func (c *Client) UpdateDatalake(id string, req UpdateDatalakeRequest) (*Datalake, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/datalakes/" + id
	headers := c.buildHeaders()

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", headers["Authorization"])
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	var updatedDatalake Datalake
	err = json.NewDecoder(resp.Body).Decode(&updatedDatalake)
	if err != nil {
		return nil, err
	}

	return &updatedDatalake, nil
}

func (c *Client) DeleteDatalake(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/datalakes/" + id
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