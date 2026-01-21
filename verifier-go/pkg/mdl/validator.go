package mdl

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
	xcrypto "github.com/moda-gov-tw/twdiw-verifier-go/pkg/crypto"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
)

// Validator handles mDL document validation
type Validator struct {
	coseValidator *xcrypto.COSEValidator
	certValidator *xcrypto.X509Validator
	trustedRoots  []*x509.Certificate
}

// NewValidator creates a new mDL validator
func NewValidator() *Validator {
	return &Validator{
		coseValidator: xcrypto.NewCOSEValidator(),
		certValidator: xcrypto.NewX509Validator(),
		trustedRoots:  make([]*x509.Certificate, 0),
	}
}

// AddTrustedRoot adds a trusted root certificate
func (v *Validator) AddTrustedRoot(cert *x509.Certificate) {
	v.trustedRoots = append(v.trustedRoots, cert)
	v.certValidator.AddTrustedRoot(cert)
}

// ParseDocument parses CBOR-encoded mDL document
func (v *Validator) ParseDocument(cborData []byte) (*models.MobileDocument, error) {
	var doc models.MobileDocument

	// Decode CBOR using fxamacker/cbor (supports CBOR tags)
	if err := cbor.Unmarshal(cborData, &doc); err != nil {
		return nil, fmt.Errorf("CBOR decode failed: %w", err)
	}

	// Validate structure
	if doc.DocType == "" {
		return nil, fmt.Errorf("missing docType in mDL document")
	}

	// Validate it's an mDL document (can be extended for other ISO 18013-5 doc types)
	if doc.DocType != "org.iso.18013.5.1.mDL" {
		return nil, fmt.Errorf("unsupported docType: %s", doc.DocType)
	}

	return &doc, nil
}

// ValidateIssuerAuth validates COSE_Sign1 issuer signature
func (v *Validator) ValidateIssuerAuth(doc *models.MobileDocument) (*x509.Certificate, *models.MobileSecurityObject, error) {
	// Extract IssuerAuth COSE_Sign1 structure
	coseSign1 := doc.IssuerSigned.IssuerAuth

	// Parse COSE_Sign1
	payload, _, protected, _, err := v.coseValidator.ParseCOSESign1(coseSign1)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse COSE_Sign1: %w", err)
	}

	// Extract X.509 certificate from protected header
	cert, err := v.coseValidator.ExtractCertificateFromCOSE(protected)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract certificate: %w", err)
	}

	// Validate certificate
	if err := v.certValidator.ValidateIssuerCert(cert, v.trustedRoots); err != nil {
		return nil, nil, fmt.Errorf("certificate validation failed: %w", err)
	}

	// Verify COSE signature using certificate public key
	if err := v.coseValidator.VerifySignature(coseSign1, cert.PublicKey); err != nil {
		return nil, nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Parse MSO (Mobile Security Object) from payload
	var mso models.MobileSecurityObject
	if err := cbor.Unmarshal(payload, &mso); err != nil {
		return nil, nil, fmt.Errorf("failed to parse MSO: %w", err)
	}

	// Validate digest integrity for each IssuerSignedItem
	if err := v.validateDigests(doc, &mso); err != nil {
		return nil, nil, fmt.Errorf("digest validation failed: %w", err)
	}

	return cert, &mso, nil
}

// validateDigests verifies hash digests of disclosed items
func (v *Validator) validateDigests(doc *models.MobileDocument, mso *models.MobileSecurityObject) error {
	// Get digest algorithm (typically "SHA-256")
	if mso.DigestAlgorithm != "SHA-256" {
		return fmt.Errorf("unsupported digest algorithm: %s", mso.DigestAlgorithm)
	}

	// For each namespace
	for ns, items := range doc.IssuerSigned.NameSpaces {
		expectedDigests, ok := mso.ValueDigests[ns]
		if !ok {
			return fmt.Errorf("no digests found for namespace: %s", ns)
		}

		for _, item := range items {
			// Encode the IssuerSignedItem for hashing
			itemCBOR, err := cbor.Marshal(item)
			if err != nil {
				return fmt.Errorf("failed to encode item for digest: %w", err)
			}

			// Compute SHA-256 digest
			hash := sha256.Sum256(itemCBOR)
			computed := hash[:]

			// Compare with digest in MSO
			expected, ok := expectedDigests[item.DigestID]
			if !ok {
				return fmt.Errorf("no expected digest for digestID %d in namespace %s", item.DigestID, ns)
			}

			if !bytes.Equal(computed, expected) {
				return fmt.Errorf("digest mismatch for %s/%s (digestID: %d): computed %s, expected %s",
					ns, item.ElementID, item.DigestID,
					hex.EncodeToString(computed), hex.EncodeToString(expected))
			}
		}
	}

	return nil
}

// ValidateDeviceAuth validates device key signature over presentation
func (v *Validator) ValidateDeviceAuth(doc *models.MobileDocument, mso *models.MobileSecurityObject) error {
	deviceAuth := doc.DeviceSigned.DeviceAuth

	// Check if device signature exists
	if len(deviceAuth.DeviceSignature) == 0 {
		return fmt.Errorf("missing device signature")
	}

	// Parse device COSE_Sign1
	payload, _, _, _, err := v.coseValidator.ParseCOSESign1(deviceAuth.DeviceSignature)
	if err != nil {
		return fmt.Errorf("failed to parse device COSE_Sign1: %w", err)
	}

	// Extract device public key from MSO
	devicePubKey, err := v.extractDevicePublicKey(mso)
	if err != nil {
		return fmt.Errorf("failed to extract device public key: %w", err)
	}

	// Verify device signature
	if err := v.coseValidator.VerifySignature(deviceAuth.DeviceSignature, devicePubKey); err != nil {
		return fmt.Errorf("device signature verification failed: %w", err)
	}

	// Parse and validate payload contains DeviceAuthentication structure
	// (namespace claims + session transcript binding)
	_ = payload // Payload validation would go here

	return nil
}

// extractDevicePublicKey extracts the device public key from MSO
func (v *Validator) extractDevicePublicKey(mso *models.MobileSecurityObject) (crypto.PublicKey, error) {
	// The deviceKey is stored as a COSE_Key structure in the MSO
	coseKey := mso.DeviceKeyInfo.DeviceKey

	// Convert COSE_Key to Go public key
	pubKey, err := v.coseValidator.GetPublicKeyFromCOSEKey(coseKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert COSE_Key to public key: %w", err)
	}

	return pubKey, nil
}

// ValidateExpiration checks validity dates from MSO
func (v *Validator) ValidateExpiration(mso *models.MobileSecurityObject) error {
	now := time.Now()

	if mso.ValidityInfo.ValidFrom.After(now) {
		return fmt.Errorf("document not yet valid (validFrom: %v)", mso.ValidityInfo.ValidFrom)
	}

	if mso.ValidityInfo.ValidUntil.Before(now) {
		return fmt.Errorf("document has expired (validUntil: %v)", mso.ValidityInfo.ValidUntil)
	}

	return nil
}

// ExtractClaims extracts and flattens claims from mDL namespaces
func (v *Validator) ExtractClaims(doc *models.MobileDocument) map[string]interface{} {
	claims := make(map[string]interface{})

	// Flatten namespace/element structure to namespace/element keys
	for ns, items := range doc.IssuerSigned.NameSpaces {
		for _, item := range items {
			key := fmt.Sprintf("%s/%s", ns, item.ElementID)
			claims[key] = item.ElementValue
		}
	}

	return claims
}

// ValidateDocument performs complete validation of an mDL document
func (v *Validator) ValidateDocument(doc *models.MobileDocument) (*models.MDLResponse, error) {
	response := &models.MDLResponse{
		DocType: doc.DocType,
		ValidationStatus: models.ValidationStatus{
			IssuerSignatureValid: false,
			DeviceSignatureValid: false,
			CertificateValid:     false,
			NotExpired:           false,
			DigestsValid:         false,
		},
	}

	// Validate issuer signature and extract MSO
	cert, mso, err := v.ValidateIssuerAuth(doc)
	if err != nil {
		return response, fmt.Errorf("issuer auth validation failed: %w", err)
	}
	response.ValidationStatus.IssuerSignatureValid = true
	response.ValidationStatus.CertificateValid = true
	response.ValidationStatus.DigestsValid = true

	// Extract issuer DID/identifier from certificate subject
	if cert != nil {
		response.IssuerDID = cert.Subject.String()
	}

	// Validate device signature
	if err := v.ValidateDeviceAuth(doc, mso); err != nil {
		return response, fmt.Errorf("device auth validation failed: %w", err)
	}
	response.ValidationStatus.DeviceSignatureValid = true

	// Validate expiration
	if err := v.ValidateExpiration(mso); err != nil {
		return response, fmt.Errorf("expiration validation failed: %w", err)
	}
	response.ValidationStatus.NotExpired = true

	// Extract claims
	claims := v.ExtractClaims(doc)
	response.NameSpaces = map[string]map[string]interface{}{
		doc.DocType: claims,
	}

	// Set dates
	response.IssuanceDate = mso.ValidityInfo.ValidFrom
	response.ExpirationDate = mso.ValidityInfo.ValidUntil

	return response, nil
}
