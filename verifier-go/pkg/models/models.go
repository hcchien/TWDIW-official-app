package models

// PresentationValidationRequest represents a request to validate presentations
type PresentationValidationRequest struct {
	Presentations []string `json:"presentations"`
}

// PresentationValidationResponse represents the response from VP validation
type PresentationValidationResponse struct {
	ClientID           string                   `json:"client_id,omitempty"`
	Nonce              string                   `json:"nonce,omitempty"`
	HolderDID          string                   `json:"holder_did,omitempty"`
	VerifiableCredentials []VerifiableCredentialData `json:"vcs,omitempty"`
}

// VerifiableCredentialData represents credential data within a VP
type VerifiableCredentialData struct {
	VPPath                   string                 `json:"vp_path,omitempty"`
	VCPath                   string                 `json:"vc_path,omitempty"`
	HolderPublicKey          map[string]interface{} `json:"holder_public_key,omitempty"`
	Credential               map[string]interface{} `json:"credential,omitempty"`
	Sub                      string                 `json:"sub,omitempty"`
	LimitDisclosureSupported bool                   `json:"limit_disclosure_supported,omitempty"`
	// Additional fields from cryptographic validation
	IssuerDID         string                 `json:"issuer_did,omitempty"`
	CredentialTypes   []string               `json:"credential_types,omitempty"`
	CredentialSubject map[string]interface{} `json:"credential_subject,omitempty"`
	IssuanceDate      string                 `json:"issuance_date,omitempty"`
	ExpirationDate    string                 `json:"expiration_date,omitempty"`
}

// VerifyResult represents the result of OID4VP verification
type VerifyResult struct {
	VerifyResult bool                       `json:"verify_result"`
	HolderDID    string                     `json:"holder_did,omitempty"`
	VCClaims     []VCResponseObject         `json:"vc_claims,omitempty"`
	CustomData   map[string]interface{}     `json:"custom_data,omitempty"`
	Error        *ErrorInfo                 `json:"error,omitempty"`
}

// VCResponseObject represents a VC response object
type VCResponseObject struct {
	CredentialType string                 `json:"credential_type"`
	Claims         map[string]interface{} `json:"claims"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OIDVPAuthorizationResponse represents OID4VP authorization response
type OIDVPAuthorizationResponse struct {
	VPToken               string `json:"vp_token"`
	PresentationSubmission string `json:"presentation_submission"`
	Error                 string `json:"error,omitempty"`
	ErrorDescription      string `json:"error_description,omitempty"`
	CustomData            string `json:"custom_data,omitempty"`
}

// IsSuccess checks if the authorization response is successful
func (r *OIDVPAuthorizationResponse) IsSuccess() bool {
	return r.Error == ""
}
