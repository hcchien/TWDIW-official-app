package vp

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/errors"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/mdl"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
)

// validateMDLPresentations validates mDL presentations in CBOR format
func (s *Service) validateMDLPresentations(ctx context.Context, presentations []string) ([]models.PresentationValidationResponse, error) {
	var results []models.PresentationValidationResponse

	// Create mDL validator
	mdlValidator := mdl.NewValidator()

	// Process each mDL presentation
	for _, presentation := range presentations {
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

		// Skip empty presentations
		if presentation == "" {
			continue
		}

		// Decode base64 CBOR data
		cborData, err := base64.StdEncoding.DecodeString(presentation)
		if err != nil {
			return nil, errors.NewVPError(
				errors.ErrMDLInvalidCBORStructure,
				fmt.Sprintf("invalid base64 encoding: %v", err),
			)
		}

		// Parse mDL document
		mdlDoc, err := mdlValidator.ParseDocument(cborData)
		if err != nil {
			return nil, errors.NewVPError(
				errors.ErrMDLInvalidCBORStructure,
				fmt.Sprintf("failed to parse mDL: %v", err),
			)
		}

		// Validate the mDL document
		mdlResponse, err := mdlValidator.ValidateDocument(mdlDoc)
		if err != nil {
			// Determine specific error code based on error type
			return nil, errors.NewVPError(
				errors.ErrMDLInvalidIssuerSignature,
				fmt.Sprintf("mDL validation failed: %v", err),
			)
		}

		// Convert to API response format
		docData := s.convertMDLResponseToDocumentData(mdlResponse, mdlDoc)

		// Build response
		response := models.PresentationValidationResponse{
			Format:       models.FormatISOMDL.String(),
			MDLDocuments: []models.MDLDocumentData{docData},
		}

		results = append(results, response)
	}

	return results, nil
}

// convertMDLResponseToDocumentData converts internal MDLResponse to API MDLDocumentData
func (s *Service) convertMDLResponseToDocumentData(mdlResp *models.MDLResponse, doc *models.MobileDocument) models.MDLDocumentData {
	// Extract issuer certificate if available
	var issuerCertPEM string
	// In a real implementation, you would extract the certificate from the COSE_Sign1
	// For now, we'll leave it empty
	issuerCertPEM = ""

	// Flatten claims from namespaces
	flattenedClaims := make(map[string]interface{})
	for ns, claims := range mdlResp.NameSpaces {
		for key, value := range claims {
			flattenedClaims[fmt.Sprintf("%s/%s", ns, key)] = value
		}
	}

	return models.MDLDocumentData{
		DocType:          mdlResp.DocType,
		IssuerCertificate: issuerCertPEM,
		DeviceKeyID:      mdlResp.DeviceKeyID,
		Claims:           flattenedClaims,
		IssuanceDate:     mdlResp.IssuanceDate.Format("2006-01-02T15:04:05Z"),
		ExpirationDate:   mdlResp.ExpirationDate.Format("2006-01-02T15:04:05Z"),
		ValidationStatus: mdlResp.ValidationStatus,
	}
}

// extractCertificatePEM converts an X.509 certificate to PEM format
func extractCertificatePEM(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}

	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	return string(pem.EncodeToMemory(pemBlock))
}
