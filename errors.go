package pasis

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Pasis API.
type APIError struct {
	StatusCode int
	Message    string
	Errors     []string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if len(e.Errors) > 0 {
		return fmt.Sprintf("API error: %v", e.Errors)
	}
	return fmt.Sprintf("API error: status code %d", e.StatusCode)
}

// AuthError represents an authentication error.
type AuthError struct {
	Message string
	Err     error
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("authentication error: %s", e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("authentication error: %v", e.Err)
	}
	return "authentication error"
}

// Unwrap returns the underlying error.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error.
type ValidationError struct {
	Message string
	Field   string
	Err     error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("validation error: %s", e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("validation error: %v", e.Err)
	}
	return "validation error"
}

// Unwrap returns the underlying error.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// parseErrorResponse parses the error response from the API.
func parseErrorResponse(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    errResp.Message,
		Errors:     errResp.Errors,
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthError{
			Message: errResp.Message,
			Err:     apiErr,
		}
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return &ValidationError{
			Message: errResp.Message,
			Err:     apiErr,
		}
	default:
		return apiErr
	}
}
