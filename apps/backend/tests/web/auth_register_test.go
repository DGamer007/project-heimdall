package web_test

import (
	"net/http"
	"strings"
	"testing"

	"heimdall/backend/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// REGISTER API TESTS
// ==========================================

// Success Tests

func TestRegisterWithPasswordSuccess(t *testing.T) {
	email := "register_success@test.com"
	username := "registersuccess"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  username,
		Password:  "user1password1",
		FirstName: "User",
		LastName:  "One",
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var registerResponse dto.RegisterWithPasswordResponse
	err := parseJSONResponse(resp.Body, &registerResponse)
	assert.NoError(t, err, "Failed to decode register response")

	assert.True(t, registerResponse.Success, "Expected success=true")
	assert.Equal(t, "Successfully registered new account", registerResponse.Message, "Expected success message")
	assert.NotEmpty(t, registerResponse.Data.AccessToken, "Expected non-empty access token")
	assert.NotEmpty(t, registerResponse.Data.RefreshToken, "Expected non-empty refresh token")
}

// Validation Error Tests

func TestRegisterValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
		description string
	}{
		{
			name: "invalid_email",
			testFunc: func(t *testing.T) {
				payload := dto.RegisterWithPasswordPayload{
					Email:     "invalid-email",
					UserName:  "testuser",
					Password:  defaultTestPassword,
					FirstName: defaultFirstName,
					LastName:  defaultLastName,
				}
				testValidationError(t, apiEndpoints.Register, payload)
			},
			description: "should return validation error when email format is invalid",
		},
		{
			name: "missing_email",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Register, registerBaseFields, "email")
			},
			description: "should return validation error when email is missing",
		},
		{
			name: "missing_username",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Register, registerBaseFields, "userName")
			},
			description: "should return validation error when username is missing",
		},
		{
			name: "missing_password",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Register, registerBaseFields, "password")
			},
			description: "should return validation error when password is missing",
		},
		{
			name: "missing_first_name",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Register, registerBaseFields, "firstName")
			},
			description: "should return validation error when first name is missing",
		},
		{
			name: "missing_last_name",
			testFunc: func(t *testing.T) {
				testMissingField(t, apiEndpoints.Register, registerBaseFields, "lastName")
			},
			description: "should return validation error when last name is missing",
		},
		{
			name: "short_password",
			testFunc: func(t *testing.T) {
				payload := dto.RegisterWithPasswordPayload{
					Email:     "test@example.com",
					UserName:  "testuser",
					Password:  shortPassword,
					FirstName: defaultFirstName,
					LastName:  defaultLastName,
				}
				testValidationError(t, apiEndpoints.Register, payload)
			},
			description: "should return validation error when password is too short",
		},
		{
			name: "empty_fields",
			testFunc: func(t *testing.T) {
				payload := dto.RegisterWithPasswordPayload{
					Email:     "",
					UserName:  "",
					Password:  "",
					FirstName: "",
					LastName:  "",
				}
				testValidationError(t, apiEndpoints.Register, payload)
			},
			description: "should return validation error when all fields are empty",
		},
		{
			name: "malformed_json",
			testFunc: func(t *testing.T) {
				malformedJSON := `{"email": "test@example.com", "userName": "testuser", "password"`
				testMalformedJSON(t, apiEndpoints.Register, malformedJSON)
			},
			description: "should return validation error when JSON is malformed",
		},
		{
			name: "empty_body",
			testFunc: func(t *testing.T) {
				testEmptyBody(t, apiEndpoints.Register)
			},
			description: "should return validation error when request body is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// Conflict Tests

func TestRegisterWithDuplicateEmail(t *testing.T) {
	email := "duplicate_email@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	// First, register a user successfully
	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "uniqueuser1",
		Password:  defaultTestPassword,
		FirstName: "First",
		LastName:  "User",
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "First registration should succeed")

	// Now try to register with the same email
	payload.UserName = "uniqueuser2" // Different username
	payload.FirstName = "Second"

	resp = postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusConflict)
}

func TestRegisterWithDuplicateUserName(t *testing.T) {
	username := "duplicateuser"
	email1 := "unique1_username@test.com"
	email2 := "unique2_username@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email1)
		cleanupUser(t, email2)
	})

	// First, register a user successfully
	payload := dto.RegisterWithPasswordPayload{
		Email:     email1,
		UserName:  username,
		Password:  defaultTestPassword,
		FirstName: "First",
		LastName:  "User",
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "First registration should succeed")

	// Now try to register with the same username
	payload.Email = email2 // Different email
	payload.FirstName = "Second"

	resp = postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusConflict)
}

// Edge Cases

func TestRegisterWithLargePayload(t *testing.T) {
	email := "large@test.com"

	// Cleanup after test (in case registration succeeds)
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	// Create a very large string for testing
	largeString := strings.Repeat("a", largeStringSize)

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "largeuser",
		Password:  defaultTestPassword,
		FirstName: largeString,
		LastName:  largeString,
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// This might succeed or fail depending on server limits
	assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestEntityTooLarge,
		"Expected 201, 400, or 413, got %d", resp.StatusCode)
}

func TestRegisterWithUnicodeCharacters(t *testing.T) {
	email := "unicode@test.com"

	// Cleanup after test (in case registration succeeds)
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "unicodeuser",
		Password:  defaultTestPassword,
		FirstName: "你好世界",
		LastName:  "🚀🎯✅",
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// Should either succeed or fail gracefully
	assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest,
		"Expected 201 or 400, got %d", resp.StatusCode)
}
