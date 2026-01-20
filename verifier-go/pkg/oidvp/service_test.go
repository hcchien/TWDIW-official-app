package oidvp

import (
	"context"
	"testing"

	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/errors"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
)

// TestNewVerifierService tests service creation
func TestNewVerifierService(t *testing.T) {
	service := NewVerifierService("http://localhost:8080/verify")
	if service == nil {
		t.Error("NewVerifierService returned nil")
	}
	if service.vpVerifyURI != "http://localhost:8080/verify" {
		t.Errorf("Expected vpVerifyURI to be set, got %s", service.vpVerifyURI)
	}
}

// TestVerify_WalletError tests verification when wallet returns an error
func TestVerify_WalletError(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()
	authzResponse := &models.OIDVPAuthorizationResponse{
		Error:            "access_denied",
		ErrorDescription: "User cancelled the request",
	}

	// When
	result, err := service.Verify(ctx, authzResponse, "test-nonce", "test-client-id", "{}")

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.VerifyResult {
		t.Error("Expected VerifyResult to be false when wallet returns error")
	}

	if result.Error == nil {
		t.Error("Expected error info in result")
	}

	if result.Error.Message == "" {
		t.Error("Expected error message")
	}
}

// TestVerify_Success tests successful verification
func TestVerify_Success(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()
	authzResponse := &models.OIDVPAuthorizationResponse{
		VPToken:               "eyJhbGciOiJFUzI1NiJ9.test.signature",
		PresentationSubmission: `{"id":"test","definition_id":"test"}`,
	}

	// When
	result, err := service.Verify(ctx, authzResponse, "test-nonce", "test-client-id", "{}")

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !result.VerifyResult {
		t.Error("Expected VerifyResult to be true")
	}

	if result.HolderDID == "" {
		t.Error("Expected holder DID to be set")
	}
}

// TestVerifyPresentation_MissingRequiredParams tests validation with missing parameters
func TestVerifyPresentation_MissingRequiredParams(t *testing.T) {
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	tests := []struct {
		name   string
		nonce  string
		clientID string
		pd     string
	}{
		{"Missing nonce", "", "client-id", "{}"},
		{"Missing clientID", "nonce", "", "{}"},
		{"Missing PD", "nonce", "client-id", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result, err := service.verifyPresentation(ctx, "vp-token", "ps", tt.nonce, tt.clientID, tt.pd)

			// Then
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result.VerifyResult {
				t.Error("Expected VerifyResult to be false for missing params")
			}

			if result.Error == nil {
				t.Error("Expected error info in result")
			}
		})
	}
}

// TestGetVerifyResult_MissingBothParams tests retrieval with missing parameters
func TestGetVerifyResult_MissingBothParams(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	// When
	result, err := service.GetVerifyResult(ctx, "", "")

	// Then
	if err == nil {
		t.Error("Expected error when both transactionID and responseCode are empty")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrIllegalArgument {
		t.Errorf("Expected error code %d, got %d", errors.ErrIllegalArgument, vpErr.Code)
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}
}

// TestGetVerifyResult_Success tests successful result retrieval
func TestGetVerifyResult_Success(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	// When
	result, err := service.GetVerifyResult(ctx, "test-transaction-id", "test-response-code")

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	if !result.VerifyResult {
		t.Error("Expected successful verification result")
	}
}

// TestModifyPresentationDefinitionData_MissingParams tests with missing parameters
func TestModifyPresentationDefinitionData_MissingParams(t *testing.T) {
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	tests := []struct {
		name       string
		mode       string
		businessID string
		serialNo   string
	}{
		{"Missing mode", "", "business-id", "serial-no"},
		{"Missing businessID", "save", "", "serial-no"},
		{"Missing serialNo", "save", "business-id", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := service.ModifyPresentationDefinitionData(ctx, tt.mode, tt.businessID, tt.serialNo, nil)

			// Then
			if err == nil {
				t.Error("Expected error for missing parameters")
			}

			vpErr, ok := err.(*errors.VPError)
			if !ok {
				t.Errorf("Expected VPError, got %T", err)
			}

			if vpErr.Code != errors.ErrIllegalArgument {
				t.Errorf("Expected error code %d, got %d", errors.ErrIllegalArgument, vpErr.Code)
			}
		})
	}
}

// TestModifyPresentationDefinitionData_SaveWithoutPD tests save mode without PD
func TestModifyPresentationDefinitionData_SaveWithoutPD(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	// When
	err := service.ModifyPresentationDefinitionData(ctx, "save", "business-id", "serial-no", nil)

	// Then
	if err == nil {
		t.Error("Expected error when saving without presentation definition")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrIllegalArgument {
		t.Errorf("Expected error code %d, got %d", errors.ErrIllegalArgument, vpErr.Code)
	}
}

// TestModifyPresentationDefinitionData_SaveSuccess tests successful save
func TestModifyPresentationDefinitionData_SaveSuccess(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()
	pd := map[string]interface{}{
		"id": "test-pd",
		"input_descriptors": []interface{}{},
	}

	// When
	err := service.ModifyPresentationDefinitionData(ctx, "save", "business-id", "serial-no", pd)

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestModifyPresentationDefinitionData_DeleteSuccess tests successful delete
func TestModifyPresentationDefinitionData_DeleteSuccess(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	// When
	err := service.ModifyPresentationDefinitionData(ctx, "delete", "business-id", "serial-no", nil)

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestModifyPresentationDefinitionData_InvalidMode tests with invalid mode
func TestModifyPresentationDefinitionData_InvalidMode(t *testing.T) {
	// Given
	service := NewVerifierService("http://localhost:8080/verify")
	ctx := context.Background()

	// When
	err := service.ModifyPresentationDefinitionData(ctx, "invalid-mode", "business-id", "serial-no", nil)

	// Then
	if err == nil {
		t.Error("Expected error for invalid mode")
	}

	vpErr, ok := err.(*errors.VPError)
	if !ok {
		t.Errorf("Expected VPError, got %T", err)
	}

	if vpErr.Code != errors.ErrIllegalArgument {
		t.Errorf("Expected error code %d, got %d", errors.ErrIllegalArgument, vpErr.Code)
	}
}
