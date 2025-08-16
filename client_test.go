package traceforce

import (
	"testing"
)

func TestNewClientWithExtraHeaders(t *testing.T) {
	// Test client creation without extra headers
	client1, err := NewClient("test-key", "https://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	if client1.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got '%s'", client1.apiKey)
	}
	
	if client1.baseURL != "https://example.com" {
		t.Errorf("Expected baseURL 'https://example.com', got '%s'", client1.baseURL)
	}
	
	if len(client1.extraHeaders) != 0 {
		t.Errorf("Expected no extra headers, got %d", len(client1.extraHeaders))
	}
	
	// Test client creation with extra headers
	options := &ClientOptions{
		ExtraHeaders: map[string]string{
			"x-vercel-protection-bypass": "test-token",
			"x-custom-header":            "test-value",
		},
	}
	
	client2, err := NewClient("test-key", "https://example.com", options)
	if err != nil {
		t.Fatalf("Failed to create client with extra headers: %v", err)
	}
	
	if len(client2.extraHeaders) != 2 {
		t.Errorf("Expected 2 extra headers, got %d", len(client2.extraHeaders))
	}
	
	if client2.extraHeaders["x-vercel-protection-bypass"] != "test-token" {
		t.Errorf("Expected bypass token 'test-token', got '%s'", client2.extraHeaders["x-vercel-protection-bypass"])
	}
	
	if client2.extraHeaders["x-custom-header"] != "test-value" {
		t.Errorf("Expected custom header 'test-value', got '%s'", client2.extraHeaders["x-custom-header"])
	}
}

func TestBuildHeaders(t *testing.T) {
	options := &ClientOptions{
		ExtraHeaders: map[string]string{
			"x-vercel-protection-bypass": "test-token",
			"x-custom-header":            "test-value",
		},
	}
	
	client, err := NewClient("test-api-key", "https://example.com", options)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	headers := client.buildHeaders()
	
	// Check authorization header
	if headers["Authorization"] != "Bearer test-api-key" {
		t.Errorf("Expected Authorization 'Bearer test-api-key', got '%s'", headers["Authorization"])
	}
	
	// Check extra headers are included
	if headers["x-vercel-protection-bypass"] != "test-token" {
		t.Errorf("Expected bypass token 'test-token', got '%s'", headers["x-vercel-protection-bypass"])
	}
	
	if headers["x-custom-header"] != "test-value" {
		t.Errorf("Expected custom header 'test-value', got '%s'", headers["x-custom-header"])
	}
	
	// Should have 3 headers total
	if len(headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(headers))
	}
}