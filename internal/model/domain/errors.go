package domain

import (
	"fmt"
)

// Domain errors for business logic
type DomainError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error constructors
func ErrNotFound(entity string) error {
	return DomainError{
		Type:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", entity),
		Code:    404,
	}
}

func ErrInvalidInput(message string) error {
	return DomainError{
		Type:    "INVALID_INPUT",
		Message: message,
		Code:    400,
	}
}

func ErrConflict(message string) error {
	return DomainError{
		Type:    "CONFLICT",
		Message: message,
		Code:    409,
	}
}

func ErrUnauthorized(message string) error {
	return DomainError{
		Type:    "UNAUTHORIZED",
		Message: message,
		Code:    401,
	}
}

func ErrForbidden(message string) error {
	return DomainError{
		Type:    "FORBIDDEN",
		Message: message,
		Code:    403,
	}
}

func ErrVersionConflict(entity string, expectedVersion, actualVersion int64) error {
	return DomainError{
		Type:    "VERSION_CONFLICT",
		Message: fmt.Sprintf("%s has been modified (expected version %d, actual version %d)", entity, expectedVersion, actualVersion),
		Code:    409,
	}
}

func ErrPermissionDenied(message string) error {
	return DomainError{
		Type:    "PERMISSION_DENIED",
		Message: message,
		Code:    403,
	}
}

func ErrBusinessRule(message string) error {
	return DomainError{
		Type:    "BUSINESS_RULE_VIOLATION",
		Message: message,
		Code:    422,
	}
}

// Error type checkers
func IsNotFoundError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Type == "NOT_FOUND"
	}
	return false
}

func IsVersionConflictError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Type == "VERSION_CONFLICT"
	}
	return false
}

func IsBusinessRuleError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Type == "BUSINESS_RULE_VIOLATION"
	}
	return false
}

func IsConflictError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Type == "CONFLICT"
	}
	return false
}

func IsInvalidInputError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Type == "INVALID_INPUT"
	}
	return false
}