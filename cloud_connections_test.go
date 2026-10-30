package traceforce

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests use a fake server (no network, no API key needed).

const testConnID = "3b0c6d1e-1111-4222-8333-444455556666"

// seen is one request the fake server got.
type seen struct {
	method, path, auth string
	body               map[string]any
}

// fakeAPI answers every request with status + reply and records it in got.
func fakeAPI(t *testing.T, status int, reply string, got *[]seen) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		*got = append(*got, seen{r.Method, r.URL.Path, r.Header.Get("Authorization"), body})
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestNewClientWithCredentialsLogsInThenUsesToken(t *testing.T) {
	var got []seen
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		got = append(got, seen{r.Method, r.URL.Path, r.Header.Get("Authorization"), body})
		if r.URL.Path == "/api-keys" {
			_, _ = w.Write([]byte(`{"access_token":"tok-123"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"` + testConnID + `"}`))
	}))
	defer srv.Close()

	c, err := NewClientWithCredentials("client-1", "secret-1", "cloud-collector-aws", srv.URL, nil)
	assert.NoError(t, err)
	_, err = c.CreateCloudConnection(CreateCloudConnectionRequest{Provider: "aws", Account: "522814721969", ModuleVersion: "1.2.0"})
	assert.NoError(t, err)

	assert.Len(t, got, 2)
	assert.Equal(t, seen{"POST", "/api-keys", "", map[string]any{
		"client_id": "client-1", "client_secret": "secret-1", "device_id": "cloud-collector-aws",
	}}, got[0], "login sends id, secret and device id, without Authorization")
	assert.Equal(t, "Bearer tok-123", got[1].auth, "later calls carry the login token")
}

func TestNewClientWithCredentialsRefused(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusUnauthorized, `{"error":"Invalid client credentials"}`, &got)

	c, err := NewClientWithCredentials("client-1", "wrong", "cloud-collector-aws", srv.URL, nil)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "login with the API client failed: HTTP 401")
}

func TestNewClientWithCredentialsNoToken(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{}`, &got)

	c, err := NewClientWithCredentials("client-1", "secret-1", "cloud-collector-aws", srv.URL, nil)
	assert.Nil(t, c)
	assert.EqualError(t, err, "login with the API client returned no token")
}

func TestCreateCloudConnection(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{"id":"`+testConnID+`"}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	id, err := c.CreateCloudConnection(CreateCloudConnectionRequest{
		Provider: "aws", Account: "522814721969", Name: "Prod", ModuleVersion: "1.2.0", ModuleCommit: "abc123",
	})
	assert.NoError(t, err)
	assert.Equal(t, testConnID, id)
	assert.Equal(t, seen{"POST", "/cloud-connections", "Bearer tok", map[string]any{
		"provider": "aws", "account": "522814721969", "name": "Prod", "moduleVersion": "1.2.0", "moduleCommit": "abc123",
	}}, got[0])
}

func TestCreateCloudConnectionOldBundle(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusUpgradeRequired, `{"error":"Collector 1.0.0 is older than 1.2.0."}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	_, err := c.CreateCloudConnection(CreateCloudConnectionRequest{Provider: "aws", Account: "1", ModuleVersion: "1.0.0"})
	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr), "426 comes back as *APIError")
	assert.Equal(t, http.StatusUpgradeRequired, apiErr.StatusCode)
	assert.EqualError(t, err, `HTTP 426: {"error":"Collector 1.0.0 is older than 1.2.0."}`, "same text as validateResponse")
}

func TestGetCloudConnection(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{"id":"`+testConnID+`","account_ref":"522814721969","status":1,
		"module_version":"1.2.0","module_commit":null,"collector_id":"arn:aws:lambda:x","identity_id":null,
		"deployed_at":"2026-10-30T09:41:25.817Z"}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.GetCloudConnection(testConnID)
	assert.NoError(t, err)
	assert.Equal(t, "GET /cloud-connections/"+testConnID, got[0].method+" "+got[0].path)
	assert.Equal(t, "522814721969", conn.AccountRef)
	assert.Equal(t, 1, conn.Status)
	assert.Equal(t, "1.2.0", *conn.ModuleVersion)
	assert.Nil(t, conn.ModuleCommit, "null stays nil")
	assert.Equal(t, "arn:aws:lambda:x", *conn.CollectorID)
	assert.Equal(t, 2026, conn.DeployedAt.Year(), "deployed_at is read as a time")
}

func TestGetCloudConnectionGone(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusNotFound, `{"error":"Connection not found"}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.GetCloudConnection(testConnID)
	assert.NoError(t, err, "404 is not an error: the connection is gone")
	assert.Nil(t, conn)
}

func TestPutCloudConnectionDeployment(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{"deployed":true}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	err := c.PutCloudConnectionDeployment(testConnID, CloudConnectionDeployment{
		CollectorID: "arn:aws:lambda:us-east-1:1:function:C", IdentityID: "arn:aws:iam::1:role/R", Regions: []string{"us-east-1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, seen{"PUT", "/cloud-connections/" + testConnID + "/deployment", "Bearer tok", map[string]any{
		"collectorId": "arn:aws:lambda:us-east-1:1:function:C", "identityId": "arn:aws:iam::1:role/R", "regions": []any{"us-east-1"},
	}}, got[0])
}

func TestDeleteCloudConnection(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"disconnected", http.StatusOK, false},
		{"already gone counts as done", http.StatusNotFound, false},
		{"server error is returned", http.StatusInternalServerError, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got []seen
			srv := fakeAPI(t, tc.status, `{}`, &got)
			c, _ := NewClient("tok", srv.URL, nil)

			err := c.DeleteCloudConnection(testConnID)
			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, "DELETE /cloud-connections/"+testConnID, got[0].method+" "+got[0].path)
		})
	}
}

func TestCloudConnectionBadIDSendsNothing(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	_, err := c.GetCloudConnection("abc")
	assert.ErrorContains(t, err, `invalid cloud connection id "abc"`)
	assert.ErrorContains(t, c.PutCloudConnectionDeployment("abc", CloudConnectionDeployment{}), "invalid cloud connection id")
	assert.ErrorContains(t, c.DeleteCloudConnection("../api-keys"), "invalid cloud connection id")
	assert.Empty(t, got, "no request is sent for a bad id")
}
