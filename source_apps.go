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
	SourceAppStatusPending      SourceAppStatus = "Pending"
	SourceAppStatusDisconnected SourceAppStatus = "Disconnected"
	SourceAppStatusConnected    SourceAppStatus = "Connected"
)

type SourceAppType string

const (
	SourceAppTypeSalesforce SourceAppType = "Salesforce"
)

type SourceApp struct {
	ID                   string          `json:"id"`
	DatalakeID           string          `json:"datalake_id"`
	PodID                string          `json:"pod_id,omitempty"`
	HostingEnvironmentID string          `json:"hosting_environment_id"`
	Type                 SourceAppType   `json:"type"`
	Name                 string          `json:"name"`
	OrgID                string          `json:"org_id"`
	Status               SourceAppStatus `json:"status"`
	CreatedAt            time.Time       `json:"created_at,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at,omitempty"`
}

func (c *Client) CreateSourceApp(sourceApp SourceApp) (*SourceApp, error) {
	url := fmt.Sprintf("%s/source-apps", c.baseURL)
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	jsonBody, err := json.Marshal(sourceApp)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", headers["Authorization"])
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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
	url := fmt.Sprintf("%s/source-apps", c.baseURL)
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", headers["Authorization"])
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

func (c *Client) GetSourceAppsByDatalake(datalakeID string) ([]SourceApp, error) {
	if datalakeID == "" {
		return nil, fmt.Errorf("datalake ID cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(datalakeID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps?datalake_id=" + url.QueryEscape(datalakeID)
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", headers["Authorization"])
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
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", headers["Authorization"])
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
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", headers["Authorization"])
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

func (c *Client) UpdateSourceApp(id string, sourceApp SourceApp) (*SourceApp, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps/" + id
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	jsonBody, err := json.Marshal(sourceApp)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", headers["Authorization"])
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", headers["Authorization"])
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