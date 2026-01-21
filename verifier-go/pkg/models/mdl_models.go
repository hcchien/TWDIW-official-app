package models

import (
	"crypto/x509"
	"time"
)

// CredentialFormat represents the format of a credential
type CredentialFormat int

const (
	FormatUnknown CredentialFormat = iota
	FormatW3CJWT                    // W3C JWT-VC
	FormatISOMDL                    // ISO 18013-5 mDL CBOR
)

// String returns the string representation of the credential format
func (f CredentialFormat) String() string {
	switch f {
	case FormatW3CJWT:
		return "w3c_jwt"
	case FormatISOMDL:
		return "iso_mdl"
	default:
		return "unknown"
	}
}

// MobileDocument represents an ISO 18013-5 mDL document
type MobileDocument struct {
	DocType      string                 `cbor:"docType"`
	IssuerSigned IssuerSignedData       `cbor:"issuerSigned"`
	DeviceSigned DeviceSignedData       `cbor:"deviceSigned"`
	Errors       map[string]interface{} `cbor:"errors,omitempty"`
}

// IssuerSignedData contains issuer signature over mDL
type IssuerSignedData struct {
	NameSpaces map[string][]IssuerSignedItem `cbor:"nameSpaces"`
	IssuerAuth []byte                        `cbor:"issuerAuth"` // COSE_Sign1 structure
}

// IssuerSignedItem represents a single claim signed by issuer
type IssuerSignedItem struct {
	DigestID     uint64      `cbor:"digestID"`
	Random       []byte      `cbor:"random"`
	ElementID    string      `cbor:"elementIdentifier"`
	ElementValue interface{} `cbor:"elementValue"`
}

// DeviceSignedData contains device signature over mDL presentation
type DeviceSignedData struct {
	NameSpaces map[string]interface{} `cbor:"nameSpaces"`
	DeviceAuth DeviceAuth             `cbor:"deviceAuth"`
}

// DeviceAuth contains device authentication data
type DeviceAuth struct {
	DeviceSignature []byte `cbor:"deviceSignature,omitempty"` // COSE_Sign1 from device key
	DeviceMAC       []byte `cbor:"deviceMac,omitempty"`       // Optional MAC for sessionTranscript
}

// MobileSecurityObject represents the MSO from IssuerAuth
type MobileSecurityObject struct {
	Version         string                       `cbor:"version"`
	DigestAlgorithm string                       `cbor:"digestAlgorithm"`
	ValueDigests    map[string]map[uint64][]byte `cbor:"valueDigests"` // namespace -> digestID -> digest
	DeviceKeyInfo   DeviceKeyInfo                `cbor:"deviceKeyInfo"`
	DocType         string                       `cbor:"docType"`
	ValidityInfo    ValidityInfo                 `cbor:"validityInfo"`
}

// DeviceKeyInfo contains the device public key
type DeviceKeyInfo struct {
	DeviceKey map[interface{}]interface{} `cbor:"deviceKey"` // COSE_Key structure
}

// ValidityInfo contains validity dates
type ValidityInfo struct {
	Signed     time.Time  `cbor:"signed"`
	ValidFrom  time.Time  `cbor:"validFrom"`
	ValidUntil time.Time  `cbor:"validUntil"`
	ExpectedUpdate *time.Time `cbor:"expectedUpdate,omitempty"`
}

// DeviceEngagement represents BLE/NFC/WiFi engagement data
type DeviceEngagement struct {
	Version         string                  `cbor:"version"`
	Security        interface{}             `cbor:"security"`
	DeviceRetrieval []DeviceRetrievalMethod `cbor:"deviceRetrievalMethods,omitempty"`
	ServerRetrieval []interface{}           `cbor:"serverRetrievalMethods,omitempty"`
}

// DeviceRetrievalMethod specifies connection method (BLE/NFC/WiFi)
type DeviceRetrievalMethod struct {
	Type       uint64                 `cbor:"type"` // 1=BLE, 2=WiFi Aware, 3=NFC
	Version    uint64                 `cbor:"version"`
	RetrievalOptions map[interface{}]interface{} `cbor:"retrievalOptions"`
}

// ReaderAuthentication contains X.509 certificate chain
type ReaderAuthentication struct {
	Certificates []*x509.Certificate
	Signature    []byte // COSE_Sign1
}

// SessionTranscript binds device engagement to session
type SessionTranscript struct {
	DeviceEngagementBytes []byte
	EReaderKeyBytes       []byte
	Handover              []byte
}

// MDLRequest represents a request for mDL data
type MDLRequest struct {
	DocType     string                       `cbor:"docType"`
	NameSpaces  map[string]map[string]bool   `cbor:"nameSpaces"` // namespace -> element -> requested
	RequestInfo map[string]interface{}       `cbor:"requestInfo,omitempty"`
}

// MDLResponse represents validated mDL presentation response
type MDLResponse struct {
	DocType          string
	IssuerDID        string // Extracted from certificate
	DeviceKeyID      string
	NameSpaces       map[string]map[string]interface{}
	IssuanceDate     time.Time
	ExpirationDate   time.Time
	ValidationStatus ValidationStatus
}

// ValidationStatus tracks mDL validation results
type ValidationStatus struct {
	IssuerSignatureValid bool `json:"issuer_signature_valid"`
	DeviceSignatureValid bool `json:"device_signature_valid"`
	CertificateValid     bool `json:"certificate_valid"`
	NotExpired           bool `json:"not_expired"`
	DigestsValid         bool `json:"digests_valid"`
}

// MDLDocumentData represents a validated mDL document in API response
type MDLDocumentData struct {
	DocType           string                 `json:"doc_type"`
	IssuerCertificate string                 `json:"issuer_certificate,omitempty"` // PEM-encoded
	DeviceKeyID       string                 `json:"device_key_id,omitempty"`
	Claims            map[string]interface{} `json:"claims"` // Flattened namespace claims
	IssuanceDate      string                 `json:"issuance_date,omitempty"`
	ExpirationDate    string                 `json:"expiration_date,omitempty"`
	ValidationStatus  ValidationStatus       `json:"validation_status"`
}
