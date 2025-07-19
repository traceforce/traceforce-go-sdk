package traceforce

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostingEnvironments(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	testEnvironmentName := "test hosting environment"
	environment := HostingEnvironment{
		Name:          testEnvironmentName,
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: CloudProviderAWS,
		NativeID:      "123456789012",
		Status:        HostingEnvironmentStatusPending,
	}

	createdEnvironment, err := client.CreateHostingEnvironment(environment)
	if err != nil {
		t.Fatalf("Failed to create hosting environment: %v", err)
	}

	t.Logf("Created hosting environment: %+v", createdEnvironment)
	assert.NotNil(t, createdEnvironment)
	assert.Equal(t, environment.Name, createdEnvironment.Name)
	assert.Equal(t, environment.Type, createdEnvironment.Type)
	assert.Equal(t, *environment.CloudProvider, *createdEnvironment.CloudProvider)
	assert.Equal(t, environment.NativeID, createdEnvironment.NativeID)
	assert.Equal(t, environment.Status, createdEnvironment.Status)

	environments, err := client.GetHostingEnvironments()
	if err != nil {
		t.Fatalf("Failed to get hosting environments: %v", err)
	}

	t.Logf("Hosting environments: %+v", environments)
	assert.NotNil(t, environments)
	assert.NotEmpty(t, environments)

	var testEnvironment HostingEnvironment
	for _, env := range environments {
		t.Logf("Hosting environment: %+v", env)
		assert.NotNil(t, env.ID)
		assert.NotEmpty(t, env.Name)
		assert.NotEmpty(t, env.Type)
		assert.NotEmpty(t, env.NativeID)
		assert.NotEmpty(t, env.Status)

		if env.Name == testEnvironmentName {
			testEnvironment = env
		}
	}

	t.Logf("Test hosting environment: %+v", testEnvironment)
	assert.NotNil(t, testEnvironment)

	environmentByName, err := client.GetHostingEnvironmentByName(testEnvironmentName)
	if err != nil {
		t.Fatalf("Failed to get hosting environment by name: %v", err)
	}
	t.Logf("Hosting environment by name: %+v", environmentByName)
	assert.NotNil(t, environmentByName)
	assert.Equal(t, testEnvironment.Name, environmentByName.Name)
	assert.Equal(t, testEnvironment.ID, environmentByName.ID)
	assert.Equal(t, testEnvironment.Type, environmentByName.Type)
	assert.Equal(t, testEnvironment.NativeID, environmentByName.NativeID)
	assert.Equal(t, testEnvironment.Status, environmentByName.Status)

	environmentByID, err := client.GetHostingEnvironment(testEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to get hosting environment by ID: %v", err)
	}
	t.Logf("Hosting environment by ID: %+v", environmentByID)
	assert.NotNil(t, environmentByID)
	assert.Equal(t, testEnvironment.ID, environmentByID.ID)
	assert.Equal(t, testEnvironment.Name, environmentByID.Name)

	testEnvironment.Status = HostingEnvironmentStatusConnected
	updatedEnvironment, err := client.UpdateHostingEnvironment(testEnvironment.ID, testEnvironment)
	if err != nil {
		t.Fatalf("Failed to update hosting environment: %v", err)
	}

	t.Logf("Updated hosting environment: %+v", updatedEnvironment)
	assert.NotNil(t, updatedEnvironment)
	assert.Equal(t, testEnvironment.Name, updatedEnvironment.Name)
	assert.Equal(t, HostingEnvironmentStatusConnected, updatedEnvironment.Status)

	err = client.DeleteHostingEnvironment(testEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to delete hosting environment: %v", err)
	}

	t.Logf("Deleted hosting environment: %+v", testEnvironment)
}

func TestHostingEnvironmentValidation(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetHostingEnvironmentByName with empty name
	_, err = client.GetHostingEnvironmentByName("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")

	// Test GetHostingEnvironment with empty ID
	_, err = client.GetHostingEnvironment("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test GetHostingEnvironment with invalid UUID
	_, err = client.GetHostingEnvironment("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test UpdateHostingEnvironment with empty ID
	env := HostingEnvironment{Name: "test"}
	_, err = client.UpdateHostingEnvironment("", env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test UpdateHostingEnvironment with invalid UUID
	_, err = client.UpdateHostingEnvironment("invalid-uuid", env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test DeleteHostingEnvironment with empty ID
	err = client.DeleteHostingEnvironment("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test DeleteHostingEnvironment with invalid UUID
	err = client.DeleteHostingEnvironment("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")
}

func TestPostConnection(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	testEnvironmentName := "test hosting environment for post connection"
	gcpProvider := CloudProviderGCP
	environment := HostingEnvironment{
		Name:          testEnvironmentName,
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: &gcpProvider,
		NativeID:      "test-project-123",
		Status:        HostingEnvironmentStatusPending,
	}

	// Create hosting environment
	createdEnvironment, err := client.CreateHostingEnvironment(environment)
	if err != nil {
		t.Fatalf("Failed to create hosting environment: %v", err)
	}
	defer func() {
		err := client.DeleteHostingEnvironment(createdEnvironment.ID)
		if err != nil {
			t.Logf("Failed to cleanup hosting environment: %v", err)
		}
	}()

	// Verify initial status is Pending
	assert.Equal(t, HostingEnvironmentStatusPending, createdEnvironment.Status)

	// Execute post-connection
	updatedEnvironment, err := client.PostConnection(createdEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to execute post-connection: %v", err)
	}

	t.Logf("Post-connection result: %+v", updatedEnvironment)
	assert.NotNil(t, updatedEnvironment)
	assert.Equal(t, createdEnvironment.ID, updatedEnvironment.ID)
	assert.Equal(t, HostingEnvironmentStatusConnected, updatedEnvironment.Status)
	assert.Equal(t, createdEnvironment.Name, updatedEnvironment.Name)
	assert.Equal(t, createdEnvironment.Type, updatedEnvironment.Type)
	assert.Equal(t, createdEnvironment.NativeID, updatedEnvironment.NativeID)
}

func TestPostConnectionValidation(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test PostConnection with empty ID
	_, err = client.PostConnection("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test PostConnection with invalid UUID
	_, err = client.PostConnection("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")
}