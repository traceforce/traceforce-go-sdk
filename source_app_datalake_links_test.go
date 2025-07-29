package traceforce

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceAppDatalakeLinks(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// First create a hosting environment
	environmentReq := CreateHostingEnvironmentRequest{
		Name:          "test hosting environment for links",
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

	// Create a datalake
	datalakeReq := CreateDatalakeRequest{
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 DatalakeTypeBigQuery,
		Name:                 "test datalake for links",
		EnvironmentNativeID:  "test-project-id",
		Region:               "us-central1",
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

	// Create a source app
	sourceAppReq := CreateSourceAppRequest{
		HostingEnvironmentID: createdEnvironment.ID,
		Type:                 SourceAppTypeSalesforce,
		Name:                 "test source app for links",
	}

	createdSourceApp, err := client.CreateSourceApp(sourceAppReq)
	if err != nil {
		t.Fatalf("Failed to create source app: %v", err)
	}
	defer func() {
		err := client.DeleteSourceApp(createdSourceApp.ID)
		if err != nil {
			t.Logf("Failed to cleanup source app: %v", err)
		}
	}()

	// Create a source app datalake link
	linkReq := CreateSourceAppDatalakeLinkRequest{
		SourceAppID: createdSourceApp.ID,
		DatalakeID:  createdDatalake.ID,
	}

	createdLink, err := client.CreateSourceAppDatalakeLink(linkReq)
	if err != nil {
		t.Fatalf("Failed to create source app datalake link: %v", err)
	}

	t.Logf("Created link: %+v", createdLink)
	assert.NotNil(t, createdLink)
	assert.Equal(t, linkReq.SourceAppID, createdLink.SourceAppID)
	assert.Equal(t, linkReq.DatalakeID, createdLink.DatalakeID)
	assert.Equal(t, createdEnvironment.ID, createdLink.HostingEnvironmentID)
	assert.NotEmpty(t, createdLink.ID)

	// Test GetSourceAppDatalakeLinks
	links, err := client.GetSourceAppDatalakeLinks()
	if err != nil {
		t.Fatalf("Failed to get source app datalake links: %v", err)
	}

	t.Logf("Links: %+v", links)
	assert.NotNil(t, links)
	assert.NotEmpty(t, links)

	var testLink SourceAppDatalakeLink
	for _, link := range links {
		t.Logf("Link: %+v", link)
		assert.NotNil(t, link.ID)
		assert.NotEmpty(t, link.SourceAppID)
		assert.NotEmpty(t, link.DatalakeID)
		assert.NotEmpty(t, link.HostingEnvironmentID)

		if link.ID == createdLink.ID {
			testLink = link
		}
	}

	t.Logf("Test link: %+v", testLink)
	assert.NotNil(t, testLink)

	// Test GetSourceAppDatalakeLink by ID
	linkByID, err := client.GetSourceAppDatalakeLink(testLink.ID)
	if err != nil {
		t.Fatalf("Failed to get source app datalake link by ID: %v", err)
	}
	t.Logf("Link by ID: %+v", linkByID)
	assert.NotNil(t, linkByID)
	assert.Equal(t, testLink.ID, linkByID.ID)
	assert.Equal(t, testLink.SourceAppID, linkByID.SourceAppID)
	assert.Equal(t, testLink.DatalakeID, linkByID.DatalakeID)

	// Test GetSourceAppDatalakeLinksBySourceApp
	linksBySourceApp, err := client.GetSourceAppDatalakeLinksBySourceApp(createdSourceApp.ID)
	if err != nil {
		t.Fatalf("Failed to get source app datalake links by source app: %v", err)
	}
	t.Logf("Links by source app: %+v", linksBySourceApp)
	assert.NotNil(t, linksBySourceApp)
	assert.NotEmpty(t, linksBySourceApp)

	found := false
	for _, link := range linksBySourceApp {
		if link.ID == testLink.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test link should be found in source app links")

	// Test GetSourceAppDatalakeLinksByDatalake
	linksByDatalake, err := client.GetSourceAppDatalakeLinksByDatalake(createdDatalake.ID)
	if err != nil {
		t.Fatalf("Failed to get source app datalake links by datalake: %v", err)
	}
	t.Logf("Links by datalake: %+v", linksByDatalake)
	assert.NotNil(t, linksByDatalake)
	assert.NotEmpty(t, linksByDatalake)

	found = false
	for _, link := range linksByDatalake {
		if link.ID == testLink.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test link should be found in datalake links")

	// Test DeleteSourceAppDatalakeLink
	err = client.DeleteSourceAppDatalakeLink(testLink.ID)
	if err != nil {
		t.Fatalf("Failed to delete source app datalake link: %v", err)
	}

	t.Logf("Deleted link: %+v", testLink)
}

func TestSourceAppDatalakeLinkValidation(t *testing.T) {
	client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test CreateSourceAppDatalakeLink with empty source app ID
	_, err = client.CreateSourceAppDatalakeLink(CreateSourceAppDatalakeLinkRequest{
		SourceAppID: "",
		DatalakeID:  "123e4567-e89b-12d3-a456-426614174000",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source app ID cannot be empty")

	// Test CreateSourceAppDatalakeLink with empty datalake ID
	_, err = client.CreateSourceAppDatalakeLink(CreateSourceAppDatalakeLinkRequest{
		SourceAppID: "123e4567-e89b-12d3-a456-426614174000",
		DatalakeID:  "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datalake ID cannot be empty")

	// Test CreateSourceAppDatalakeLink with invalid source app ID UUID
	_, err = client.CreateSourceAppDatalakeLink(CreateSourceAppDatalakeLinkRequest{
		SourceAppID: "invalid-uuid",
		DatalakeID:  "123e4567-e89b-12d3-a456-426614174000",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source app ID UUID format")

	// Test CreateSourceAppDatalakeLink with invalid datalake ID UUID
	_, err = client.CreateSourceAppDatalakeLink(CreateSourceAppDatalakeLinkRequest{
		SourceAppID: "123e4567-e89b-12d3-a456-426614174000",
		DatalakeID:  "invalid-uuid",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid datalake ID UUID format")

	// Test GetSourceAppDatalakeLinksBySourceApp with empty ID
	_, err = client.GetSourceAppDatalakeLinksBySourceApp("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source app ID cannot be empty")

	// Test GetSourceAppDatalakeLinksBySourceApp with invalid UUID
	_, err = client.GetSourceAppDatalakeLinksBySourceApp("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test GetSourceAppDatalakeLinksByDatalake with empty ID
	_, err = client.GetSourceAppDatalakeLinksByDatalake("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datalake ID cannot be empty")

	// Test GetSourceAppDatalakeLinksByDatalake with invalid UUID
	_, err = client.GetSourceAppDatalakeLinksByDatalake("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test GetSourceAppDatalakeLink with empty ID
	_, err = client.GetSourceAppDatalakeLink("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test GetSourceAppDatalakeLink with invalid UUID
	_, err = client.GetSourceAppDatalakeLink("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")

	// Test DeleteSourceAppDatalakeLink with empty ID
	err = client.DeleteSourceAppDatalakeLink("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")

	// Test DeleteSourceAppDatalakeLink with invalid UUID
	err = client.DeleteSourceAppDatalakeLink("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID format")
}