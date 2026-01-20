package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/moda-gov-tw/twdiw-issuer-go/pkg/errors"
	"github.com/moda-gov-tw/twdiw-issuer-go/pkg/models"
)

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	service := NewService("did:example:issuer", "issuer-key")
	if service == nil {
		t.Error("NewService returned nil")
	}
	if service.issuerDID != "did:example:issuer" {
		t.Errorf("Expected issuerDID to be set, got %s", service.issuerDID)
	}
}

// TestGenerate_NullRequest tests generation with nil request
func TestGenerate_NullRequest(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	result, status, err := service.Generate(ctx, nil)

	// Then
	if err == nil {
		t.Error("Expected error for nil request")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialGenerationRequest {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialGenerationRequest, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}

	// Verify response format
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(float64); !ok || int(code) != errors.ErrCredInvalidCredentialGenerationRequest {
		t.Error("Response does not contain expected error code")
	}
}

// TestGenerate_MissingIssuerDID tests generation without issuer DID
func TestGenerate_MissingIssuerDID(t *testing.T) {
	// Given
	service := NewService("", "") // No issuer DID/key
	ctx := context.Background()
	request := &models.CredentialRequestDTO{
		CredentialType: "IdentityCredential",
		CredentialSubject: map[string]interface{}{
			"name": "Test User",
		},
	}

	// When
	result, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for missing issuer DID")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrSysNotRegisterDIDYetError {
		t.Errorf("Expected error code %d, got %d", errors.ErrSysNotRegisterDIDYetError, vcErr.Code)
	}

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, status)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}

// TestGenerate_MissingCredentialType tests generation without credential type
func TestGenerate_MissingCredentialType(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()
	request := &models.CredentialRequestDTO{
		CredentialSubject: map[string]interface{}{
			"name": "Test User",
		},
	}

	// When
	_, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for missing credential type")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialType {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialType, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestGenerate_MissingCredentialSubject tests generation without credential subject
func TestGenerate_MissingCredentialSubject(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()
	request := &models.CredentialRequestDTO{
		CredentialType: "IdentityCredential",
	}

	// When
	_, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for missing credential subject")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialSubject {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialSubject, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestGenerate_Success tests successful credential generation
func TestGenerate_Success(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()
	request := &models.CredentialRequestDTO{
		IssuerDID:      "did:example:issuer",
		CredentialType: "IdentityCredential",
		CredentialSubject: map[string]interface{}{
			"name": "Test User",
			"age":  30,
		},
		Nonce: "test-nonce-123",
	}

	// When
	result, status, err := service.Generate(ctx, request)

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Verify response format
	var response models.CredentialResponseDTO
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.CID == "" {
		t.Error("Expected CID to be set")
	}

	if response.Credential == "" {
		t.Error("Expected credential to be set")
	}

	if response.Nonce != request.Nonce {
		t.Errorf("Expected nonce %s, got %s", request.Nonce, response.Nonce)
	}
}

// TestQuery_InvalidCID tests query with invalid CID
func TestQuery_InvalidCID(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	_, status, err := service.Query(ctx, "")

	// Then
	if err == nil {
		t.Error("Expected error for invalid CID")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialID {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialID, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestQuery_NotFound tests query for non-existent credential
func TestQuery_NotFound(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	result, status, err := service.Query(ctx, "non-existent-cid")

	// Then
	if err == nil {
		t.Error("Expected error for non-existent credential")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredCredentialNotFound {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredCredentialNotFound, vcErr.Code)
	}

	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, status)
	}

	// Verify response contains error info
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	_ = response // used for verification
}

// TestQueryByNonce_InvalidNonce tests query with invalid nonce
func TestQueryByNonce_InvalidNonce(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	_, status, err := service.QueryByNonce(ctx, "")

	// Then
	if err == nil {
		t.Error("Expected error for invalid nonce")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidNonce {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidNonce, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestQueryByNonce_NotFound tests query by nonce for non-existent credential
func TestQueryByNonce_NotFound(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	result, status, err := service.QueryByNonce(ctx, "non-existent-nonce")

	// Then
	if err == nil {
		t.Error("Expected error for non-existent credential")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredCredentialNotFound {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredCredentialNotFound, vcErr.Code)
	}

	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, status)
	}

	_ = result // result verified via error
}

// TestRevoke_InvalidCID tests revoke with invalid CID
func TestRevoke_InvalidCID(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	_, status, err := service.Revoke(ctx, "")

	// Then
	if err == nil {
		t.Error("Expected error for invalid CID")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialID {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialID, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestRevoke_Success tests successful credential revocation
func TestRevoke_Success(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	result, status, err := service.Revoke(ctx, "test-cid")

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "REVOKED" {
		t.Errorf("Expected status REVOKED, got %v", response["status"])
	}
}

// TestSuspend_Success tests successful credential suspension
func TestSuspend_Success(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	result, status, err := service.Suspend(ctx, "test-cid")

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "SUSPENDED" {
		t.Errorf("Expected status SUSPENDED, got %v", response["status"])
	}
}

// TestRecover_Success tests successful credential recovery
func TestRecover_Success(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// When
	result, status, err := service.Recover(ctx, "test-cid")

	// Then
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "ACTIVE" {
		t.Errorf("Expected status ACTIVE, got %v", response["status"])
	}
}

// TestGenerate_CredentialSubjectTooLarge tests generation with too many fields in credential subject
func TestGenerate_CredentialSubjectTooLarge(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// Create credential subject with more than MaxCredentialSubjectEntries
	credentialSubject := make(map[string]interface{})
	for i := 0; i < MaxCredentialSubjectEntries+1; i++ {
		credentialSubject[fmt.Sprintf("field%d", i)] = "value"
	}

	request := &models.CredentialRequestDTO{
		IssuerDID:         "did:example:issuer",
		CredentialType:    "IdentityCredential",
		CredentialSubject: credentialSubject,
	}

	// When
	_, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for oversized credential subject")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialSubject {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialSubject, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestGenerate_StringTooLong tests generation with oversized string in credential subject
func TestGenerate_StringTooLong(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// Create a very long string (exceeds MaxStringLength)
	longString := make([]byte, MaxStringLength+1)
	for i := range longString {
		longString[i] = 'A'
	}

	request := &models.CredentialRequestDTO{
		IssuerDID:      "did:example:issuer",
		CredentialType: "IdentityCredential",
		CredentialSubject: map[string]interface{}{
			"longField": string(longString),
		},
	}

	// When
	_, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for oversized string")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialSubject {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialSubject, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestGenerate_DeeplyNestedMap tests generation with deeply nested credential subject
func TestGenerate_DeeplyNestedMap(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// Create nested maps exceeding MaxMapDepth
	deeplyNested := make(map[string]interface{})
	current := deeplyNested
	for i := 0; i < MaxMapDepth+1; i++ {
		nested := make(map[string]interface{})
		current[fmt.Sprintf("level%d", i)] = nested
		current = nested
	}
	current["value"] = "too deep"

	request := &models.CredentialRequestDTO{
		IssuerDID:         "did:example:issuer",
		CredentialType:    "IdentityCredential",
		CredentialSubject: deeplyNested,
	}

	// When
	_, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for deeply nested map")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialSubject {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialSubject, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}

// TestGenerate_CredentialTypeTooLong tests generation with oversized credential type
func TestGenerate_CredentialTypeTooLong(t *testing.T) {
	// Given
	service := NewService("did:example:issuer", "issuer-key")
	ctx := context.Background()

	// Create a very long credential type
	longType := make([]byte, MaxStringLength+1)
	for i := range longType {
		longType[i] = 'B'
	}

	request := &models.CredentialRequestDTO{
		IssuerDID:      "did:example:issuer",
		CredentialType: string(longType),
		CredentialSubject: map[string]interface{}{
			"name": "Test",
		},
	}

	// When
	_, status, err := service.Generate(ctx, request)

	// Then
	if err == nil {
		t.Error("Expected error for oversized credential type")
	}

	vcErr, ok := err.(*errors.VCError)
	if !ok {
		t.Errorf("Expected VCError, got %T", err)
	}

	if vcErr.Code != errors.ErrCredInvalidCredentialType {
		t.Errorf("Expected error code %d, got %d", errors.ErrCredInvalidCredentialType, vcErr.Code)
	}

	if status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, status)
	}
}
