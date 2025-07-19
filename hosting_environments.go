package traceforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type HostingEnvironmentStatus string

const (
	HostingEnvironmentStatusPending      HostingEnvironmentStatus = "Pending"
	HostingEnvironmentStatusDisconnected HostingEnvironmentStatus = "Disconnected"
	HostingEnvironmentStatusConnected    HostingEnvironmentStatus = "Connected"
)

type HostingEnvironmentType string

const (
	HostingEnvironmentTypeCustomerManaged  HostingEnvironmentType = "Customer Managed"
	HostingEnvironmentTypeTraceForceManaged HostingEnvironmentType = "TraceForce Managed"
)

type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "AWS"
	CloudProviderGCP   CloudProvider = "GCP"
	CloudProviderAzure CloudProvider = "Azure"
)

type HostingEnvironment struct {
	ID                       string                   `json:"id"`
	Name                     string                   `json:"name"`
	Type                     HostingEnvironmentType   `json:"type"`
	CloudProvider            CloudProvider            `json:"cloud_provider,omitempty"`
	NativeID                 string                   `json:"native_id"`
	Status                   HostingEnvironmentStatus `json:"status"`
	OrgID                    string                   `json:"org_id"`
	ControlPlaneAwsAccountId string                   `json:"control_plane_aws_account_id,omitempty"`
	ControlPlaneRoleName     string                   `json:"control_plane_role_name,omitempty"`
	CreatedAt                time.Time                `json:"created_at,omitempty"`
	UpdatedAt                time.Time                `json:"updated_at,omitempty"`
}

func (c *Client) CreateHostingEnvironment(env HostingEnvironment) (*HostingEnvironment, error) {
	url := fmt.Sprintf("%s/hosting-environments", c.baseURL)
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	jsonBody, err := json.Marshal(env)
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

	var createdEnv HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&createdEnv)
	if err != nil {
		return nil, err
	}

	return &createdEnv, nil
}

func (c *Client) GetHostingEnvironments() ([]HostingEnvironment, error) {
	url := fmt.Sprintf("%s/hosting-environments", c.baseURL)
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

	var environments []HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&environments)
	if err != nil {
		return nil, err
	}

	return environments, nil
}

func (c *Client) GetHostingEnvironmentByName(name string) (*HostingEnvironment, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	url := c.baseURL + "/hosting-environments?name=" + url.QueryEscape(name)
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

	var environment HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

func (c *Client) GetHostingEnvironment(id string) (*HostingEnvironment, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/hosting-environments/" + id
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

	var environment HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

func (c *Client) UpdateHostingEnvironment(id string, env HostingEnvironment) (*HostingEnvironment, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/hosting-environments/" + id
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	jsonBody, err := json.Marshal(env)
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

	var updatedEnv HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&updatedEnv)
	if err != nil {
		return nil, err
	}

	return &updatedEnv, nil
}

func (c *Client) DeleteHostingEnvironment(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/hosting-environments/" + id
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

func (c *Client) PostConnection(id string) (*HostingEnvironment, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/hosting-environments/" + id + "/post-connection"
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	req, err := http.NewRequest("POST", url, nil)
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

	var updatedEnv HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&updatedEnv)
	if err != nil {
		return nil, err
	}

	return &updatedEnv, nil
}

func validateResponse(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response body: %v", err)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}