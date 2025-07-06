package web_test

import (
	"net/http"
	"testing"

	"heimdall/backend/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// GENERAL HTTP ERROR TESTS
// ==========================================

// Wrong HTTP Method Tests

func TestWrongHTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		correctMethod  string
		wrongMethods   []string
		expectedStatus int
		description    string
	}{
		{
			name:           "register_wrong_methods",
			endpoint:       apiEndpoints.Register,
			correctMethod:  "POST",
			wrongMethods:   []string{"GET", "PUT", "DELETE", "PATCH"},
			expectedStatus: http.StatusNotFound,
			description:    "should return 404 when using wrong HTTP methods for register (only POST allowed)",
		},
		{
			name:           "login_wrong_methods",
			endpoint:       apiEndpoints.Login,
			correctMethod:  "POST",
			wrongMethods:   []string{"GET", "PUT", "DELETE", "PATCH"},
			expectedStatus: http.StatusNotFound,
			description:    "should return 404 when using wrong HTTP methods for login (only POST allowed)",
		},
		{
			name:           "logout_wrong_methods",
			endpoint:       apiEndpoints.Logout,
			correctMethod:  "POST",
			wrongMethods:   []string{"GET", "PUT", "DELETE", "PATCH"},
			expectedStatus: http.StatusNotFound,
			description:    "should return 404 when using wrong HTTP methods for logout (only POST allowed)",
		},
		{
			name:           "refresh_wrong_methods",
			endpoint:       apiEndpoints.Refresh,
			correctMethod:  "POST",
			wrongMethods:   []string{"GET", "PUT", "DELETE", "PATCH"},
			expectedStatus: http.StatusNotFound,
			description:    "should return 404 when using wrong HTTP methods for refresh (only POST allowed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, method := range tt.wrongMethods {
				t.Run(method, func(t *testing.T) {
					expectSpecificWrongMethod(t, tt.endpoint, method, tt.expectedStatus)
				})
			}
		})
	}
}

// Wrong Content-Type Tests

func TestWrongContentType(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		payload     any
		description string
	}{
		{
			name:     "register_wrong_content_type",
			endpoint: apiEndpoints.Register,
			payload: dto.RegisterWithPasswordPayload{
				Email:     "content@test.com",
				UserName:  "contentuser",
				Password:  defaultTestPassword,
				FirstName: "Content",
				LastName:  "Test",
			},
			description: "should return error when using wrong content type for register",
		},
		{
			name:     "login_wrong_content_type",
			endpoint: apiEndpoints.Login,
			payload: dto.LoginWithPasswordPayload{
				Identifier: "wrongcontenttype@test.com",
				Password:   "wrongcontenttypepassword",
			},
			description: "should return error when using wrong content type for login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWrongContentType(t, tt.endpoint, tt.payload)
		})
	}
}

// ==========================================
// EDGE CASES AND BOUNDARY TESTS
// ==========================================

func TestLoginWithSpecialCharacters(t *testing.T) {
	payload := dto.LoginWithPasswordPayload{
		Identifier: "user@test.com!@#$%^&*()",
		Password:   "password!@#$%^&*()",
	}

	resp := postJSON(t, apiEndpoints.Login, payload)
	defer resp.Body.Close()

	expectErrorResponse(t, resp, http.StatusUnauthorized)
}

func TestRegisterWithSQLInjectionAttempt(t *testing.T) {
	email := "sqlinjection@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "sql'; DROP TABLE users; --", // SQL injection in username
		Password:  defaultTestPassword,
		FirstName: "SQL'; DROP TABLE users; --", // SQL injection in first name
		LastName:  "Injection",
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// Should reject SQL injection attempts with validation error due to comprehensive input validation
	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func TestLoginWithSQLInjectionAttempt(t *testing.T) {
	payload := dto.LoginWithPasswordPayload{
		Identifier: "admin' OR '1'='1",
		Password:   "anything",
	}

	resp := postJSON(t, apiEndpoints.Login, payload)
	defer resp.Body.Close()

	// Should reject SQL injection attempts with validation error due to comprehensive input validation
	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func TestRegisterWithSQLKeywords(t *testing.T) {
	email := "sqlkeywords@example.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "DROP", // Dangerous SQL keyword
		Password:  defaultTestPassword,
		FirstName: "SELECT", // Dangerous SQL keyword
		LastName:  "DELETE", // Dangerous SQL keyword
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// Should reject dangerous SQL keywords
	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func TestRegisterWithAdvancedSQLInjection(t *testing.T) {
	email := "advancedinjection@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	testCases := []struct {
		name      string
		firstName string
		lastName  string
		username  string
	}{
		{
			name:      "Union injection",
			firstName: "John' UNION SELECT * FROM users --",
			lastName:  "Doe",
			username:  "unionuser",
		},
		{
			name:      "Comment injection",
			firstName: "John",
			lastName:  "Doe'/*comment*/",
			username:  "commentuser",
		},
		{
			name:      "Function injection",
			firstName: "concat('admin',password)",
			lastName:  "Doe",
			username:  "funcuser",
		},
		{
			name:      "Boolean injection",
			firstName: "John",
			lastName:  "Doe",
			username:  "user' OR 1=1 --",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := dto.RegisterWithPasswordPayload{
				Email:     email,
				UserName:  tc.username,
				Password:  defaultTestPassword,
				FirstName: tc.firstName,
				LastName:  tc.lastName,
			}

			resp := postJSON(t, apiEndpoints.Register, payload)
			defer resp.Body.Close()

			// All SQL injection attempts should be rejected
			expectErrorResponse(t, resp, http.StatusBadRequest)
		})
	}
}

func TestRegisterWithLegitimateContent(t *testing.T) {
	email := "legitimate@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "legitimate_user_123",
		Password:  defaultTestPassword,
		FirstName: "John",     // Normal name
		LastName:  "O'Connor", // Name with apostrophe (common in Irish names)
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// Should succeed - legitimate content should pass validation
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Legitimate content should be allowed")
}

func TestRegisterWithXSSAttempt(t *testing.T) {
	email := "xss@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "xssuser",
		Password:  defaultTestPassword,
		FirstName: "<script>alert('xss')</script>",
		LastName:  "<img src=x onerror=alert('xss')>",
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// Should reject suspicious XSS content with validation error due to bluemonday sanitization
	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func TestLoginWithXSSAttempt(t *testing.T) {
	payload := dto.LoginWithPasswordPayload{
		Identifier: "<script>alert('xss')</script>admin@test.com",
		Password:   "password",
	}

	resp := postJSON(t, apiEndpoints.Login, payload)
	defer resp.Body.Close()

	// Should reject suspicious XSS content with validation error due to bluemonday sanitization
	expectErrorResponse(t, resp, http.StatusBadRequest)
}

func TestRegisterWithSafeHTMLContent(t *testing.T) {
	email := "safe@test.com"

	// Cleanup after test
	t.Cleanup(func() {
		cleanupUser(t, email)
	})

	payload := dto.RegisterWithPasswordPayload{
		Email:     email,
		UserName:  "safeuser",
		Password:  defaultTestPassword,
		FirstName: "John <strong>", // Safe HTML will be stripped
		LastName:  "Doe & Co.",     // HTML entities will be preserved as text
	}

	resp := postJSON(t, apiEndpoints.Register, payload)
	defer resp.Body.Close()

	// Should succeed - safe HTML content is cleaned but allowed
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Safe HTML content should be allowed after sanitization")
}
