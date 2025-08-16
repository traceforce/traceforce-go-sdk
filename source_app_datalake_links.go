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

// Request types
type CreateSourceAppDatalakeLinkRequest struct {
	SourceAppID string `json:"source_app_id"`
	DatalakeID  string `json:"datalake_id"`
}

// Response type
type SourceAppDatalakeLink struct {
	ID                   string    `json:"id"`
	SourceAppID          string    `json:"source_app_id"`
	DatalakeID           string    `json:"datalake_id"`
	HostingEnvironmentID string    `json:"hosting_environment_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (c *Client) CreateSourceAppDatalakeLink(req CreateSourceAppDatalakeLinkRequest) (*SourceAppDatalakeLink, error) {
	if req.SourceAppID == "" {
		return nil, fmt.Errorf("source app ID cannot be empty")
	}

	if req.DatalakeID == "" {
		return nil, fmt.Errorf("datalake ID cannot be empty")
	}

	// Validate UUID format for source app ID
	_, err := uuid.Parse(req.SourceAppID)
	if err != nil {
		return nil, fmt.Errorf("invalid source app ID UUID format: %v", err)
	}

	// Validate UUID format for datalake ID
	_, err = uuid.Parse(req.DatalakeID)
	if err != nil {
		return nil, fmt.Errorf("invalid datalake ID UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps-datalakes"
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

	var createdLink SourceAppDatalakeLink
	err = json.NewDecoder(resp.Body).Decode(&createdLink)
	if err != nil {
		return nil, err
	}

	return &createdLink, nil
}

func (c *Client) GetSourceAppDatalakeLinks() ([]SourceAppDatalakeLink, error) {
	url := c.baseURL + "/source-apps-datalakes"
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

	var links []SourceAppDatalakeLink
	err = json.NewDecoder(resp.Body).Decode(&links)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (c *Client) GetSourceAppDatalakeLinksBySourceApp(sourceAppID string) ([]SourceAppDatalakeLink, error) {
	if sourceAppID == "" {
		return nil, fmt.Errorf("source app ID cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(sourceAppID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps-datalakes?source_app_id=" + url.QueryEscape(sourceAppID)
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

	var links []SourceAppDatalakeLink
	err = json.NewDecoder(resp.Body).Decode(&links)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (c *Client) GetSourceAppDatalakeLinksByDatalake(datalakeID string) ([]SourceAppDatalakeLink, error) {
	if datalakeID == "" {
		return nil, fmt.Errorf("datalake ID cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(datalakeID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps-datalakes?datalake_id=" + url.QueryEscape(datalakeID)
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

	var links []SourceAppDatalakeLink
	err = json.NewDecoder(resp.Body).Decode(&links)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (c *Client) GetSourceAppDatalakeLink(id string) (*SourceAppDatalakeLink, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps-datalakes/" + id
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

	var link SourceAppDatalakeLink
	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (c *Client) DeleteSourceAppDatalakeLink(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	url := c.baseURL + "/source-apps-datalakes/" + id
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