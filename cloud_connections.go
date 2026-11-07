package traceforce

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Cloud connections: one per customer cloud account (an AWS account or a
// Microsoft 365 tenant) whose collector runs in the customer's cloud.
//
// These calls need a collector token: a client from NewClientWithCredentials
// with DeviceID CloudCollectorDeviceID(provider). A token may only reach its
// own cloud's connections (another cloud's answers HTTP 403).

// CloudCollectorDeviceID is the device id of a cloud collector's token, e.g. "cloud-collector-aws".
func CloudCollectorDeviceID(provider string) string { return "cloud-collector-" + provider }

// CloudConnectionStatus is where a connection is.
type CloudConnectionStatus int

const (
	CloudConnectionStatusPending   CloudConnectionStatus = 1 // registered, no upload yet
	CloudConnectionStatusConnected CloudConnectionStatus = 2 // the last upload was stored
	CloudConnectionStatusFailed    CloudConnectionStatus = 3 // the last upload carried an error
)

// CreateCloudConnectionRequest registers a cloud account.
type CreateCloudConnectionRequest struct {
	Provider      string `json:"provider"`               // "aws" or "m365"
	Account       string `json:"account"`                // AWS account id or tenant id
	Name          string `json:"name,omitempty"`         // display name until the first upload
	ModuleVersion string `json:"moduleVersion"`          // collector bundle release, e.g. "1.2.0"
	ModuleCommit  string `json:"moduleCommit,omitempty"` // git commit of that release
}

// CloudConnectionDeployment is the collector deployed for a connection.
type CloudConnectionDeployment struct {
	CollectorID string   `json:"collectorId"`       // Lambda ARN / Function App id
	IdentityID  string   `json:"identityId"`        // IAM role ARN / managed identity object id
	Regions     []string `json:"regions,omitempty"` // AWS only; omitted keeps the stored regions
}

// CloudConnection is one connection as the API answers it (same names as the requests).
type CloudConnection struct {
	ID            string                `json:"id"`
	Provider      string                `json:"provider"` // "aws" or "m365"
	Account       string                `json:"account"`
	Name          *string               `json:"name"`
	Status        CloudConnectionStatus `json:"status"`
	ModuleVersion *string               `json:"moduleVersion"`
	ModuleCommit  *string               `json:"moduleCommit"`
	CollectorID   *string               `json:"collectorId"`
	IdentityID    *string               `json:"identityId"`
	Regions       []string              `json:"regions"`
	DeployedAt    *time.Time            `json:"deployedAt"` // nil = no collector recorded
	// Created is set by CreateCloudConnection only: false when the account was registered before.
	Created bool `json:"created,omitempty"`
}

// APIError codes of the cloud-connection routes.
const (
	// CodeCollectorTooOld: the bundle is older than the cloud's minimum version (HTTP 426); nothing was created.
	CodeCollectorTooOld = "collector_too_old"
	// CodeConnectionNotFound: the org has no such cloud connection (HTTP 404).
	CodeConnectionNotFound = "connection_not_found"
)

// CreateCloudConnection registers the account, or answers the same connection
// when this org already has it (Created false).
func (c *Client) CreateCloudConnection(req CreateCloudConnectionRequest) (*CloudConnection, error) {
	var conn CloudConnection
	if err := c.callAPI(http.MethodPost, "/cloud-connections", req, &conn); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(conn.ID); err != nil {
		return nil, fmt.Errorf("register cloud connection: the API answered no valid id (%q)", conn.ID)
	}
	return &conn, nil
}

// GetCloudConnection returns the connection, or nil, nil when it does not
// exist (or was disconnected): see IsCloudConnectionNotFound.
func (c *Client) GetCloudConnection(id string) (*CloudConnection, error) {
	return c.connectionCall(http.MethodGet, id, "", nil)
}

// PutCloudConnectionDeployment records the collector deployed for the
// connection (sets DeployedAt) and returns the connection.
func (c *Client) PutCloudConnectionDeployment(id string, d CloudConnectionDeployment) (*CloudConnection, error) {
	conn, err := c.connectionCall(http.MethodPut, id, "/deployment", d)
	if err == nil && conn == nil {
		return nil, fmt.Errorf("record deployment: cloud connection %s not found", id)
	}
	return conn, err
}

// DeleteCloudConnectionDeployment clears the recorded collector (DeployedAt,
// CollectorID, IdentityID) and returns the connection, also a disconnected
// one; nil, nil when the org has no such connection.
func (c *Client) DeleteCloudConnectionDeployment(id string) (*CloudConnection, error) {
	return c.connectionCall(http.MethodDelete, id, "/deployment", nil)
}

// DeleteCloudConnection disconnects the connection. Not found counts as done.
func (c *Client) DeleteCloudConnection(id string) error {
	path, err := connectionPath(id)
	if err != nil {
		return err
	}
	if err := c.callAPI(http.MethodDelete, path, nil, nil); err != nil && !IsCloudConnectionNotFound(err) {
		return err
	}
	return nil
}

// connectionCall sends one request about a connection and reads the
// connection answered; nil, nil when the connection is not found.
func (c *Client) connectionCall(method, id, suffix string, body any) (*CloudConnection, error) {
	path, err := connectionPath(id)
	if err != nil {
		return nil, err
	}
	var conn CloudConnection
	if err := c.callAPI(method, path+suffix, body, &conn); err != nil {
		if IsCloudConnectionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &conn, nil
}

// connectionPath checks the id before any request is sent and writes it in
// the canonical form (any form uuid.Parse accepts, e.g. "{...}", is sent as
// lowercase 8-4-4-4-12).
func connectionPath(id string) (string, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid cloud connection id %q: %v", id, err)
	}
	return "/cloud-connections/" + u.String(), nil
}

// IsCloudConnectionNotFound reports whether err says the cloud connection does
// not exist (or was disconnected). Only the API's own answer counts: HTTP 404
// with code CodeConnectionNotFound, which a Traceforce backend with the
// cloud-connection routes sends. Any other 404 (a wrong base URL, an older
// backend) is a real error.
func IsCloudConnectionNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound && apiErr.Code == CodeConnectionNotFound
}
