package internal_string

import (
	"slices"
	"strings"

	internal_errors "heimdall/backend/pkg/errors"

	"github.com/microcosm-cc/bluemonday"
)

// StrictPolicy removes all HTML tags and attributes - use for names, titles, etc.
var StrictPolicy = bluemonday.StrictPolicy()

// UGCPolicy allows safe user-generated content with basic formatting - use for descriptions, comments
var UGCPolicy = bluemonday.UGCPolicy()

// SanitizeStrict removes all HTML tags and normalizes whitespace
// Use this for fields like names, usernames, emails where no HTML should be allowed
func SanitizeStrict(input string) string {
	if input == "" {
		return input
	}

	// Remove all HTML tags
	sanitized := StrictPolicy.Sanitize(input)

	// Normalize whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeUGC allows safe HTML tags for user-generated content
// Use this for fields like descriptions, comments where basic formatting is allowed
func SanitizeUGC(input string) string {
	if input == "" {
		return input
	}

	// Allow safe HTML tags
	sanitized := UGCPolicy.Sanitize(input)

	// Normalize whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ContainsSuspiciousContent checks if input contains potentially dangerous patterns
func ContainsSuspiciousContent(input string) bool {
	suspiciousPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"on\x20*=", // matches onclick, onload, etc.
		"eval(",
		"expression(",
		"vbscript:",
		"data:text/html",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}

	return false
}

// ContainsSQLInjection checks if input contains potential SQL injection patterns
func ContainsSQLInjection(input string) bool {
	// Normalize input for checking
	inputLower := strings.ToLower(strings.ReplaceAll(input, " ", ""))

	// SQL injection patterns - be specific to avoid false positives
	sqlPatterns := []string{
		// SQL injection with quotes and keywords
		"'or",
		"'and",
		"'union",
		"'select",
		"'insert",
		"'update",
		"'delete",
		"'drop",
		"'create",
		"'alter",
		"'exec",
		"'execute",

		// Common injection patterns
		"1=1",
		"1='1'",
		"'='",
		"--",
		"/*",
		"*/",
		";;",
		"admin'--",

		// SQL functions that shouldn't be in user input
		"concat(",
		"substring(",
		"ascii(",
		"char(",
		"cast(",
		"convert(",
		"@@",
		"information_schema",
		"sys.tables",
		"pg_tables",
		"sqlite_master",
	}

	// Check for dangerous patterns
	for _, pattern := range sqlPatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}

	// Special check for boolean injection patterns
	booleanPatterns := []string{
		"' or 1=1",
		"' or '1'='1",
		"' or 'a'='a",
		"' or true",
		"' and 1=1",
		"' and '1'='1",
	}

	originalLower := strings.ToLower(input)
	for _, pattern := range booleanPatterns {
		if strings.Contains(originalLower, pattern) {
			return true
		}
	}

	return false
}

// IsSQLKeyword checks if the input is a dangerous SQL keyword
func IsSQLKeyword(input string) bool {
	dangerousKeywords := []string{
		"select", "insert", "update", "delete", "drop", "create", "alter",
		"exec", "execute", "union", "declare", "cast", "convert",
		"information_schema", "sysobjects", "syscolumns", "pg_tables",
		"sqlite_master", "mysql", "postgres", "mssql",
	}

	inputLower := strings.ToLower(strings.TrimSpace(input))

	return slices.Contains(dangerousKeywords, inputLower)
}

// ValidateUserInput performs comprehensive input validation
func ValidateUserInput(input string, fieldName string) error {
	if input == "" {
		return nil // Empty input is valid
	}

	// Check for XSS patterns
	if ContainsSuspiciousContent(input) {
		return internal_errors.NewValidationError("Suspicious XSS content detected", map[string]any{
			"field":   fieldName,
			"pattern": "XSS",
			"input":   input,
		})
	}

	// Check for SQL injection patterns
	if ContainsSQLInjection(input) {
		return internal_errors.NewValidationError("Potential SQL injection detected", map[string]any{
			"field":   fieldName,
			"pattern": "SQL_INJECTION",
			"input":   input,
		})
	}

	// Check if entire input is a dangerous SQL keyword
	if IsSQLKeyword(input) {
		return internal_errors.NewValidationError("Dangerous SQL keyword not allowed", map[string]any{
			"field":   fieldName,
			"pattern": "SQL_KEYWORD",
			"input":   input,
		})
	}

	return nil
}
