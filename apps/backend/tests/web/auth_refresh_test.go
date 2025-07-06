package web_test

import (
	"net/http"
	"testing"

	"heimdall/backend/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// REFRESH API TESTS
// ==========================================

// Success Tests

func TestRefreshSuccess(t *testing.T) {
	email := "refresh_success@test.com"
	username := "refreshsuccess"
	password := "refreshpassword123"

	// Create and login user
	_, loginResponse := createAndLoginUser(t, email, username, password)

	// Test refresh using correct auth pattern
	resp := postRefreshRequest(t, loginResponse.Data.RefreshToken, loginResponse.Data.AccessToken)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var refreshResponse map[string]any
	err := parseJSONResponse(resp.Body, &refreshResponse)
	assert.NoError(t, err, "Failed to decode refresh response")

	assert.Equal(t, true, refreshResponse["success"], "Expected success=true")

	// Check that we have new tokens
	data, ok := refreshResponse["data"].(map[string]any)
	assert.True(t, ok, "Expected data object")
	assert.NotEmpty(t, data["accessToken"], "Expected new access token")
	assert.NotEmpty(t, data["refreshToken"], "Expected new refresh token")
}

// Error Tests

func TestRefreshTokenErrors(t *testing.T) {
	tests := []struct {
		name         string
		setupUser    bool
		email        string
		username     string
		password     string
		refreshToken string
		accessToken  string
		description  string
	}{
		{
			name:         "missing_refresh_token",
			setupUser:    true,
			email:        "refresh_no_refresh@test.com",
			username:     "refreshnorefresh",
			password:     "refreshpassword123",
			refreshToken: "",
			accessToken:  "use_real",
			description:  "should return unauthorized when refresh token header is missing",
		},
		{
			name:         "missing_access_token",
			setupUser:    true,
			email:        "refresh_no_access@test.com",
			username:     "refreshnoaccess",
			password:     "refreshpassword123",
			refreshToken: "use_real",
			accessToken:  "",
			description:  "should return unauthorized when access token header is missing",
		},
		{
			name:         "invalid_refresh_token",
			setupUser:    true,
			email:        "refresh_invalid_refresh@test.com",
			username:     "refreshinvalidrefresh",
			password:     "refreshpassword123",
			refreshToken: "invalid-refresh-token",
			accessToken:  "use_real",
			description:  "should return unauthorized when refresh token is invalid",
		},
		{
			name:         "invalid_access_token",
			setupUser:    true,
			email:        "refresh_invalid_access@test.com",
			username:     "refreshinvalidaccess",
			password:     "refreshpassword123",
			refreshToken: "use_real",
			accessToken:  "invalid-access-token",
			description:  "should return unauthorized when access token is invalid",
		},
		{
			name:         "expired_refresh_token",
			setupUser:    false,
			refreshToken: "expired.refresh.token",
			accessToken:  "someAccessToken",
			description:  "should return unauthorized when refresh token is expired",
		},
		{
			name:         "expired_access_token",
			setupUser:    false,
			refreshToken: "someRefreshToken",
			accessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalidSignature",
			description:  "should return unauthorized when access token is expired",
		},
		{
			name:         "malformed_access_token",
			setupUser:    false,
			refreshToken: "someRefreshToken",
			accessToken:  "not.a.jwt.token.at.all",
			description:  "should return unauthorized when access token is malformed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginResponse *dto.LoginWithPasswordResponse

			if tt.setupUser {
				_, loginResponse = createAndLoginUser(t, tt.email, tt.username, tt.password)
			}

			refreshToken := tt.refreshToken
			accessToken := tt.accessToken

			if refreshToken == "use_real" && loginResponse != nil {
				refreshToken = loginResponse.Data.RefreshToken
			}
			if accessToken == "use_real" && loginResponse != nil {
				accessToken = loginResponse.Data.AccessToken
			}

			resp := postRefreshRequest(t, refreshToken, accessToken)
			defer resp.Body.Close()

			expectErrorResponse(t, resp, http.StatusUnauthorized)
		})
	}
}
