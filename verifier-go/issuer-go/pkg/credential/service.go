package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/moda-gov-tw/twdiw-issuer-go/pkg/errors"
	"github.com/moda-gov-tw/twdiw-issuer-go/pkg/models"
)

// Validation limits to prevent DoS attacks
const (
	MaxCredentialSubjectEntries = 1000    // Maximum number of fields in credential subject
	MaxStringLength             = 1048576 // 1MB - Maximum length of any string field
	MaxMapDepth                 = 10      // Maximum nesting depth for maps
)

// Service handles credential issuance and management
type Service struct {
	// Dependencies would go here (repositories, crypto services, etc.)
	issuerDID string
	issuerKey string
}

// NewService creates a new credential service
func NewService(issuerDID, issuerKey string) *Service {
	return &Service{
		issuerDID: issuerDID,
		issuerKey: issuerKey,
	}
}

// Generate generates a new verifiable credential
// Equivalent to Java's CredentialService.generate()
func (s *Service) Generate(ctx context.Context, request *models.CredentialRequestDTO) (string, int, error) {
	// Validate request
	if request == nil {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialGenerationRequest,
			"invalid credential generation request",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Validate issuer DID is set
	if s.issuerDID == "" || s.issuerKey == "" {
		vcErr := errors.NewVCError(
			errors.ErrSysNotRegisterDIDYetError,
			"issuer has not yet registered a DID",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Validate credential type
	if request.CredentialType == "" {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialType,
			"credential type is required",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Validate credential subject
	if request.CredentialSubject == nil || len(request.CredentialSubject) == 0 {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialSubject,
			"credential subject is required",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Validate credential subject size to prevent DoS
	if len(request.CredentialSubject) > MaxCredentialSubjectEntries {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialSubject,
			fmt.Sprintf("credential subject exceeds maximum %d entries", MaxCredentialSubjectEntries),
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Validate string lengths in credential subject
	if err := validateMapStringLengths(request.CredentialSubject, 0); err != nil {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialSubject,
			err.Error(),
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Validate credential type length
	if len(request.CredentialType) > MaxStringLength {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialType,
			fmt.Sprintf("credential type exceeds maximum length of %d", MaxStringLength),
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// Check for context cancellation before expensive operations
	select {
	case <-ctx.Done():
		vcErr := errors.NewVCError(
			errors.Unknown,
			"operation cancelled",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	default:
		// Continue processing
	}

	// In a full implementation, this would:
	// 1. Load credential policy
	// 2. Validate against schema
	// 3. Generate ticket number
	// 4. Create VC JSON structure
	// 5. Sign with issuer key
	// 6. Update status list
	// 7. Save to database

	// For now, return a placeholder response
	credentialResponse := &models.CredentialResponseDTO{
		CID:        fmt.Sprintf("cred-%d", time.Now().Unix()),
		Credential: "eyJhbGciOiJFUzI1NiJ9.credential.signature",
		Nonce:      request.Nonce,
	}

	response, _ := json.Marshal(credentialResponse)
	return string(response), http.StatusOK, nil
}

// Query queries a credential by CID or nonce
// Equivalent to Java's CredentialService.query()
func (s *Service) Query(ctx context.Context, cid string) (string, int, error) {
	// Validate CID
	if cid == "" {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialID,
			"invalid credential ID",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// In a full implementation, this would:
	// 1. Query database by CID
	// 2. Return credential if found

	// For now, check if credential exists (placeholder)
	// Simulate not found for demo
	vcErr := errors.NewVCError(
		errors.ErrCredCredentialNotFound,
		fmt.Sprintf("credential not found: %s", cid),
	)
	response, _ := json.Marshal(vcErr.Response())
	return string(response), vcErr.HTTPStatus(), vcErr
}

// QueryByNonce queries a credential by nonce
// Equivalent to Java's CredentialService.queryByNonce()
func (s *Service) QueryByNonce(ctx context.Context, nonce string) (string, int, error) {
	// Validate nonce
	if nonce == "" {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidNonce,
			"invalid nonce",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// In a full implementation, this would:
	// 1. Query database by nonce
	// 2. Return credential if found

	// For now, return not found
	vcErr := errors.NewVCError(
		errors.ErrCredCredentialNotFound,
		fmt.Sprintf("credential not found with nonce: %s", nonce),
	)
	response, _ := json.Marshal(vcErr.Response())
	return string(response), vcErr.HTTPStatus(), vcErr
}

// Revoke revokes a credential
// Equivalent to Java's CredentialService.revoke()
func (s *Service) Revoke(ctx context.Context, cid string) (string, int, error) {
	// Validate CID
	if cid == "" {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialID,
			"invalid credential ID",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// In a full implementation, this would:
	// 1. Query credential from database
	// 2. Check current status
	// 3. Update status to REVOKED
	// 4. Update status list
	// 5. Save changes

	// For now, return success placeholder
	result := map[string]interface{}{
		"cid":    cid,
		"status": "REVOKED",
	}
	response, _ := json.Marshal(result)
	return string(response), http.StatusOK, nil
}

// Suspend suspends a credential
// Equivalent to Java's CredentialService.suspend()
func (s *Service) Suspend(ctx context.Context, cid string) (string, int, error) {
	// Validate CID
	if cid == "" {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialID,
			"invalid credential ID",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// In a full implementation, this would:
	// 1. Query credential from database
	// 2. Check if already revoked (cannot suspend revoked credential)
	// 3. Update status to SUSPENDED
	// 4. Update status list
	// 5. Save changes

	// For now, return success placeholder
	result := map[string]interface{}{
		"cid":    cid,
		"status": "SUSPENDED",
	}
	response, _ := json.Marshal(result)
	return string(response), http.StatusOK, nil
}

// Recover recovers a suspended credential
// Equivalent to Java's CredentialService.recover()
func (s *Service) Recover(ctx context.Context, cid string) (string, int, error) {
	// Validate CID
	if cid == "" {
		vcErr := errors.NewVCError(
			errors.ErrCredInvalidCredentialID,
			"invalid credential ID",
		)
		response, _ := json.Marshal(vcErr.Response())
		return string(response), vcErr.HTTPStatus(), vcErr
	}

	// In a full implementation, this would:
	// 1. Query credential from database
	// 2. Check if revoked (cannot recover revoked credential)
	// 3. Update status to ACTIVE
	// 4. Update status list
	// 5. Save changes

	// For now, return success placeholder
	result := map[string]interface{}{
		"cid":    cid,
		"status": "ACTIVE",
	}
	response, _ := json.Marshal(result)
	return string(response), http.StatusOK, nil
}

// validateMapStringLengths validates string lengths and nesting depth in a map
func validateMapStringLengths(m map[string]interface{}, depth int) error {
	// Check max depth to prevent stack overflow
	if depth > MaxMapDepth {
		return fmt.Errorf("map nesting exceeds maximum depth of %d", MaxMapDepth)
	}

	for key, value := range m {
		// Validate key length
		if len(key) > MaxStringLength {
			return fmt.Errorf("map key exceeds maximum length of %d", MaxStringLength)
		}

		// Validate value based on type
		switch v := value.(type) {
		case string:
			if len(v) > MaxStringLength {
				return fmt.Errorf("string value for key '%s' exceeds maximum length of %d", key, MaxStringLength)
			}
		case map[string]interface{}:
			// Recursively validate nested maps
			if err := validateMapStringLengths(v, depth+1); err != nil {
				return err
			}
		case []interface{}:
			// Validate array elements
			for i, item := range v {
				if str, ok := item.(string); ok {
					if len(str) > MaxStringLength {
						return fmt.Errorf("string in array at key '%s'[%d] exceeds maximum length of %d", key, i, MaxStringLength)
					}
				} else if nestedMap, ok := item.(map[string]interface{}); ok {
					if err := validateMapStringLengths(nestedMap, depth+1); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
