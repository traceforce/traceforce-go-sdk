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
	postConnReq := &PostConnectionRequest{
		Infrastructure: &Infrastructure{
			Base: &BaseInfrastructure{
				DataplaneIdentityIdentifier:  "test-service-account@test-project.iam.gserviceaccount.com",
				WorkloadIdentityProviderName: "projects/123/locations/global/workloadIdentityPools/test-pool/providers/test-provider",
			},
		},
		TerraformModuleVersions: "{}",
	}
	err = client.PostConnection(createdEnvironment.ID, postConnReq)
	if err != nil {
		t.Fatalf("Failed to execute post-connection: %v", err)
	}
}

func TestPostConnectionValidation(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a valid request for ID validation tests
	validReq := &PostConnectionRequest{
		TerraformModuleVersions: "{}",
	}

	// Test PostConnection with empty ID
	err = client.PostConnection("", validReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test PostConnection with invalid UUID
	err = client.PostConnection("invalid-uuid", validReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test PostConnection with nil request
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	err = client.PostConnection(validUUID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")

	// Test PostConnection with empty terraform_module_versions
	emptyReq := &PostConnectionRequest{
		TerraformModuleVersions: "",
	}
	err = client.PostConnection(validUUID, emptyReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform_module_versions cannot be empty")

	// Test PostConnection with invalid JSON terraform_module_versions
	invalidJSONReq := &PostConnectionRequest{
		TerraformModuleVersions: "invalid-json",
	}
	err = client.PostConnection(validUUID, invalidJSONReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid terraform_module_versions JSON")

	// Test PostConnection with valid JSON terraform_module_versions (should parse correctly)
	// Note: This will fail with HTTP error since it's a real API call, but it validates JSON parsing
	validJSONReq := &PostConnectionRequest{
		TerraformModuleVersions: `{"base": {"version": "1.0.0"}, "connectors": {"bigquery": {"version": "2.0.0"}}}`,
	}
	err = client.PostConnection(validUUID, validJSONReq)
	// This should fail with HTTP error, not JSON parsing error
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "invalid terraform_module_versions JSON")
}