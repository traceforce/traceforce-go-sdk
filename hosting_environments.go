package traceforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type HostingEnvironmentStatus string

const (
	HostingEnvironmentStatusPending      HostingEnvironmentStatus = "pending"
	HostingEnvironmentStatusDisconnected HostingEnvironmentStatus = "disconnected"
	HostingEnvironmentStatusConnected    HostingEnvironmentStatus = "connected"
)

// PostConnectionRequest represents the infrastructure configuration for post-connection setup
type PostConnectionRequest struct {
	Infrastructure              *Infrastructure `json:"infrastructure"`
	TerraformURL                string          `json:"terraform_url"`
	TerraformModuleVersions     string          `json:"terraform_module_versions"`     // JSON string
	TerraformModuleVersionsHash string          `json:"terraform_module_versions_hash"`
	DeployedDatalakeIds         []string        `json:"deployed_datalake_ids,omitempty"`
	DeployedSourceAppIds        []string        `json:"deployed_source_app_ids,omitempty"`
}

// Infrastructure represents all connector-specific infrastructure outputs
type Infrastructure struct {
	Base       *BaseInfrastructure       `json:"base,omitempty"`
	BigQuery   *BigQueryInfrastructure   `json:"bigquery,omitempty"`
	Salesforce *SalesforceInfrastructure `json:"salesforce,omitempty"`
}

// BaseInfrastructure represents base infrastructure outputs
type BaseInfrastructure struct {
	DataplaneIdentityIdentifier  string `json:"dataplane_identity_identifier"`
	WorkloadIdentityProviderName string `json:"workload_identity_provider_name,omitempty"`
}

// BigQueryInfrastructure represents BigQuery datalake infrastructure outputs
type BigQueryInfrastructure struct {
	TraceforceSchema       string `json:"traceforce_schema"`
	EventsSubscriptionName string `json:"events_subscription_name"`
}

// SalesforceInfrastructure represents Salesforce source app infrastructure outputs
type SalesforceInfrastructure struct {
	ClientID     string `json:"salesforce_client_id"`
	Domain       string `json:"salesforce_domain"`
	ClientSecret string `json:"salesforce_client_secret"`
}

type HostingEnvironmentType string

const (
	HostingEnvironmentTypeCustomerManaged  HostingEnvironmentType = "customer_managed"
	HostingEnvironmentTypeTraceForceManaged HostingEnvironmentType = "traceforce_managed"
)

type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
	CloudProviderAzure CloudProvider = "azure"
)

// Request types
type CreateHostingEnvironmentRequest struct {
	Name          string                   `json:"name"`
	Type          HostingEnvironmentType   `json:"type"`
	CloudProvider CloudProvider            `json:"cloud_provider"`
	NativeID      string                   `json:"native_id"`
}

type UpdateHostingEnvironmentRequest struct {
	Name *string `json:"name,omitempty"`
}

// Response type
type HostingEnvironment struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Type          HostingEnvironmentType   `json:"type"`
	CloudProvider CloudProvider            `json:"cloud_provider"`
	NativeID      string                   `json:"native_id"`
	Status        HostingEnvironmentStatus `json:"status"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

func (c *Client) CreateHostingEnvironment(req CreateHostingEnvironmentRequest) (*HostingEnvironment, error) {
	url := c.baseURL + "/hosting-environments"
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

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

	var createdEnv HostingEnvironment
	err = json.NewDecoder(resp.Body).Decode(&createdEnv)
	if err != nil {
		return nil, err
	}

	return &createdEnv, nil
}

func (c *Client) GetHostingEnvironments() ([]HostingEnvironment, error) {
	url := c.baseURL + "/hosting-environments"
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

func (c *Client) UpdateHostingEnvironment(id string, req UpdateHostingEnvironmentRequest) (*HostingEnvironment, error) {
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

func (c *Client) PostConnection(id string, req *PostConnectionRequest) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	// Validate request is not nil
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate and parse JSON
	if req.TerraformModuleVersions == "" {
		return fmt.Errorf("terraform_module_versions cannot be empty")
	}
	
	var terraformModuleVersions interface{}
	if err := json.Unmarshal([]byte(req.TerraformModuleVersions), &terraformModuleVersions); err != nil {
		return fmt.Errorf("invalid terraform_module_versions JSON: %v", err)
	}

	url := c.baseURL + "/hosting-environments/" + id + "/post-connection"
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	// Create request payload with infrastructure configuration and terraform metadata
	payload := map[string]interface{}{
		"infrastructure":                req.Infrastructure,
		"terraform_url":                 req.TerraformURL,
		"terraform_module_versions":     terraformModuleVersions,
		"terraform_module_versions_hash": req.TerraformModuleVersionsHash,
		"deployed_datalake_ids":         req.DeployedDatalakeIds,
		"deployed_source_app_ids":       req.DeployedSourceAppIds,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal infrastructure configuration: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", headers["Authorization"])
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return err
	}

	return nil
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