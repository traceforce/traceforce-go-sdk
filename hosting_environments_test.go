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
	environmentReq := CreateHostingEnvironmentRequest{
		Name:          testEnvironmentName,
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: CloudProviderAWS,
		NativeID:      "123456789012",
	}

	createdEnvironment, err := client.CreateHostingEnvironment(environmentReq)
	if err != nil {
		t.Fatalf("Failed to create hosting environment: %v", err)
	}

	t.Logf("Created hosting environment: %+v", createdEnvironment)
	assert.NotNil(t, createdEnvironment)
	assert.Equal(t, environmentReq.Name, createdEnvironment.Name)
	assert.Equal(t, environmentReq.Type, createdEnvironment.Type)
	assert.Equal(t, environmentReq.CloudProvider, createdEnvironment.CloudProvider)
	assert.Equal(t, environmentReq.NativeID, createdEnvironment.NativeID)
	assert.Equal(t, HostingEnvironmentStatusPending, createdEnvironment.Status)

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

	environmentByID, err := client.GetHostingEnvironment(testEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to get hosting environment by ID: %v", err)
	}
	t.Logf("Hosting environment by ID: %+v", environmentByID)
	assert.NotNil(t, environmentByID)
	assert.Equal(t, testEnvironment.ID, environmentByID.ID)
	assert.Equal(t, testEnvironment.Name, environmentByID.Name)

	newName := testEnvironment.Name + " updated"
	updateReq := UpdateHostingEnvironmentRequest{
		Name: &newName,
	}
	updatedEnvironment, err := client.UpdateHostingEnvironment(testEnvironment.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update hosting environment: %v", err)
	}

	t.Logf("Updated hosting environment: %+v", updatedEnvironment)
	assert.NotNil(t, updatedEnvironment)
	assert.Equal(t, newName, updatedEnvironment.Name)
	// Note: Status update is not supported via UpdateHostingEnvironmentRequest

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

	// Test GetHostingEnvironment with empty ID
	_, err = client.GetHostingEnvironment("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test GetHostingEnvironment with invalid UUID
	_, err = client.GetHostingEnvironment("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test UpdateHostingEnvironment with empty ID
	testName := "test"
	updateReq := UpdateHostingEnvironmentRequest{Name: &testName}
	_, err = client.UpdateHostingEnvironment("", updateReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test UpdateHostingEnvironment with invalid UUID
	_, err = client.UpdateHostingEnvironment("invalid-uuid", updateReq)
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
	environmentReq := CreateHostingEnvironmentRequest{
		Name:          testEnvironmentName,
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: CloudProviderGCP,
		NativeID:      "test-project-123",
	}

	// Create hosting environment
	createdEnvironment, err := client.CreateHostingEnvironment(environmentReq)
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
	assert.Equal(t, environmentReq.Name, updatedEnvironment.Name)
	assert.Equal(t, environmentReq.Type, updatedEnvironment.Type)
	assert.Equal(t, environmentReq.NativeID, updatedEnvironment.NativeID)
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