package traceforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ConnectionsResponse struct {
	Data []ConnectionsModel `json:"data"`
}

type ConnectionsModel struct {
	ID                  string    `json:"id" omitempty:"true"`
	CreatedAt           time.Time `json:"created_at" omitempty:"true"`
	UpdatedAt           time.Time `json:"updated_at" omitempty:"true"`
	Name                string    `json:"name"`
	EnvironmentType     string    `json:"environment_type"`
	EnvironmentNativeId string    `json:"environment_native_id"`
	Status              string    `json:"status"`
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

	var createdConnection ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&createdConnection)
	if err != nil {
		return nil, err
	}

	return &createdConnection, nil
}

func (c *Client) GetConnections() (*ConnectionsResponse, error) {
	url := fmt.Sprintf("%s/connections", c.baseURL) + "/connections"
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

	var connections ConnectionsResponse
	err = json.NewDecoder(resp.Body).Decode(&connections)
	if err != nil {
		return nil, err
	}

	return &connections, nil
}

func (c *Client) GetConnection(id string) (*ConnectionsModel, error) {
	url := fmt.Sprintf("%s/connections/%s", c.baseURL, id)
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

	var connection ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&connection)
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

func (c *Client) UpdateConnection(id string, connection ConnectionsModel) (*ConnectionsModel, error) {
	url := fmt.Sprintf("%s/connections/%s", c.baseURL, id)
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

	var updatedConnection ConnectionsModel
	err = json.NewDecoder(resp.Body).Decode(&updatedConnection)
	if err != nil {
		return nil, err
	}

	return &updatedConnection, nil
}

func (c *Client) DeleteConnection(id string) error {
	url := fmt.Sprintf("%s/connections/%s", c.baseURL, id)
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

	return nil
}
