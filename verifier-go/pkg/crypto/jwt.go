package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// VCClaims represents the claims in a Verifiable Credential JWT
type VCClaims struct {
	jwt.RegisteredClaims
	VC CredentialSubject `json:"vc"`
}

// VPClaims represents the claims in a Verifiable Presentation JWT
type VPClaims struct {
	jwt.RegisteredClaims
	VP PresentationSubject `json:"vp"`
}

// CredentialSubject represents the credential subject in a VC
type CredentialSubject struct {
	Context           []string               `json:"@context"`
	Type              []string               `json:"type"`
	CredentialSubject map[string]interface{} `json:"credentialSubject"`
	Issuer            string                 `json:"issuer,omitempty"`
	IssuanceDate      string                 `json:"issuanceDate,omitempty"`
	ExpirationDate    string                 `json:"expirationDate,omitempty"`
	CredentialStatus  *CredentialStatus      `json:"credentialStatus,omitempty"`
}

// PresentationSubject represents the presentation in a VP
type PresentationSubject struct {
	Context              []string `json:"@context"`
	Type                 []string `json:"type"`
	VerifiableCredential []string `json:"verifiableCredential"`
	Holder               string   `json:"holder,omitempty"`
}

// CredentialStatus represents the credential status
type CredentialStatus struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	StatusListIndex string `json:"statusListIndex,omitempty"`
}

// JWTValidator handles JWT validation
type JWTValidator struct {
	// KeyResolver resolves public keys from DID
	KeyResolver KeyResolver
}

// KeyResolver interface for resolving public keys
type KeyResolver interface {
	ResolveKey(did string) (interface{}, error)
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(resolver KeyResolver) *JWTValidator {
	return &JWTValidator{
		KeyResolver: resolver,
	}
}

// ValidateVC validates a Verifiable Credential JWT
func (v *JWTValidator) ValidateVC(vcJWT string) (*VCClaims, error) {
	// Parse JWT without validation first to get issuer
	token, _, err := new(jwt.Parser).ParseUnverified(vcJWT, &VCClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse VC JWT: %w", err)
	}

	claims, ok := token.Claims.(*VCClaims)
	if !ok {
		return nil, fmt.Errorf("invalid VC claims")
	}

	// Get issuer DID from claims
	issuerDID := claims.Issuer
	if issuerDID == "" && claims.VC.Issuer != "" {
		issuerDID = claims.VC.Issuer
	}
	if issuerDID == "" {
		return nil, fmt.Errorf("issuer not found in VC")
	}

	// Resolve public key
	publicKey, err := v.KeyResolver.ResolveKey(issuerDID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve issuer key: %w", err)
	}

	// Parse and validate JWT with public key
	validatedToken, err := jwt.ParseWithClaims(vcJWT, &VCClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA, *jwt.SigningMethodRSA, *jwt.SigningMethodEd25519:
			return publicKey, nil
		default:
			return nil, fmt.Errorf("unsupported signing method: %v", token.Method.Alg())
		}
	})

	if err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	validatedClaims, ok := validatedToken.Claims.(*VCClaims)
	if !ok {
		return nil, fmt.Errorf("invalid validated claims")
	}

	// Validate expiration
	if validatedClaims.ExpiresAt != nil && validatedClaims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("credential has expired")
	}

	// Validate not before
	if validatedClaims.NotBefore != nil && validatedClaims.NotBefore.After(time.Now()) {
		return nil, fmt.Errorf("credential not yet valid")
	}

	// Validate expiration date in VC
	if validatedClaims.VC.ExpirationDate != "" {
		expTime, err := time.Parse(time.RFC3339, validatedClaims.VC.ExpirationDate)
		if err == nil && expTime.Before(time.Now()) {
			return nil, fmt.Errorf("credential has expired (VC expirationDate)")
		}
	}

	return validatedClaims, nil
}

// ValidateVP validates a Verifiable Presentation JWT
func (v *JWTValidator) ValidateVP(vpJWT string, expectedNonce string, expectedAudience string) (*VPClaims, error) {
	// Parse JWT without validation first to get holder
	token, _, err := new(jwt.Parser).ParseUnverified(vpJWT, &VPClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse VP JWT: %w", err)
	}

	claims, ok := token.Claims.(*VPClaims)
	if !ok {
		return nil, fmt.Errorf("invalid VP claims")
	}

	// Get holder DID
	holderDID := claims.Subject
	if holderDID == "" && claims.VP.Holder != "" {
		holderDID = claims.VP.Holder
	}
	if holderDID == "" {
		return nil, fmt.Errorf("holder not found in VP")
	}

	// Resolve public key
	publicKey, err := v.KeyResolver.ResolveKey(holderDID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve holder key: %w", err)
	}

	// Parse and validate JWT with public key
	validatedToken, err := jwt.ParseWithClaims(vpJWT, &VPClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA, *jwt.SigningMethodRSA, *jwt.SigningMethodEd25519:
			return publicKey, nil
		default:
			return nil, fmt.Errorf("unsupported signing method: %v", token.Method.Alg())
		}
	})

	if err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	validatedClaims, ok := validatedToken.Claims.(*VPClaims)
	if !ok {
		return nil, fmt.Errorf("invalid validated claims")
	}

	// Validate nonce
	if expectedNonce != "" && validatedClaims.ID != expectedNonce {
		return nil, fmt.Errorf("nonce mismatch: expected %s, got %s", expectedNonce, validatedClaims.ID)
	}

	// Validate audience
	expectedAudiences := validatedClaims.Audience
	if expectedAudience != "" {
		found := false
		for _, aud := range expectedAudiences {
			if aud == expectedAudience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("audience mismatch: %s not in %v", expectedAudience, expectedAudiences)
		}
	}

	// Validate expiration
	if validatedClaims.ExpiresAt != nil && validatedClaims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("presentation has expired")
	}

	// Validate not before
	if validatedClaims.NotBefore != nil && validatedClaims.NotBefore.After(time.Now()) {
		return nil, fmt.Errorf("presentation not yet valid")
	}

	return validatedClaims, nil
}

// SignVC creates a signed Verifiable Credential JWT
func SignVC(claims *VCClaims, privateKey interface{}, kid string) (string, error) {
	// Determine signing method based on key type
	var method jwt.SigningMethod
	switch privateKey.(type) {
	case *ecdsa.PrivateKey:
		method = jwt.SigningMethodES256
	case *rsa.PrivateKey:
		method = jwt.SigningMethodRS256
	case ed25519.PrivateKey:
		method = jwt.SigningMethodEdDSA
	default:
		return "", fmt.Errorf("unsupported private key type")
	}

	// Create token
	token := jwt.NewWithClaims(method, claims)

	// Add key ID to header
	if kid != "" {
		token.Header["kid"] = kid
	}

	// Sign token
	signedString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign VC: %w", err)
	}

	return signedString, nil
}

// SignVP creates a signed Verifiable Presentation JWT
func SignVP(claims *VPClaims, privateKey interface{}, kid string) (string, error) {
	// Determine signing method based on key type
	var method jwt.SigningMethod
	switch privateKey.(type) {
	case *ecdsa.PrivateKey:
		method = jwt.SigningMethodES256
	case *rsa.PrivateKey:
		method = jwt.SigningMethodRS256
	case ed25519.PrivateKey:
		method = jwt.SigningMethodEdDSA
	default:
		return "", fmt.Errorf("unsupported private key type")
	}

	// Create token
	token := jwt.NewWithClaims(method, claims)

	// Add key ID to header
	if kid != "" {
		token.Header["kid"] = kid
	}

	// Sign token
	signedString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign VP: %w", err)
	}

	return signedString, nil
}

// ParsePublicKeyPEM parses a PEM-encoded public key
func ParsePublicKeyPEM(pemData string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try parsing as different key types
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try EC public key
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse public key")
}

// ParsePrivateKeyPEM parses a PEM-encoded private key
func ParsePrivateKeyPEM(pemData string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try parsing as PKCS8
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try EC private key
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try RSA private key
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse private key")
}

// ExtractDIDFromJWT extracts the DID from JWT without validation
func ExtractDIDFromJWT(jwtString string, claimName string) (string, error) {
	parts := strings.Split(jwtString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	// Extract DID from specified claim
	if did, ok := claims[claimName].(string); ok && did != "" {
		return did, nil
	}

	return "", fmt.Errorf("DID not found in claim: %s", claimName)
}
