package errors

import (
	"net/http"
	"testing"
)

// TestNewVPError tests VPError creation
func TestNewVPError(t *testing.T) {
	code := ErrPresInvalidPresentationValidationRequest
	message := "test error message"

	err := NewVPError(code, message)

	if err.Code != code {
		t.Errorf("Expected code %d, got %d", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}
}

// TestVPError_Error tests error string formatting
func TestVPError_Error(t *testing.T) {
	err := NewVPError(71001, "invalid request")
	expected := "[71001] invalid request"

	if err.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, err.Error())
	}
}

// TestVPError_HTTPStatus tests HTTP status code mapping
func TestVPError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Bad request error", ErrPresInvalidPresentationValidationRequest, http.StatusBadRequest},
		{"Unknown error", Unknown, http.StatusInternalServerError},
		{"VP validation error", ErrPresValidateVPError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewVPError(tt.code, "test")
			status := err.HTTPStatus()

			if status != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, status)
			}
		})
	}
}

// TestVPError_Response tests response map generation
func TestVPError_Response(t *testing.T) {
	code := 71002
	message := "validation failed"
	err := NewVPError(code, message)

	response := err.Response()

	if response["code"] != code {
		t.Errorf("Expected response code %d, got %v", code, response["code"])
	}

	if response["message"] != message {
		t.Errorf("Expected response message %s, got %v", message, response["message"])
	}
}

// TestErrorConstants tests that error constants have expected values
func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Unknown", Unknown, 99999},
		{"ErrIllegalArgument", ErrIllegalArgument, 70001},
		{"ErrPresInvalidPresentationValidationRequest", ErrPresInvalidPresentationValidationRequest, 71001},
		{"ErrPresValidateVPError", ErrPresValidateVPError, 71002},
		{"ErrPresValidateVPContentError", ErrPresValidateVPContentError, 71003},
		{"ErrCredValidateVCContentError", ErrCredValidateVCContentError, 72001},
		{"ErrDBQueryError", ErrDBQueryError, 78001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected %s to be %d, got %d", tt.name, tt.expected, tt.code)
			}
		})
	}
}
