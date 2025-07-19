package traceforce

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatalakes(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// First create a hosting environment for the datalake
	awsProvider := CloudProviderAWS
	environment := HostingEnvironment{
		Name:          "test hosting environment for datalake",
		Type:          HostingEnvironmentTypeCustomerManaged,
		CloudProvider: &awsProvider,
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

	testDatalakeName := "test datalake"
	datalake := Datalake{
		PodID:                "", // Optional - backend can assign
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 DatalakeTypeBigQuery,
		Name:                 testDatalakeName,
		Status:               DatalakeStatusPending,
	}

	createdDatalake, err := client.CreateDatalake(datalake)
	if err != nil {
		t.Fatalf("Failed to create datalake: %v", err)
	}

	t.Logf("Created datalake: %+v", createdDatalake)
	assert.NotNil(t, createdDatalake)
	assert.Equal(t, datalake.Name, createdDatalake.Name)
	assert.Equal(t, datalake.Type, createdDatalake.Type)
	assert.Equal(t, datalake.HostingEnvironmentID, createdDatalake.HostingEnvironmentID)
	assert.Equal(t, datalake.Status, createdDatalake.Status)

	datalakes, err := client.GetDatalakes()
	if err != nil {
		t.Fatalf("Failed to get datalakes: %v", err)
	}

	t.Logf("Datalakes: %+v", datalakes)
	assert.NotNil(t, datalakes)
	assert.NotEmpty(t, datalakes)

	var testDatalake Datalake
	for _, dl := range datalakes {
		t.Logf("Datalake: %+v", dl)
		assert.NotNil(t, dl.ID)
		assert.NotEmpty(t, dl.Name)
		assert.NotEmpty(t, dl.Type)
		assert.NotEmpty(t, dl.HostingEnvironmentID)
		assert.NotEmpty(t, dl.Status)

		if dl.Name == testDatalakeName {
			testDatalake = dl
		}
	}

	t.Logf("Test datalake: %+v", testDatalake)
	assert.NotNil(t, testDatalake)

	datalakeByID, err := client.GetDatalake(testDatalake.ID)
	if err != nil {
		t.Fatalf("Failed to get datalake by ID: %v", err)
	}
	t.Logf("Datalake by ID: %+v", datalakeByID)
	assert.NotNil(t, datalakeByID)
	assert.Equal(t, testDatalake.ID, datalakeByID.ID)
	assert.Equal(t, testDatalake.Name, datalakeByID.Name)

	datalakesByEnvironment, err := client.GetDatalakesByHostingEnvironment(createdEnvironment.ID)
	if err != nil {
		t.Fatalf("Failed to get datalakes by hosting environment: %v", err)
	}
	t.Logf("Datalakes by hosting environment: %+v", datalakesByEnvironment)
	assert.NotNil(t, datalakesByEnvironment)
	assert.NotEmpty(t, datalakesByEnvironment)

	found := false
	for _, dl := range datalakesByEnvironment {
		if dl.ID == testDatalake.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test datalake should be found in hosting environment datalakes")

	testDatalake.Status = DatalakeStatusReady
	updatedDatalake, err := client.UpdateDatalake(testDatalake.ID, testDatalake)
	if err != nil {
		t.Fatalf("Failed to update datalake: %v", err)
	}

	t.Logf("Updated datalake: %+v", updatedDatalake)
	assert.NotNil(t, updatedDatalake)
	assert.Equal(t, testDatalake.Name, updatedDatalake.Name)
	assert.Equal(t, DatalakeStatusReady, updatedDatalake.Status)

	err = client.DeleteDatalake(testDatalake.ID)
	if err != nil {
		t.Fatalf("Failed to delete datalake: %v", err)
	}

	t.Logf("Deleted datalake: %+v", testDatalake)
}

func TestDatalakeValidation(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetDatalakesByHostingEnvironment with empty ID
	_, err = client.GetDatalakesByHostingEnvironment("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hosting environment ID cannot be empty")

	// Test GetDatalakesByHostingEnvironment with invalid UUID
	_, err = client.GetDatalakesByHostingEnvironment("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test GetDatalake with empty ID
	_, err = client.GetDatalake("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test GetDatalake with invalid UUID
	_, err = client.GetDatalake("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test UpdateDatalake with empty ID
	datalake := Datalake{Name: "test"}
	_, err = client.UpdateDatalake("", datalake)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test UpdateDatalake with invalid UUID
	_, err = client.UpdateDatalake("invalid-uuid", datalake)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test DeleteDatalake with empty ID
	err = client.DeleteDatalake("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test DeleteDatalake with invalid UUID
	err = client.DeleteDatalake("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")
}