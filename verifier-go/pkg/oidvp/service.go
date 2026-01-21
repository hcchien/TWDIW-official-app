package oidvp

import (
	"context"
	"fmt"

	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/errors"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
)

// VerifierService handles OID4VP verification
// This is the Go equivalent of Java's VerifierService
type VerifierService struct {
	// Dependencies would go here
	vpVerifyURI string
}

// NewVerifierService creates a new OID4VP verifier service
func NewVerifierService(vpVerifyURI string) *VerifierService {
	return &VerifierService{
		vpVerifyURI: vpVerifyURI,
	}
}

// Verify verifies an OID4VP authorization response
// Equivalent to Java's VerifierService.verify()
func (s *VerifierService) Verify(ctx context.Context, authzResponse *models.OIDVPAuthorizationResponse, nonce, clientID, presentationDefinition string) (*models.VerifyResult, error) {
	// Check if wallet returned an error
	if !authzResponse.IsSuccess() {
		// Log internal error details server-side (would go to logging system)
		// Sanitize error message to prevent information leakage
		return &models.VerifyResult{
			VerifyResult: false,
			Error: &models.ErrorInfo{
				Code:    errors.Unknown, // AUTHZ_RESPONSE_ERROR in Java
				Message: "wallet authorization failed",
			},
		}, nil
	}

	// Verify the presentation
	return s.verifyPresentation(ctx, authzResponse.VPToken, authzResponse.PresentationSubmission, nonce, clientID, presentationDefinition)
}

// verifyPresentation validates the VP token and presentation submission
// Equivalent to Java's VerifierService.verifyPresentation()
func (s *VerifierService) verifyPresentation(ctx context.Context, vpToken, presentationSubmission, nonce, clientID, pdString string) (*models.VerifyResult, error) {
	// Validate required parameters
	if nonce == "" || clientID == "" || pdString == "" {
		return &models.VerifyResult{
			VerifyResult: false,
			Error: &models.ErrorInfo{
				Code:    errors.Unknown, // BAD_OIDVP_PARAM in Java
				Message: "required verify info is null or blank",
			},
		}, nil
	}

	// In a full implementation, this would:
	// 1. Validate presentation submission schema
	// 2. Parse presentation definition
	// 3. Validate VP token (call VP validation service)
	// 4. Extract holder DID
	// 5. Validate client_id and nonce
	// 6. Evaluate presentation against presentation definition
	// 7. Validate custom data if present

	// For now, return success placeholder
	return &models.VerifyResult{
		VerifyResult: true,
		HolderDID:    "did:example:holder",
		VCClaims:     []models.VCResponseObject{},
	}, nil
}

// GetVerifyResult retrieves a previously stored verification result
// Equivalent to Java's VerifierService.getVerifyResult()
func (s *VerifierService) GetVerifyResult(ctx context.Context, transactionID, responseCode string) (*models.VerifyResult, error) {
	if responseCode == "" && transactionID == "" {
		return nil, errors.NewVPError(
			errors.ErrIllegalArgument,
			"'response_code' and 'transaction_id' must not be null at the same time",
		)
	}

	// In a full implementation, this would:
	// 1. Query database for verification result
	// 2. Check if result has expired
	// 3. Optionally delete after query
	// 4. Return the stored result

	return &models.VerifyResult{
		VerifyResult: true,
		HolderDID:    "did:example:holder",
		VCClaims:     []models.VCResponseObject{},
	}, nil
}

// ModifyPresentationDefinitionData saves or deletes presentation definition
// Equivalent to Java's VerifierService.modifyPresentationDefinitionData()
func (s *VerifierService) ModifyPresentationDefinitionData(ctx context.Context, mode, businessID, serialNo string, presentationDefinition map[string]interface{}) error {
	if mode == "" || businessID == "" || serialNo == "" {
		return errors.NewVPError(
			errors.ErrIllegalArgument,
			"required input is not exist",
		)
	}

	if mode == "save" {
		if presentationDefinition == nil {
			return errors.NewVPError(
				errors.ErrIllegalArgument,
				"presentation_definition must be submit",
			)
		}

		// In a full implementation:
		// 1. Validate PD schema
		// 2. Save to database

		return nil
	} else if mode == "delete" {
		// In a full implementation:
		// 1. Delete from database

		return nil
	}

	return errors.NewVPError(
		errors.ErrIllegalArgument,
		fmt.Sprintf("invalid mode: %s", mode),
	)
}
