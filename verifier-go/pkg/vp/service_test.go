package vp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/errors"
)

// TestValidate_NullPresentationList tests validation with nil presentation list
func TestValidate_NullPresentationList(t *testing.T) {
	// Given
	service := NewService()
	ctx := context.Background()
	var presentations []string = nil

	// When
	result, status, err := service.Validate(ctx, presentations)

	// Then
	if err == nil {
		t.Error("Expected error for nil presentation list")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrPresInvalidPresentationValidationRequest {
		t.Errorf("Expected error code %d, got %d", errors.ErrPresInvalidPresentationValidationRequest, vpErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	// Verify response contains error code
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(float64); !ok || int(code) != errors.ErrPresInvalidPresentationValidationRequest {
		t.Errorf("Response does not contain expected error code")
	}

	if msg, ok := response["message"].(string); !ok || msg == "" {
		t.Errorf("Response does not contain error message")
	}
}

// TestValidate_EmptyPresentationList tests validation with empty presentation list
func TestValidate_EmptyPresentationList(t *testing.T) {
	// Given
	service := NewService()
	ctx := context.Background()
	presentations := []string{}

	// When
	result, status, err := service.Validate(ctx, presentations)

	// Then
	if err == nil {
		t.Error("Expected error for empty presentation list")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrPresInvalidPresentationValidationRequest {
		t.Errorf("Expected error code %d, got %d", errors.ErrPresInvalidPresentationValidationRequest, vpErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	// Verify response format
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if _, ok := response["code"]; !ok {
		t.Error("Response missing 'code' field")
	}

	if _, ok := response["message"]; !ok {
		t.Error("Response missing 'message' field")
	}
}

// TestValidate_BlankPresentationEntries tests validation with blank entries
func TestValidate_BlankPresentationEntries(t *testing.T) {
	// Given
	service := NewService()
	ctx := context.Background()
	presentations := []string{"", "   ", ""}

	// When
	result, status, err := service.Validate(ctx, presentations)

	// Then - blank entries should be skipped, resulting in empty results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Response should be an empty array (all blank entries skipped)
	var response []interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected empty response array, got length %d", len(response))
	}
}

// NOTE: Tests with real JWT validation are in service_integration_test.go
// These placeholder tests have been removed because they used fake JWT strings.
// Use TestValidate_WithRealJWT and related integration tests instead.

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Error("NewService returned nil")
	}
}

// TestGetVPPath tests VP path generation
func TestGetVPPath(t *testing.T) {
	tests := []struct {
		name     string
		vpIndex  int
		isArray  bool
		expected string
	}{
		{"Single VP", 0, false, "$"},
		{"First VP in array", 0, true, "$[0]"},
		{"Second VP in array", 1, true, "$[1]"},
		{"Third VP in array", 2, true, "$[2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getVPPath(tt.vpIndex, tt.isArray)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestGetVCPath tests VC path generation
func TestGetVCPath(t *testing.T) {
	tests := []struct {
		name     string
		vcIndex  int
		expected string
	}{
		{"First VC", 0, "$.vp.verifiableCredential[0]"},
		{"Second VC", 1, "$.vp.verifiableCredential[1]"},
		{"Third VC", 2, "$.vp.verifiableCredential[2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getVCPath(tt.vcIndex)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestValidate_TooManyPresentations tests validation with too many presentations (DoS protection)
func TestValidate_TooManyPresentations(t *testing.T) {
	// Given
	service := NewService()
	ctx := context.Background()

	// Create more than MaxPresentations (101 presentations)
	presentations := make([]string, MaxPresentations+1)
	for i := range presentations {
		presentations[i] = "eyJhbGciOiJFUzI1NiJ9.test.sig"
	}

	// When
	_, status, err := service.Validate(ctx, presentations)

	// Then
	if err == nil {
		t.Error("Expected error for too many presentations")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrPresInvalidPresentationValidationRequest {
		t.Errorf("Expected error code %d, got %d", errors.ErrPresInvalidPresentationValidationRequest, vpErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestValidate_PresentationTooLarge tests validation with presentation exceeding size limit
func TestValidate_PresentationTooLarge(t *testing.T) {
	// Given
	service := NewService()
	ctx := context.Background()

	// Create a presentation larger than MaxPresentationSize (1MB + 1 byte)
	largePresentation := make([]byte, MaxPresentationSize+1)
	for i := range largePresentation {
		largePresentation[i] = 'A'
	}

	presentations := []string{string(largePresentation)}

	// When
	_, status, err := service.Validate(ctx, presentations)

	// Then
	if err == nil {
		t.Error("Expected error for oversized presentation")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrPresInvalidPresentationValidationRequest {
		t.Errorf("Expected error code %d, got %d", errors.ErrPresInvalidPresentationValidationRequest, vpErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestValidate_TotalPayloadTooLarge tests validation with total payload exceeding limit
func TestValidate_TotalPayloadTooLarge(t *testing.T) {
	// Given
	service := NewService()
	ctx := context.Background()

	// Create presentations that together exceed MaxTotalPayloadSize
	// Each presentation is 5MB, create 3 of them (15MB total, exceeds 10MB limit)
	singleSize := 5 * 1024 * 1024
	largePresentation := make([]byte, singleSize)
	for i := range largePresentation {
		largePresentation[i] = 'B'
	}

	presentations := []string{
		string(largePresentation),
		string(largePresentation),
		string(largePresentation),
	}

	// When
	_, status, err := service.Validate(ctx, presentations)

	// Then
	if err == nil {
		t.Error("Expected error for oversized total payload")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrPresInvalidPresentationValidationRequest {
		t.Errorf("Expected error code %d, got %d", errors.ErrPresInvalidPresentationValidationRequest, vpErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}
