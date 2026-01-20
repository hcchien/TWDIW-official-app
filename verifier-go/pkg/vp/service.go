package vp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/crypto"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/errors"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
)

// Validation limits to prevent DoS attacks
const (
	MaxPresentations       = 100       // Maximum number of presentations in a single request
	MaxPresentationSize    = 1048576   // 1MB - Maximum size of a single presentation string
	MaxTotalPayloadSize    = 10485760  // 10MB - Maximum total size of all presentations
)

// Service handles VP (Verifiable Presentation) validation
type Service struct {
	// JWT validator for cryptographic validation
	jwtValidator *crypto.JWTValidator
	// DID resolver for resolving public keys
	didResolver *crypto.DIDResolver
}

// NewService creates a new VP validation service
func NewService() *Service {
	resolver := crypto.NewDIDResolver()
	return &Service{
		jwtValidator: crypto.NewJWTValidator(resolver),
		didResolver:  resolver,
	}
}

// NewServiceWithResolver creates a new VP validation service with custom resolver
func NewServiceWithResolver(resolver *crypto.DIDResolver) *Service {
	return &Service{
		jwtValidator: crypto.NewJWTValidator(resolver),
		didResolver:  resolver,
	}
}

// Validate validates a list of verifiable presentations
// This is the Go equivalent of PresentationServiceAsync.validate()
func (s *Service) Validate(ctx context.Context, presentations []string) (string, int, error) {
	// Check for nil or empty presentation list
	if presentations == nil || len(presentations) == 0 {
		vpErr := errors.NewVPError(
			errors.ErrPresInvalidPresentationValidationRequest,
			"presentations list cannot be empty",
		)
		response, _ := json.Marshal(vpErr.Response())
		return string(response), vpErr.HTTPStatus(), vpErr
	}

	// Validate array size to prevent DoS
	if len(presentations) > MaxPresentations {
		vpErr := errors.NewVPError(
			errors.ErrPresInvalidPresentationValidationRequest,
			fmt.Sprintf("too many presentations: maximum %d allowed", MaxPresentations),
		)
		response, _ := json.Marshal(vpErr.Response())
		return string(response), vpErr.HTTPStatus(), vpErr
	}

	// Validate individual presentation sizes and total payload size
	var totalSize int
	for i, presentation := range presentations {
		presentationSize := len(presentation)

		// Check individual presentation size
		if presentationSize > MaxPresentationSize {
			vpErr := errors.NewVPError(
				errors.ErrPresInvalidPresentationValidationRequest,
				fmt.Sprintf("presentation at index %d exceeds maximum size of %d bytes", i, MaxPresentationSize),
			)
			response, _ := json.Marshal(vpErr.Response())
			return string(response), vpErr.HTTPStatus(), vpErr
		}

		// Accumulate total size
		totalSize += presentationSize

		// Check total payload size
		if totalSize > MaxTotalPayloadSize {
			vpErr := errors.NewVPError(
				errors.ErrPresInvalidPresentationValidationRequest,
				fmt.Sprintf("total payload exceeds maximum size of %d bytes", MaxTotalPayloadSize),
			)
			response, _ := json.Marshal(vpErr.Response())
			return string(response), vpErr.HTTPStatus(), vpErr
		}
	}

	// Validate each VP
	results, err := s.validateVPs(ctx, presentations)
	if err != nil {
		if vpErr, ok := err.(*errors.VPError); ok {
			response, _ := json.Marshal(vpErr.Response())
			return string(response), vpErr.HTTPStatus(), vpErr
		}
		// Unexpected error - sanitize message to prevent information leakage
		// Log internal error details server-side (would go to logging system)
		// Return generic error to client
		vpErr := errors.NewVPError(
			errors.ErrPresValidateVPError,
			"presentation validation failed",
		)
		response, _ := json.Marshal(vpErr.Response())
		return string(response), vpErr.HTTPStatus(), vpErr
	}

	// Return successful response
	response, _ := json.Marshal(results)
	return string(response), http.StatusOK, nil
}

// validateVPs validates multiple VPs
func (s *Service) validateVPs(ctx context.Context, presentations []string) ([]models.PresentationValidationResponse, error) {
	var results []models.PresentationValidationResponse
	isArray := len(presentations) > 1

	for vpIndex, presentation := range presentations {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, errors.NewVPError(
				errors.Unknown,
				"operation cancelled",
			)
		default:
			// Continue processing
		}

		// Skip blank or empty presentations
		if presentation == "" {
			continue
		}

		// Skip whitespace-only presentations (use efficient trimming)
		trimmed := strings.TrimSpace(presentation)
		if trimmed == "" {
			continue
		}

		// Validate individual VP
		result, err := s.validateVP(ctx, presentation, vpIndex, isArray)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// validateVP validates a single VP
func (s *Service) validateVP(ctx context.Context, presentation string, vpIndex int, isArray bool) (models.PresentationValidationResponse, error) {
	// 1. Parse and validate VP JWT signature
	vpClaims, err := s.jwtValidator.ValidateVP(presentation, "", "") // Nonce and audience validation optional for now
	if err != nil {
		return models.PresentationValidationResponse{}, errors.NewVPError(
			errors.ErrPresValidateVPError,
			fmt.Sprintf("VP validation failed: %v", err),
		)
	}

	// 2. Extract holder DID
	holderDID := vpClaims.Subject
	if holderDID == "" && vpClaims.VP.Holder != "" {
		holderDID = vpClaims.VP.Holder
	}

	// 3. Extract client_id and nonce
	clientID := ""
	nonce := vpClaims.ID

	// Extract client_id from audience if present
	if len(vpClaims.Audience) > 0 {
		clientID = vpClaims.Audience[0]
	}

	// 4. Validate embedded VCs
	var vcResults []models.VerifiableCredentialData
	for vcIndex, vcJWT := range vpClaims.VP.VerifiableCredential {
		vcResult, err := s.validateVC(ctx, vcJWT, vcIndex, holderDID)
		if err != nil {
			// For now, continue processing other VCs but log the error
			// In production, you might want to fail fast or collect all errors
			continue
		}
		vcResults = append(vcResults, vcResult)
	}

	// 5. Return validation response
	return models.PresentationValidationResponse{
		ClientID:              clientID,
		Nonce:                 nonce,
		HolderDID:             holderDID,
		VerifiableCredentials: vcResults,
	}, nil
}

// validateVC validates a single VC and extracts its data
func (s *Service) validateVC(ctx context.Context, vcJWT string, vcIndex int, expectedHolderDID string) (models.VerifiableCredentialData, error) {
	// 1. Parse and validate VC JWT signature
	vcClaims, err := s.jwtValidator.ValidateVC(vcJWT)
	if err != nil {
		return models.VerifiableCredentialData{}, errors.NewVPError(
			errors.ErrCredValidateVCProofError,
			fmt.Sprintf("VC validation failed: %v", err),
		)
	}

	// 2. Verify VC subject matches holder DID
	if vcClaims.Subject != expectedHolderDID {
		return models.VerifiableCredentialData{}, errors.NewVPError(
			errors.ErrPresHolderPublicKeyInconsistent,
			fmt.Sprintf("VC subject (%s) does not match VP holder (%s)", vcClaims.Subject, expectedHolderDID),
		)
	}

	// 3. Extract credential data
	issuerDID := vcClaims.Issuer
	if issuerDID == "" && vcClaims.VC.Issuer != "" {
		issuerDID = vcClaims.VC.Issuer
	}

	// 4. Extract credential types
	credentialTypes := vcClaims.VC.Type
	if credentialTypes == nil {
		credentialTypes = []string{}
	}

	// 5. Extract credential subject data
	credentialSubject := vcClaims.VC.CredentialSubject
	if credentialSubject == nil {
		credentialSubject = make(map[string]interface{})
	}

	// 6. Return VC data
	return models.VerifiableCredentialData{
		IssuerDID:         issuerDID,
		CredentialTypes:   credentialTypes,
		CredentialSubject: credentialSubject,
		IssuanceDate:      vcClaims.VC.IssuanceDate,
		ExpirationDate:    vcClaims.VC.ExpirationDate,
	}, nil
}

// Helper functions for path generation (matching Java implementation)
func getVPPath(vpIndex int, isArray bool) string {
	if isArray {
		return fmt.Sprintf("$[%d]", vpIndex)
	}
	return "$"
}

func getVCPath(vcIndex int) string {
	return fmt.Sprintf("$.vp.verifiableCredential[%d]", vcIndex)
}
