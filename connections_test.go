package traceforce

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnections(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	connection := ConnectionsModel{
		Name:                "Test Connection",
		EnvironmentType:     "test",
		EnvironmentNativeId: "test",
		Status:              "disconnected",
	}

	createdConnection, err := client.CreateConnection(connection)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	t.Logf("Created connection: %+v", createdConnection)
	assert.NotNil(t, createdConnection)
	assert.Equal(t, connection.Name, createdConnection.Name)
	assert.Equal(t, connection.EnvironmentType, createdConnection.EnvironmentType)
	assert.Equal(t, connection.EnvironmentNativeId, createdConnection.EnvironmentNativeId)
	assert.Equal(t, connection.Status, createdConnection.Status)

	connections, err := client.GetConnections()
	if err != nil {
		t.Fatalf("Failed to get connections: %v", err)
	}

	t.Logf("Connections: %+v", connections)
	assert.NotNil(t, connections)
	assert.NotEmpty(t, connections)

	var testConnection ConnectionsModel
	for _, connection := range connections {
		t.Logf("Connection: %+v", connection)
		assert.NotNil(t, connection.ID)
		assert.NotEmpty(t, connection.Name)
		assert.NotEmpty(t, connection.EnvironmentType)
		assert.NotEmpty(t, connection.EnvironmentNativeId)
		assert.NotEmpty(t, connection.Status)

		if connection.Name == "Test Connection" {
			testConnection = connection
		}
	}

	t.Logf("Connection: %+v", testConnection)
	assert.NotNil(t, testConnection)

	testConnection.Status = "connected"
	updatedConnection, err := client.UpdateConnection(testConnection.ID, testConnection)
	if err != nil {
		t.Fatalf("Failed to update connection: %v", err)
	}

	t.Logf("Updated connection: %+v", updatedConnection)
	assert.NotNil(t, updatedConnection)
	assert.Equal(t, testConnection.Name, updatedConnection.Name)

	err = client.DeleteConnection(testConnection.ID)
	if err != nil {
		t.Fatalf("Failed to delete connection: %v", err)
	}

	t.Logf("Deleted connection: %+v", testConnection)
}

func TestDeleteConnection(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	id := "355defff-6999-4f34-addf-983e3f6052d1"

	err = client.DeleteConnection(id)
	if err != nil {
		t.Fatalf("Failed to delete connection: %v", err)
	}

	t.Logf("Deleted connection %v", id)
}
