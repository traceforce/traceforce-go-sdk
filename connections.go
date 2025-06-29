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

type ConnectionsModel struct {
	ID                  string    `json:"id" omitempty:"true"`
	OrgID               string    `json:"org_id" omitempty:"true"`
	Name                string    `json:"name"`
	EnvironmentType     string    `json:"environment_type"`
	EnvironmentNativeId string    `json:"environment_native_id"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at" omitempty:"true"`
	UpdatedAt           time.Time `json:"updated_at" omitempty:"true"`
}

func (c *Client) CreateConnection(connection ConnectionsModel) (*ConnectionsModel, error) {
	url := fmt.Sprintf("%s/connections", c.baseURL)
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	jsonBody, err := json.Marshal(connection)
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

	var createdConnection ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&createdConnection)
	if err != nil {
		return nil, err
	}

	return &createdConnection, nil
}

func (c *Client) GetConnections() ([]ConnectionsModel, error) {
	url := fmt.Sprintf("%s/connections", c.baseURL)
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

	var connections []ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&connections)
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func (c *Client) GetConnectionByName(name string) (*ConnectionsModel, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	url := c.baseURL + "/connections?name=" + url.QueryEscape(name)
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

	var connection ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&connection)
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

func (c *Client) UpdateConnection(id string, connection ConnectionsModel) (*ConnectionsModel, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/connections/" + id
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	jsonBody, err := json.Marshal(connection)
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

	var updatedConnection ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&updatedConnection)
	if err != nil {
		return nil, err
	}

	return &updatedConnection, nil
}

func (c *Client) DeleteConnection(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/connections/" + id
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
