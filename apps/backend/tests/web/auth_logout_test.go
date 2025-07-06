package web_test

import (
	"net/http"
	"testing"

	"heimdall/backend/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// LOGOUT API TESTS
// ==========================================

// Success Tests

func TestLogoutSuccess(t *testing.T) {
	email := "logout_success@test.com"
	username := "logoutsuccess"
	password := "logoutpassword123"

	// Create and login user
	_, loginResponse := createAndLoginUser(t, email, username, password)

	// Test logout using standard Bearer auth
	req, err := http.NewRequest("POST", testServer.URL+apiEndpoints.Logout, nil)
	assert.NoError(t, err, "Failed to create logout request")

	req.Header.Set(headerAuthorization, bearerPrefix+loginResponse.Data.AccessToken)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err, "Failed to send logout request")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var logoutResponse map[string]any
	err = parseJSONResponse(resp.Body, &logoutResponse)
	assert.NoError(t, err, "Failed to decode logout response")

	assert.Equal(t, true, logoutResponse["success"], "Expected success=true")
	assert.Equal(t, "Successfully logged out", logoutResponse["message"], "Expected logout success message")
}

// Error Tests

func TestLogoutAuthenticationErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupUser   bool
		email       string
		username    string
		password    string
		authHeader  string
		useToken    bool
		tokenValue  string
		description string
	}{
		{
			name:        "no_authorization_header",
			setupUser:   true,
			email:       "logout_no_auth@test.com",
			username:    "logoutnoauth",
			password:    "logoutpassword123",
			authHeader:  "",
			useToken:    false,
			description: "should return unauthorized when authorization header is missing",
		},
		{
			name:        "invalid_token_format_basic",
			setupUser:   true,
			email:       "logout_invalid_format@test.com",
			username:    "logoutinvalidformat",
			password:    "logoutpassword123",
			authHeader:  basicPrefix,
			useToken:    true,
			description: "should return unauthorized when using Basic instead of Bearer",
		},
		{
			name:        "missing_bearer_prefix",
			setupUser:   true,
			email:       "logout_missing_bearer@test.com",
			username:    "logoutmissingbearer",
			password:    "logoutpassword123",
			authHeader:  "",
			useToken:    true,
			description: "should return unauthorized when Bearer prefix is missing",
		},
		{
			name:        "invalid_malformed_token",
			setupUser:   true,
			email:       "logout_invalid_token@test.com",
			username:    "logoutinvalidtoken",
			password:    "logoutpassword123",
			authHeader:  bearerPrefix,
			tokenValue:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			description: "should return unauthorized when token is malformed",
		},
		{
			name:        "expired_token",
			setupUser:   false,
			authHeader:  bearerPrefix,
			tokenValue:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalidSignature",
			description: "should return unauthorized when token is expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginResponse *dto.LoginWithPasswordResponse

			if tt.setupUser {
				_, loginResponse = createAndLoginUser(t, tt.email, tt.username, tt.password)
			}

			req, err := http.NewRequest("POST", testServer.URL+apiEndpoints.Logout, nil)
			assert.NoError(t, err, "Failed to create logout request")

			if tt.authHeader != "" {
				var token string
				if tt.tokenValue != "" {
					token = tt.tokenValue
				} else if tt.useToken && loginResponse != nil {
					token = loginResponse.Data.AccessToken
				}
				req.Header.Set(headerAuthorization, tt.authHeader+token)
			} else if tt.useToken && loginResponse != nil {
				req.Header.Set(headerAuthorization, loginResponse.Data.AccessToken)
			}

			resp, err := httpClient.Do(req)
			assert.NoError(t, err, "Failed to send logout request")
			defer resp.Body.Close()

			expectErrorResponse(t, resp, http.StatusUnauthorized)
		})
	}
}
