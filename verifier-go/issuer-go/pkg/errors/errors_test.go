package errors

import (
	"net/http"
	"testing"
)

// TestNewVCError tests VCError creation
func TestNewVCError(t *testing.T) {
	code := ErrCredInvalidCredentialID
	message := "test error message"

	err := NewVCError(code, message)

	if err.Code != code {
		t.Errorf("Expected code %d, got %d", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}
}

// TestVCError_Error tests error string formatting
func TestVCError_Error(t *testing.T) {
	err := NewVCError(61006, "invalid credential ID")
	expected := "[61006] invalid credential ID"

	if err.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, err.Error())
	}
}

// TestVCError_HTTPStatus tests HTTP status code mapping
func TestVCError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Bad request - invalid CID", ErrCredInvalidCredentialID, http.StatusBadRequest},
		{"Bad request - invalid type", ErrCredInvalidCredentialType, http.StatusBadRequest},
		{"Not found - credential", ErrCredCredentialNotFound, http.StatusNotFound},
		{"Not found - schema", ErrInfoSchemaNotFound, http.StatusNotFound},
		{"Internal server error", ErrCredGenerateVCError, http.StatusInternalServerError},
		{"Unknown error", Unknown, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewVCError(tt.code, "test")
			status := err.HTTPStatus()

			if status != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, status)
			}
		})
	}
}

// TestVCError_Response tests response map generation
func TestVCError_Response(t *testing.T) {
	code := 61002
	message := "generation failed"
	err := NewVCError(code, message)

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
		{"ErrCredInvalidCredentialGenerationRequest", ErrCredInvalidCredentialGenerationRequest, 61001},
		{"ErrCredGenerateVCError", ErrCredGenerateVCError, 61002},
		{"ErrCredInvalidCredentialID", ErrCredInvalidCredentialID, 61006},
		{"ErrCredCredentialNotFound", ErrCredCredentialNotFound, 61010},
		{"ErrCredInvalidNonce", ErrCredInvalidNonce, 61012},
		{"ErrSLGenerateStatusListError", ErrSLGenerateStatusListError, 62001},
		{"ErrDIDFrontendGenerateDIDError", ErrDIDFrontendGenerateDIDError, 63001},
		{"ErrInfoInvalidCredentialType", ErrInfoInvalidCredentialType, 64001},
		{"ErrDBQueryError", ErrDBQueryError, 68001},
		{"ErrSysGenerateKeyError", ErrSysGenerateKeyError, 69001},
		{"ErrSysNotRegisterDIDYetError", ErrSysNotRegisterDIDYetError, 69004},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected %s to be %d, got %d", tt.name, tt.expected, tt.code)
			}
		})
	}
}
