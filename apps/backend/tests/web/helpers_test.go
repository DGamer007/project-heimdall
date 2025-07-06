package web_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"heimdall/backend/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

// Test Constants for better reusability
const (
	// Test passwords
	defaultTestPassword = "testpassword123"
	shortPassword       = "short"

	// Common test data
	defaultFirstName = "Test"
	defaultLastName  = "User"

	// Large data for boundary testing
	largeStringSize = 10000
)

// API endpoints - centralized for easy maintenance
var apiEndpoints = struct {
	Register string
	Login    string
	Logout   string
	Refresh  string
	Health   string
}{
	Register: "/api/v1/auth/local/register",
	Login:    "/api/v1/auth/local/login",
	Logout:   "/api/v1/auth/logout",
	Refresh:  "/api/v1/auth/refresh",
	Health:   "/healthz",
}

// Standard HTTP headers
const (
	headerAuthorization = "Authorization"
	headerAccessToken   = "X-Access-Token" // Used for refresh endpoint
	headerContentType   = "Content-Type"

	bearerPrefix = "Bearer "
	basicPrefix  = "Basic "
)

// Test cleanup helper functions
func cleanupUser(t *testing.T, email string) {
	t.Helper()
	_, err := testDataStore.Postgres.Exec("DELETE FROM core.sessions WHERE user_id IN (SELECT id FROM core.users WHERE email = $1)", email)
	if err != nil {
		t.Logf("Warning: Failed to cleanup sessions for user %s: %v", email, err)
	}
	_, err = testDataStore.Postgres.Exec("DELETE FROM core.users WHERE email = $1", email)
	if err != nil {
		t.Logf("Warning: Failed to cleanup user %s: %v", email, err)
	}
}

// Test helper functions for user creation and authentication
type TestUser struct {
	Email    string
	Username string
	Password string
}

func createTestUser(t *testing.T, email, username, password string) *TestUser {
	t.Helper()

	user := &TestUser{
		Email:    email,
		Username: username,
		Password: password,
	}

	// Setup cleanup
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	// Create user
	registerPayload, err := json.Marshal(dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  username,
		Password:  password,
		FirstName: defaultFirstName,
		LastName:  defaultLastName,
	})
	assert.NoError(t, err, "Failed to marshal register payload")

	resp, err := http.Post(testServer.URL+apiEndpoints.Register, "application/json", bytes.NewBuffer(registerPayload))
	assert.NoError(t, err, "Failed to register test user")
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "User registration should succeed")

	return user
}

func createAndLoginUser(t *testing.T, email, username, password string) (*TestUser, *dto.LoginWithPasswordResponse) {
	t.Helper()

	// Create user
	user := createTestUser(t, email, username, password)

	// Login to get tokens
	loginResp := loginUser(t, username, password)

	return user, loginResp
}

func loginUser(t *testing.T, identifier, password string) *dto.LoginWithPasswordResponse {
	t.Helper()

	payload, err := json.Marshal(dto.LoginWithPasswordPayload{
		Identifier: identifier,
		Password:   password,
	})
	assert.NoError(t, err, "Failed to marshal login payload")

	resp, err := http.Post(testServer.URL+apiEndpoints.Login, "application/json", bytes.NewBuffer(payload))
	assert.NoError(t, err, "Failed to send login request")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Login should succeed")

	var loginResponse dto.LoginWithPasswordResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	assert.NoError(t, err, "Failed to decode login response")

	return &loginResponse
}

// HTTP Test Helpers

func expectErrorResponse(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()

	assert.Equal(t, expectedStatus, resp.StatusCode, "Expected status %d", expectedStatus)

	var errorResponse dto.ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errorResponse)
	assert.NoError(t, err, "Failed to decode error response")

	assert.False(t, errorResponse.Success, "Expected success=false")
	assert.NotEmpty(t, errorResponse.Message, "Expected error message")

	if expectedStatus == http.StatusBadRequest {
		assert.NotEmpty(t, errorResponse.Errors, "Expected validation errors")
	}
}

// Optimized HTTP request helpers with reusable client
var httpClient = &http.Client{}

func postJSON(t *testing.T, endpoint string, payload any) *http.Response {
	t.Helper()

	jsonData, err := json.Marshal(payload)
	assert.NoError(t, err, "Failed to marshal payload")

	resp, err := http.Post(testServer.URL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err, "Failed to send POST request")

	return resp
}

func postJSONString(t *testing.T, endpoint string, jsonString string) *http.Response {
	t.Helper()

	resp, err := http.Post(testServer.URL+endpoint, "application/json", bytes.NewBufferString(jsonString))
	assert.NoError(t, err, "Failed to send POST request")

	return resp
}

func postEmptyBody(t *testing.T, endpoint string) *http.Response {
	t.Helper()

	resp, err := http.Post(testServer.URL+endpoint, "application/json", &bytes.Buffer{})
	assert.NoError(t, err, "Failed to send POST request")

	return resp
}

func postWithContentType(t *testing.T, endpoint string, payload any, contentType string) *http.Response {
	t.Helper()

	jsonData, err := json.Marshal(payload)
	assert.NoError(t, err, "Failed to marshal payload")

	req, err := http.NewRequest("POST", testServer.URL+endpoint, bytes.NewBuffer(jsonData))
	assert.NoError(t, err, "Failed to create POST request")
	req.Header.Set(headerContentType, contentType)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err, "Failed to send POST request")

	return resp
}

// Helper for refresh endpoint (different auth pattern)
func postRefreshRequest(t *testing.T, refreshToken, accessToken string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("POST", testServer.URL+apiEndpoints.Refresh, nil)
	assert.NoError(t, err, "Failed to create refresh request")

	if refreshToken != "" {
		req.Header.Set(headerAuthorization, refreshToken) // No Bearer prefix for refresh
	}
	if accessToken != "" {
		req.Header.Set(headerAccessToken, accessToken)
	}

	req.Header.Set(headerContentType, "application/json")

	resp, err := httpClient.Do(req)
	assert.NoError(t, err, "Failed to send refresh request")

	return resp
}

func expectSpecificWrongMethod(t *testing.T, endpoint, method string, expectedStatus int) {
	t.Helper()

	req, err := http.NewRequest(method, testServer.URL+endpoint, nil)
	assert.NoError(t, err, "Failed to create %s request", method)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err, "Failed to send %s request", method)
	defer resp.Body.Close()

	assert.Equal(t, expectedStatus, resp.StatusCode, "Expected %d for %s method on %s", expectedStatus, method, endpoint)
}

// Validation Test Helpers - using constants for consistency

var registerBaseFields = map[string]string{
	"email":     "test@example.com",
	"userName":  "testuser",
	"password":  defaultTestPassword,
	"firstName": defaultFirstName,
	"lastName":  defaultLastName,
}

var loginBaseFields = map[string]string{
	"identifier": "test@example.com",
	"password":   defaultTestPassword,
}

func testMissingField(t *testing.T, endpoint string, baseFields map[string]string, missingField string) {
	t.Helper()

	payload := make(map[string]string)
	for k, v := range baseFields {
		if k != missingField {
			payload[k] = v
		}
	}

	resp := postJSON(t, endpoint, payload)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func testValidationError(t *testing.T, endpoint string, payload any) {
	t.Helper()

	resp := postJSON(t, endpoint, payload)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func testMalformedJSON(t *testing.T, endpoint string, malformedJSON string) {
	t.Helper()

	resp := postJSONString(t, endpoint, malformedJSON)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func testEmptyBody(t *testing.T, endpoint string) {
	t.Helper()

	resp := postEmptyBody(t, endpoint)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func testWrongContentType(t *testing.T, endpoint string, payload any) {
	t.Helper()

	resp := postWithContentType(t, endpoint, payload, "text/plain")
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func parseJSONResponse(body io.Reader, target any) error {
	return json.NewDecoder(body).Decode(target)
}
