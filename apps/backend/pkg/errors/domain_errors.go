package internal_errors

import "fmt"

type DomainError struct {
	Type    string
	Message string
	Details map[string]any
}

func (e DomainError) Error() string {
	return e.Message
}

type BusinessLogicError struct {
	DomainError
}

func NewBusinessLogicError(message string, details map[string]any) *BusinessLogicError {
	return &BusinessLogicError{
		DomainError: DomainError{
			Type:    "BUSINESS_LOGIC_ERROR",
			Message: message,
			Details: details,
		},
	}
}

type ValidationError struct {
	DomainError
}

func NewValidationError(message string, details map[string]any) *ValidationError {
	return &ValidationError{
		DomainError: DomainError{
			Type:    "VALIDATION_ERROR",
			Message: message,
			Details: details,
		},
	}
}

type NotFoundError struct {
	DomainError
}

func NewNotFoundError(resource string, identifier string) *NotFoundError {
	return &NotFoundError{
		DomainError: DomainError{
			Type:    "NOT_FOUND_ERROR",
			Message: fmt.Sprintf("%s not found", resource),
			Details: map[string]any{
				"resource":   resource,
				"identifier": identifier,
			},
		},
	}
}

type DuplicateResourceError struct {
	DomainError
}

func NewDuplicateResourceError(resource string, field string, value string) *DuplicateResourceError {
	return &DuplicateResourceError{
		DomainError: DomainError{
			Type:    "DUPLICATE_RESOURCE_ERROR",
			Message: fmt.Sprintf("%s with %s '%s' already exists", resource, field, value),
			Details: map[string]any{
				"resource": resource,
				"field":    field,
				"value":    value,
			},
		},
	}
}

type AuthenticationError struct {
	DomainError
}

func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		DomainError: DomainError{
			Type:    "AUTHENTICATION_ERROR",
			Message: message,
			Details: nil,
		},
	}
}

type InfrastructureError struct {
	DomainError
	OriginalError error
}

func NewInfrastructureError(message string, originalError error) *InfrastructureError {
	return &InfrastructureError{
		DomainError: DomainError{
			Type:    "INFRASTRUCTURE_ERROR",
			Message: message,
			Details: map[string]any{
				"original_error": originalError.Error(),
			},
		},
		OriginalError: originalError,
	}
}
