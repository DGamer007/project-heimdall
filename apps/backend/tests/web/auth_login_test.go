package web_test

import (
	"net/http"
	"testing"

	"heimdall/backend/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// LOGIN API TESTS
// ==========================================

// Success Tests

func TestLoginWithPasswordSuccessForEmailIdentifier(t *testing.T) {
	user := createTestUser(t, "login_email@test.com", "loginemail", "loginpassword123")

	// Test login with email
	loginResponse := loginUser(t, user.Email, user.Password)

	assert.True(t, loginResponse.Success, "Expected success=true")
	assert.Equal(t, "Successfully logged in", loginResponse.Message, "Expected login success message")
	assert.NotEmpty(t, loginResponse.Data.AccessToken, "Expected non-empty access token")
	assert.NotEmpty(t, loginResponse.Data.RefreshToken, "Expected non-empty refresh token")
}

func TestLoginWithPasswordSuccessForUserNameIdentifier(t *testing.T) {
	user := createTestUser(t, "login_username@test.com", "loginusername", "loginpassword123")

	// Test login with username
	loginResponse := loginUser(t, user.Username, user.Password)

	assert.True(t, loginResponse.Success, "Expected success=true")
	assert.Equal(t, "Successfully logged in", loginResponse.Message, "Expected login success message")
	assert.NotEmpty(t, loginResponse.Data.AccessToken, "Expected non-empty access token")
	assert.NotEmpty(t, loginResponse.Data.RefreshToken, "Expected non-empty refresh token")
}

// Validation Error Tests

func TestLoginValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "missing_identifier",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Login, loginBaseFields, "identifier")
			},
			description: "should return validation error when identifier is missing",
		},
		{
			name: "missing_password",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Login, loginBaseFields, "password")
			},
			description: "should return validation error when password is missing",
		},
		{
			name: "short_password",
			testFunc: func(t *testing.T) {
				payload := dto.LoginWithPasswordPayload{
					Identifier: "user1@test.com",
					Password:   shortPassword,
				}
				testValidationError(t, apiEndpoints.Login, payload)
			},
			description: "should return validation error when password is too short",
		},
		{
			name: "empty_fields",
			testFunc: func(t *testing.T) {
				payload := dto.LoginWithPasswordPayload{
					Identifier: "",
					Password:   "",
				}
				testValidationError(t, apiEndpoints.Login, payload)
			},
			description: "should return validation error when fields are empty",
		},
		{
			name: "malformed_json",
			testFunc: func(t *testing.T) {
				malformedJSON := `{"identifier": "user1@test.com", "password"`
				testMalformedJSON(t, apiEndpoints.Login, malformedJSON)
			},
			description: "should return validation error when JSON is malformed",
		},
		{
			name: "empty_body",
			testFunc: func(t *testing.T) {
				testEmptyBody(t, apiEndpoints.Login)
			},
			description: "should return validation error when body is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// Authentication Error Tests

func TestLoginAuthenticationErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupUser   bool
		email       string
		username    string
		password    string
		identifier  string
		loginPass   string
		description string
	}{
		{
			name:        "invalid_email",
			setupUser:   false,
			identifier:  "nonexistent@test.com",
			loginPass:   "anypassword123",
			description: "should return unauthorized when email doesn't exist",
		},
		{
			name:        "invalid_username",
			setupUser:   false,
			identifier:  "nonexistentuser",
			loginPass:   "anypassword123",
			description: "should return unauthorized when username doesn't exist",
		},
		{
			name:        "wrong_password_for_email",
			setupUser:   true,
			email:       "wrong_password_email@test.com",
			username:    "wrongpasswordemail",
			password:    "correctpassword123",
			identifier:  "wrong_password_email@test.com",
			loginPass:   "wrongpassword123",
			description: "should return unauthorized when password is wrong for email",
		},
		{
			name:        "wrong_password_for_username",
			setupUser:   true,
			email:       "wrong_password_username@test.com",
			username:    "wrongpassworduser",
			password:    "correctpassword123",
			identifier:  "wrongpassworduser",
			loginPass:   "wrongpassword123",
			description: "should return unauthorized when password is wrong for username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupUser {
				createTestUser(t, tt.email, tt.username, tt.password)
			}

			payload := dto.LoginWithPasswordPayload{
				Identifier: tt.identifier,
				Password:   tt.loginPass,
			}

			resp := postJSON(t, apiEndpoints.Login, payload)
			defer resp.Body.Close()

			expectErrorResponse(t, resp, http.StatusUnauthorized)
		})
	}
}
