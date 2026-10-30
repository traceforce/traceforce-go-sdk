package traceforce

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Cloud connections: one per customer cloud account (AWS account or
// Microsoft 365 tenant). The Terraform provider registers one at the start of
// apply, records the collector after it is created, and disconnects it on
// destroy. Use a client from NewClientWithCredentials with device id
// "cloud-collector-<provider>".

// CreateCloudConnectionRequest is sent at the start of terraform apply.
type CreateCloudConnectionRequest struct {
	Provider      string `json:"provider"`               // "aws" or "m365"
	Account       string `json:"account"`                // AWS account id or tenant id
	Name          string `json:"name,omitempty"`         // shown until the first upload
	ModuleVersion string `json:"moduleVersion"`          // bundle release, e.g. "1.2.0"
	ModuleCommit  string `json:"moduleCommit,omitempty"` // git commit of that release
}

// CloudConnectionDeployment is what Terraform created in the customer's cloud.
type CloudConnectionDeployment struct {
	CollectorID string   `json:"collectorId"`       // Lambda ARN / Function App id
	IdentityID  string   `json:"identityId"`        // IAM role ARN / managed identity object id
	Regions     []string `json:"regions,omitempty"` // AWS only
}

// CloudConnection is one connection as Traceforce stores it.
type CloudConnection struct {
	ID            string     `json:"id"`
	AccountRef    string     `json:"account_ref"`
	Status        int        `json:"status"` // 1 pending, 2 connected, 3 last upload failed
	ModuleVersion *string    `json:"module_version"`
	ModuleCommit  *string    `json:"module_commit"`
	CollectorID   *string    `json:"collector_id"`
	IdentityID    *string    `json:"identity_id"`
	DeployedAt    *time.Time `json:"deployed_at"` // nil = collector not recorded yet
}

// CreateCloudConnection registers the account (same id when it already exists).
// An *APIError with status 426 means the bundle is older than the minimum version.
func (c *Client) CreateCloudConnection(req CreateCloudConnectionRequest) (string, error) {
	var res struct {
		ID string `json:"id"`
	}
	if err := c.callAPI(http.MethodPost, "/cloud-connections", req, &res); err != nil {
		return "", err
	}
	return res.ID, nil
}

// GetCloudConnection returns the connection, or nil when it is gone (deleted or disconnected).
func (c *Client) GetCloudConnection(id string) (*CloudConnection, error) {
	path, err := connectionPath(id)
	if err != nil {
		return nil, err
	}
	var conn CloudConnection
	if err := c.callAPI(http.MethodGet, path, nil, &conn); err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &conn, nil
}

// PutCloudConnectionDeployment records the collector Terraform created.
func (c *Client) PutCloudConnectionDeployment(id string, d CloudConnectionDeployment) error {
	path, err := connectionPath(id)
	if err != nil {
		return err
	}
	return c.callAPI(http.MethodPut, path+"/deployment", d, nil)
}

// DeleteCloudConnection disconnects the connection. Already gone counts as done.
func (c *Client) DeleteCloudConnection(id string) error {
	path, err := connectionPath(id)
	if err != nil {
		return err
	}
	if err := c.callAPI(http.MethodDelete, path, nil, nil); err != nil && !isNotFound(err) {
		return err
	}
	return nil
}

// connectionPath checks the id before any request is sent.
func connectionPath(id string) (string, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("invalid cloud connection id %q: %v", id, err)
	}
	return "/cloud-connections/" + id, nil
}

func isNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
