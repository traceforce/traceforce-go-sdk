package traceforce

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceApps(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// First create a hosting environment for the datalake
	environmentReq := CreateHostingEnvironmentRequest{
		Name:          "test hosting environment for source app",
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: CloudProviderAWS,
		NativeID:      "123456789012",
	}

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

	// Create a datalake for the source app
	datalakeReq := CreateDatalakeRequest{
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 DatalakeTypeBigQuery,
		Name:                 "test datalake for source app",
	}

	createdDatalake, err := client.CreateDatalake(datalakeReq)
	if err != nil {
		t.Fatalf("Failed to create datalake: %v", err)
	}
	defer func() {
		err := client.DeleteDatalake(createdDatalake.ID)
		if err != nil {
			t.Logf("Failed to cleanup datalake: %v", err)
		}
	}()

	testSourceAppName := "test source app"
	sourceAppReq := CreateSourceAppRequest{
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 SourceAppTypeSalesforce,
		Name:                 testSourceAppName,
	}

	createdSourceApp, err := client.CreateSourceApp(sourceAppReq)
	if err != nil {
		t.Fatalf("Failed to create source app: %v", err)
	}

	t.Logf("Created source app: %+v", createdSourceApp)
	assert.NotNil(t, createdSourceApp)
	assert.Equal(t, sourceAppReq.Name, createdSourceApp.Name)
	assert.Equal(t, sourceAppReq.Type, createdSourceApp.Type)
	assert.Equal(t, sourceAppReq.HostingEnvironmentID, createdSourceApp.HostingEnvironmentID)
	assert.Equal(t, SourceAppStatusPending, createdSourceApp.Status)

	sourceApps, err := client.GetSourceApps()
	if err != nil {
		t.Fatalf("Failed to get source apps: %v", err)
	}

	t.Logf("Source apps: %+v", sourceApps)
	assert.NotNil(t, sourceApps)
	assert.NotEmpty(t, sourceApps)

	var testSourceApp SourceApp
	for _, sa := range sourceApps {
		t.Logf("Source app: %+v", sa)
		assert.NotNil(t, sa.ID)
		assert.NotEmpty(t, sa.Name)
		assert.NotEmpty(t, sa.Type)
		assert.NotEmpty(t, sa.HostingEnvironmentID)
		assert.NotEmpty(t, sa.Status)

		if sa.Name == testSourceAppName {
			testSourceApp = sa
		}
	}

	t.Logf("Test source app: %+v", testSourceApp)
	assert.NotNil(t, testSourceApp)

	sourceAppByID, err := client.GetSourceApp(testSourceApp.ID)
	if err != nil {
		t.Fatalf("Failed to get source app by ID: %v", err)
	}
	t.Logf("Source app by ID: %+v", sourceAppByID)
	assert.NotNil(t, sourceAppByID)
	assert.Equal(t, testSourceApp.ID, sourceAppByID.ID)
	assert.Equal(t, testSourceApp.Name, sourceAppByID.Name)

	sourceAppsByEnvironment, err := client.GetSourceAppsByHostingEnvironment(createdEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to get source apps by hosting environment: %v", err)
	}
	t.Logf("Source apps by hosting environment: %+v", sourceAppsByEnvironment)
	assert.NotNil(t, sourceAppsByEnvironment)
	assert.NotEmpty(t, sourceAppsByEnvironment)

	found := false
	for _, sa := range sourceAppsByEnvironment {
		if sa.ID == testSourceApp.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test source app should be found in hosting environment source apps")

	newName := testSourceApp.Name + " updated"
	updateReq := UpdateSourceAppRequest{
		Name: &newName,
	}
	updatedSourceApp, err := client.UpdateSourceApp(testSourceApp.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update source app: %v", err)
	}

	t.Logf("Updated source app: %+v", updatedSourceApp)
	assert.NotNil(t, updatedSourceApp)
	assert.Equal(t, newName, updatedSourceApp.Name)
	// Note: Status update is not supported via UpdateSourceAppRequest

	err = client.DeleteSourceApp(testSourceApp.ID)
	if err != nil {
		t.Fatalf("Failed to delete source app: %v", err)
	}

	t.Logf("Deleted source app: %+v", testSourceApp)
}

func TestSourceAppValidation(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}


	// Test GetSourceAppsByHostingEnvironment with empty ID
	_, err = client.GetSourceAppsByHostingEnvironment("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hosting environment ID cannot be empty")

	// Test GetSourceAppsByHostingEnvironment with invalid UUID
	_, err = client.GetSourceAppsByHostingEnvironment("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test GetSourceApp with empty ID
	_, err = client.GetSourceApp("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test GetSourceApp with invalid UUID
	_, err = client.GetSourceApp("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test UpdateSourceApp with empty ID
	testName := "test"
	updateReq := UpdateSourceAppRequest{Name: &testName}
	_, err = client.UpdateSourceApp("", updateReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test UpdateSourceApp with invalid UUID
	_, err = client.UpdateSourceApp("invalid-uuid", updateReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test DeleteSourceApp with empty ID
	err = client.DeleteSourceApp("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test DeleteSourceApp with invalid UUID
	err = client.DeleteSourceApp("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")
}