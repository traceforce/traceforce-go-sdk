package traceforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests use a fake server (no network, no API key needed).

const testConnID = "3b0c6d1e-1111-4222-8333-444455556666"

const notFoundBody = `{"error":"Connection not found","code":"connection_not_found"}`

var testCreds = Credentials{ClientID: "client-1", ClientSecret: "secret-1", DeviceID: "cloud-collector-aws"}

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

// connJSON is a connection as the API answers it.
const connJSON = `{"id":"` + testConnID + `","provider":"aws","account":"522814721969","name":null,"status":1,
	"moduleVersion":"1.2.0","moduleCommit":null,"collectorId":"arn:aws:lambda:x","identityId":null,
	"regions":["us-east-1"],"deployedAt":"2026-10-30T09:41:25.817Z"}`

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
		_, _ = w.Write([]byte(connJSON))
	}))
	defer srv.Close()

	c, err := NewClientWithCredentials(testCreds, srv.URL, nil)
	assert.NoError(t, err)
	_, err = c.CreateCloudConnection(CreateCloudConnectionRequest{Provider: "aws", Account: "522814721969", ModuleVersion: "1.2.0"})
	assert.NoError(t, err)

	assert.Len(t, got, 2)
	assert.Equal(t, seen{"POST", "/api-keys", "", map[string]any{
		"client_id": "client-1", "client_secret": "secret-1", "device_id": "cloud-collector-aws",
	}}, got[0], "login sends id, secret and device id, without Authorization")
	assert.Equal(t, "Bearer tok-123", got[1].auth, "later calls carry the login token")
}

func TestExpiredTokenLogsInAgainOnce(t *testing.T) {
	var got []seen
	logins := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, seen{r.Method, r.URL.Path, r.Header.Get("Authorization"), nil})
		switch {
		case r.URL.Path == "/api-keys":
			logins++
			_, _ = w.Write([]byte(`{"access_token":"tok-` + string(rune('0'+logins)) + `"}`))
		case r.Header.Get("Authorization") == "Bearer tok-2":
			_, _ = w.Write([]byte(connJSON))
		default: // the first token has expired
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Token expired"}`))
		}
	}))
	defer srv.Close()

	c, err := NewClientWithCredentials(testCreds, srv.URL, nil)
	assert.NoError(t, err)
	conn, err := c.GetCloudConnection(testConnID)
	assert.NoError(t, err)
	assert.Equal(t, testConnID, conn.ID)
	assert.Equal(t, 2, logins, "one new login after the 401")
	assert.Equal(t, []string{"/api-keys", "/cloud-connections/" + testConnID, "/api-keys", "/cloud-connections/" + testConnID},
		[]string{got[0].path, got[1].path, got[2].path, got[3].path}, "the request is repeated once with the new token")
	assert.Equal(t, "", got[2].auth, "the renewal login is sent without the expired token")
}

func TestUnauthorizedWithoutCredentialsIsAnError(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusUnauthorized, `{"error":"Token expired"}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	_, err := c.GetCloudConnection(testConnID)
	assert.EqualError(t, err, `HTTP 401: {"error":"Token expired"}`)
	assert.Len(t, got, 1, "a client made from a token does not log in")
}

func TestNewClientWithCredentialsRefused(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusUnauthorized, `{"error":"Invalid client credentials"}`, &got)

	c, err := NewClientWithCredentials(testCreds, srv.URL, nil)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "login with the API client failed: HTTP 401")
	assert.Len(t, got, 1, "a refused login is not repeated")
}

func TestNewClientWithCredentialsNoToken(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{}`, &got)

	c, err := NewClientWithCredentials(testCreds, srv.URL, nil)
	assert.Nil(t, c)
	assert.EqualError(t, err, "login with the API client returned no token")
}

func TestCreateCloudConnection(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, strings.Replace(connJSON, `"id"`, `"created":true,"id"`, 1), &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.CreateCloudConnection(CreateCloudConnectionRequest{
		Provider: "aws", Account: "522814721969", Name: "Prod", ModuleVersion: "1.2.0", ModuleCommit: "abc123",
	})
	assert.NoError(t, err)
	assert.Equal(t, testConnID, conn.ID)
	assert.True(t, conn.Created)
	assert.Equal(t, CloudConnectionStatusPending, conn.Status)
	assert.Equal(t, seen{"POST", "/cloud-connections", "Bearer tok", map[string]any{
		"provider": "aws", "account": "522814721969", "name": "Prod", "moduleVersion": "1.2.0", "moduleCommit": "abc123",
	}}, got[0])
}

func TestCreateCloudConnectionWithoutIDIsAnError(t *testing.T) {
	for _, reply := range []string{`{}`, `{"id":""}`, `{"id":"abc"}`} {
		var got []seen
		srv := fakeAPI(t, http.StatusOK, reply, &got)
		c, _ := NewClient("tok", srv.URL, nil)

		conn, err := c.CreateCloudConnection(CreateCloudConnectionRequest{Provider: "aws", Account: "1", ModuleVersion: "1.2.0"})
		assert.Nil(t, conn, reply)
		assert.ErrorContains(t, err, "the API answered no valid id", reply)
	}
}

func TestCreateCloudConnectionOldBundle(t *testing.T) {
	var got []seen
	reply := `{"error":"Collector 1.0.0 is older than 1.2.0.","code":"collector_too_old","moduleVersion":"1.0.0","minVersion":"1.2.0"}`
	srv := fakeAPI(t, http.StatusUpgradeRequired, reply, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	_, err := c.CreateCloudConnection(CreateCloudConnectionRequest{Provider: "aws", Account: "1", ModuleVersion: "1.0.0"})
	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr), "426 comes back as *APIError")
	assert.Equal(t, http.StatusUpgradeRequired, apiErr.StatusCode)
	assert.Equal(t, CodeCollectorTooOld, apiErr.Code)
	assert.EqualError(t, err, "HTTP 426: "+reply)
}

func TestGetCloudConnection(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, connJSON, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.GetCloudConnection(testConnID)
	assert.NoError(t, err)
	assert.Equal(t, "GET /cloud-connections/"+testConnID, got[0].method+" "+got[0].path)
	assert.Equal(t, "aws", conn.Provider)
	assert.Equal(t, "522814721969", conn.Account)
	assert.Nil(t, conn.Name, "null stays nil")
	assert.Equal(t, CloudConnectionStatusPending, conn.Status)
	assert.Equal(t, "1.2.0", *conn.ModuleVersion)
	assert.Equal(t, "arn:aws:lambda:x", *conn.CollectorID)
	assert.Equal(t, []string{"us-east-1"}, conn.Regions)
	assert.Equal(t, 2026, conn.DeployedAt.Year(), "deployedAt is read as a time")
}

func TestGetCloudConnectionAnswers(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		reply   string
		gone    bool
		wantErr string
	}{
		{"gone: the API's own 404", http.StatusNotFound, notFoundBody, true, ""},
		{"a 404 without the code (wrong URL, old backend) is an error", http.StatusNotFound, `<html>Not Found</html>`, false, "HTTP 404"},
		{"a 404 with another JSON body is an error", http.StatusNotFound, `{"error":"Not found"}`, false, "HTTP 404"},
		{"other cloud's connection", http.StatusForbidden, `{"error":"This connection belongs to another cloud's collector"}`, false, "HTTP 403"},
		{"server error", http.StatusInternalServerError, `{"error":"Failed"}`, false, "HTTP 500"},
		{"a redirect is not success", http.StatusNotModified, ``, false, "HTTP 304"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got []seen
			srv := fakeAPI(t, tc.status, tc.reply, &got)
			c, _ := NewClient("tok", srv.URL, nil)

			conn, err := c.GetCloudConnection(testConnID)
			assert.Nil(t, conn)
			if tc.gone {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestPutCloudConnectionDeployment(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, connJSON, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.PutCloudConnectionDeployment(testConnID, CloudConnectionDeployment{
		CollectorID: "arn:aws:lambda:us-east-1:1:function:C", IdentityID: "arn:aws:iam::1:role/R", Regions: []string{"us-east-1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2026, conn.DeployedAt.Year())
	assert.Equal(t, seen{"PUT", "/cloud-connections/" + testConnID + "/deployment", "Bearer tok", map[string]any{
		"collectorId": "arn:aws:lambda:us-east-1:1:function:C", "identityId": "arn:aws:iam::1:role/R", "regions": []any{"us-east-1"},
	}}, got[0])
}

func TestPutCloudConnectionDeploymentNotFound(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusNotFound, notFoundBody, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.PutCloudConnectionDeployment(testConnID, CloudConnectionDeployment{CollectorID: "c", IdentityID: "i"})
	assert.Nil(t, conn)
	assert.ErrorContains(t, err, "not found", "nothing was recorded, so it is an error")
}

func TestDeleteCloudConnectionDeployment(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, strings.Replace(connJSON, `"deployedAt":"2026-10-30T09:41:25.817Z"`, `"deployedAt":null`, 1), &got)
	c, _ := NewClient("tok", srv.URL, nil)

	conn, err := c.DeleteCloudConnectionDeployment(testConnID)
	assert.NoError(t, err)
	assert.Nil(t, conn.DeployedAt)
	assert.Equal(t, "DELETE /cloud-connections/"+testConnID+"/deployment", got[0].method+" "+got[0].path)

	gone := fakeAPI(t, http.StatusNotFound, notFoundBody, &got)
	c, _ = NewClient("tok", gone.URL, nil)
	conn, err = c.DeleteCloudConnectionDeployment(testConnID)
	assert.NoError(t, err, "a connection that is gone has nothing to clear")
	assert.Nil(t, conn)
}

func TestDeleteCloudConnection(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		reply   string
		wantErr bool
	}{
		{"disconnected", http.StatusOK, `{"disconnected":true}`, false},
		{"already gone counts as done", http.StatusNotFound, notFoundBody, false},
		{"a 404 that is not the API's (wrong URL) is an error", http.StatusNotFound, `Not Found`, true},
		{"server error is returned", http.StatusInternalServerError, `{}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got []seen
			srv := fakeAPI(t, tc.status, tc.reply, &got)
			c, _ := NewClient("tok", srv.URL, nil)

			err := c.DeleteCloudConnection(testConnID)
			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, "DELETE /cloud-connections/"+testConnID, got[0].method+" "+got[0].path)
		})
	}
}

func TestNoContentIsSuccess(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusNoContent, ``, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	assert.NoError(t, c.DeleteCloudConnection(testConnID))
	conn, err := c.GetCloudConnection(testConnID)
	assert.NoError(t, err, "a 204 is not read as JSON")
	assert.Equal(t, "", conn.ID)
}

func TestCloudConnectionIDIsSentCanonical(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, connJSON, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	for _, id := range []string{"{3B0C6D1E-1111-4222-8333-444455556666}", "urn:uuid:" + testConnID, strings.ReplaceAll(testConnID, "-", "")} {
		_, err := c.GetCloudConnection(id)
		assert.NoError(t, err)
		assert.Equal(t, "/cloud-connections/"+testConnID, got[len(got)-1].path, id)
	}
}

func TestCloudConnectionBadIDSendsNothing(t *testing.T) {
	var got []seen
	srv := fakeAPI(t, http.StatusOK, `{}`, &got)
	c, _ := NewClient("tok", srv.URL, nil)

	_, err := c.GetCloudConnection("abc")
	assert.ErrorContains(t, err, `invalid cloud connection id "abc"`)
	_, err = c.PutCloudConnectionDeployment("abc", CloudConnectionDeployment{})
	assert.ErrorContains(t, err, "invalid cloud connection id")
	_, err = c.DeleteCloudConnectionDeployment("")
	assert.ErrorContains(t, err, "invalid cloud connection id")
	assert.ErrorContains(t, c.DeleteCloudConnection("../api-keys"), "invalid cloud connection id")
	assert.Empty(t, got, "no request is sent for a bad id")
}

func TestValidateResponseIsAnAPIError(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusBadRequest)
	_, _ = rec.WriteString(`{"error":"bad","code":"x"}`)
	err := validateResponse(rec.Result())

	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr), "the existing calls return *APIError too")
	assert.Equal(t, "x", apiErr.Code)
	assert.EqualError(t, err, `HTTP 400: {"error":"bad","code":"x"}`, "same text as before")

	ok := httptest.NewRecorder()
	assert.NoError(t, validateResponse(ok.Result()))
}

func TestBuildHeadersWithoutKey(t *testing.T) {
	c, _ := NewClient("", "https://example.com", &ClientOptions{ExtraHeaders: map[string]string{"x-a": "1"}})
	headers := c.buildHeaders()
	_, hasAuth := headers["Authorization"]
	assert.False(t, hasAuth, "no Authorization header without a key (it was \"Bearer \" before)")
	assert.Equal(t, "1", headers["x-a"])
}

func TestCloudCollectorDeviceID(t *testing.T) {
	assert.Equal(t, "cloud-collector-aws", CloudCollectorDeviceID("aws"))
	assert.Equal(t, "cloud-collector-m365", CloudCollectorDeviceID("m365"))
}

func TestExpiredTokenIsRenewedOnceForManyCalls(t *testing.T) {
	var mu sync.Mutex
	logins := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path == "/api-keys" {
			logins++
			_, _ = w.Write([]byte(fmt.Sprintf(`{"access_token":"tok-%d"}`, logins)))
			return
		}
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer tok-%d", logins) || logins == 1 {
			w.WriteHeader(http.StatusUnauthorized) // only the newest token, and not the first one, is valid
			return
		}
		_, _ = w.Write([]byte(connJSON))
	}))
	defer srv.Close()

	c, err := NewClientWithCredentials(testCreds, srv.URL, nil)
	assert.NoError(t, err)
	var wg sync.WaitGroup
	start := make(chan struct{}) // all calls start together, so their 401s overlap
	errs := make([]error, 10)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = c.GetCloudConnection(testConnID)
		}(i)
	}
	close(start)
	wg.Wait()
	for _, err := range errs {
		assert.NoError(t, err, "every call succeeds with the renewed token")
	}
	assert.Equal(t, 2, logins, "10 calls that get 401 at once make one new login, not 10")
}

func TestExistingCallsUseTheSameSuccessRule(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusNotModified)
	var apiErr *APIError
	assert.True(t, errors.As(validateResponse(rec.Result()), &apiErr), "a 3xx is an *APIError on the existing calls too")
}

func TestFailedRenewalIsSharedByTheWaitingCalls(t *testing.T) {
	var mu sync.Mutex
	logins := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path == "/api-keys" {
			logins++
			if logins > 1 { // the login service is down after the first login
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"tok-1"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized) // the token has expired
	}))
	defer srv.Close()

	c, err := NewClientWithCredentials(testCreds, srv.URL, nil)
	assert.NoError(t, err)
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, 10)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = c.GetCloudConnection(testConnID)
		}(i)
	}
	close(start)
	wg.Wait()
	for _, err := range errs {
		assert.ErrorContains(t, err, "login with the API client failed: HTTP 503")
	}
	assert.Equal(t, 2, logins, "10 calls waiting on one failed renewal share its error: one failed login, not 10")
}
