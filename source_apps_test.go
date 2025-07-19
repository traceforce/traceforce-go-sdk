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
	environment := HostingEnvironment{
		Name:          "test hosting environment for source app",
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: CloudProviderAWS,
		NativeID:      "123456789012",
		Status:        HostingEnvironmentStatusPending,
	}

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

	// Create a datalake for the source app
	datalake := Datalake{
		PodID:                "", // Optional - backend can assign
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 DatalakeTypeBigQuery,
		Name:                 "test datalake for source app",
		Status:               DatalakeStatusPending,
	}

	createdDatalake, err := client.CreateDatalake(datalake)
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
	sourceApp := SourceApp{
		DatalakeID:           createdDatalake.ID,
		PodID:                "", // Optional - backend can assign
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 SourceAppTypeSalesforce,
		Name:                 testSourceAppName,
		Status:               SourceAppStatusPending,
	}

	createdSourceApp, err := client.CreateSourceApp(sourceApp)
	if err != nil {
		t.Fatalf("Failed to create source app: %v", err)
	}

	t.Logf("Created source app: %+v", createdSourceApp)
	assert.NotNil(t, createdSourceApp)
	assert.Equal(t, sourceApp.Name, createdSourceApp.Name)
	assert.Equal(t, sourceApp.Type, createdSourceApp.Type)
	assert.Equal(t, sourceApp.DatalakeID, createdSourceApp.DatalakeID)
	assert.Equal(t, sourceApp.HostingEnvironmentID, createdSourceApp.HostingEnvironmentID)
	assert.Equal(t, sourceApp.Status, createdSourceApp.Status)

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
		assert.NotEmpty(t, sa.DatalakeID)
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

	sourceAppsByDatalake, err := client.GetSourceAppsByDatalake(createdDatalake.ID)
	if err != nil {
		t.Fatalf("Failed to get source apps by datalake: %v", err)
	}
	t.Logf("Source apps by datalake: %+v", sourceAppsByDatalake)
	assert.NotNil(t, sourceAppsByDatalake)
	assert.NotEmpty(t, sourceAppsByDatalake)

	found := false
	for _, sa := range sourceAppsByDatalake {
		if sa.ID == testSourceApp.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test source app should be found in datalake source apps")

	sourceAppsByEnvironment, err := client.GetSourceAppsByHostingEnvironment(createdEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to get source apps by hosting environment: %v", err)
	}
	t.Logf("Source apps by hosting environment: %+v", sourceAppsByEnvironment)
	assert.NotNil(t, sourceAppsByEnvironment)
	assert.NotEmpty(t, sourceAppsByEnvironment)

	found = false
	for _, sa := range sourceAppsByEnvironment {
		if sa.ID == testSourceApp.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test source app should be found in hosting environment source apps")

	testSourceApp.Status = SourceAppStatusConnected
	updatedSourceApp, err := client.UpdateSourceApp(testSourceApp.ID, testSourceApp)
	if err != nil {
		t.Fatalf("Failed to update source app: %v", err)
	}

	t.Logf("Updated source app: %+v", updatedSourceApp)
	assert.NotNil(t, updatedSourceApp)
	assert.Equal(t, testSourceApp.Name, updatedSourceApp.Name)
	assert.Equal(t, SourceAppStatusConnected, updatedSourceApp.Status)

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

	// Test GetSourceAppsByDatalake with empty ID
	_, err = client.GetSourceAppsByDatalake("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datalake ID cannot be empty")

	// Test GetSourceAppsByDatalake with invalid UUID
	_, err = client.GetSourceAppsByDatalake("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

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
	sourceApp := SourceApp{Name: "test"}
	_, err = client.UpdateSourceApp("", sourceApp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test UpdateSourceApp with invalid UUID
	_, err = client.UpdateSourceApp("invalid-uuid", sourceApp)
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