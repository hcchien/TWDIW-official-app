package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignAndValidateVC(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create DID resolver with local key
	resolver := NewDIDResolver()
	issuerDID := "did:example:issuer123"
	resolver.RegisterLocalKey(issuerDID, &privateKey.PublicKey)

	// Create VC claims
	vcClaims := &VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   "did:example:holder456",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "vc-12345",
		},
		VC: CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential", "NationalIDCredential"},
			CredentialSubject: map[string]interface{}{
				"id":         "did:example:holder456",
				"nationalID": "A123456789",
				"name":       "Test User",
			},
			Issuer:       issuerDID,
			IssuanceDate: time.Now().Format(time.RFC3339),
		},
	}

	// Sign VC
	vcJWT, err := SignVC(vcClaims, privateKey, issuerDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VC: %v", err)
	}

	// Validate VC
	validator := NewJWTValidator(resolver)
	validatedClaims, err := validator.ValidateVC(vcJWT)
	if err != nil {
		t.Fatalf("Failed to validate VC: %v", err)
	}

	// Verify claims
	if validatedClaims.Issuer != issuerDID {
		t.Errorf("Issuer mismatch: got %s, want %s", validatedClaims.Issuer, issuerDID)
	}

	if validatedClaims.Subject != "did:example:holder456" {
		t.Errorf("Subject mismatch: got %s, want %s", validatedClaims.Subject, "did:example:holder456")
	}

	if validatedClaims.ID != "vc-12345" {
		t.Errorf("ID mismatch: got %s, want %s", validatedClaims.ID, "vc-12345")
	}
}

func TestValidateVC_ExpiredCredential(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create DID resolver with local key
	resolver := NewDIDResolver()
	issuerDID := "did:example:issuer123"
	resolver.RegisterLocalKey(issuerDID, &privateKey.PublicKey)

	// Create expired VC claims
	vcClaims := &VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   "did:example:holder456",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ID:        "vc-12345",
		},
		VC: CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential"},
		},
	}

	// Sign VC
	vcJWT, err := SignVC(vcClaims, privateKey, issuerDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VC: %v", err)
	}

	// Validate VC - should fail
	validator := NewJWTValidator(resolver)
	_, err = validator.ValidateVC(vcJWT)
	if err == nil {
		t.Error("Expected validation to fail for expired credential")
	}
}

func TestValidateVC_InvalidSignature(t *testing.T) {
	// Generate two different key pairs
	privateKey1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	// Create DID resolver with wrong public key
	resolver := NewDIDResolver()
	issuerDID := "did:example:issuer123"
	resolver.RegisterLocalKey(issuerDID, &privateKey2.PublicKey) // Register different key

	// Create and sign VC with privateKey1
	vcClaims := &VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   "did:example:holder456",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "vc-12345",
		},
		VC: CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential"},
		},
	}

	vcJWT, err := SignVC(vcClaims, privateKey1, issuerDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VC: %v", err)
	}

	// Validate VC - should fail due to signature mismatch
	validator := NewJWTValidator(resolver)
	_, err = validator.ValidateVC(vcJWT)
	if err == nil {
		t.Error("Expected validation to fail for invalid signature")
	}
}

func TestSignAndValidateVP(t *testing.T) {
	// Generate test key pair for holder
	holderPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate holder key: %v", err)
	}

	// Generate test key pair for issuer
	issuerPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate issuer key: %v", err)
	}

	// Create DID resolver with local keys
	resolver := NewDIDResolver()
	holderDID := "did:example:holder456"
	issuerDID := "did:example:issuer123"
	resolver.RegisterLocalKey(holderDID, &holderPrivateKey.PublicKey)
	resolver.RegisterLocalKey(issuerDID, &issuerPrivateKey.PublicKey)

	// Create a VC
	vcClaims := &VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   holderDID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "vc-12345",
		},
		VC: CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential", "NationalIDCredential"},
			CredentialSubject: map[string]interface{}{
				"id":         holderDID,
				"nationalID": "A123456789",
			},
		},
	}

	vcJWT, err := SignVC(vcClaims, issuerPrivateKey, issuerDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VC: %v", err)
	}

	// Create VP claims
	nonce := "random-nonce-12345"
	audience := "did:example:verifier789"

	vpClaims := &VPClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        nonce,
			Subject:   holderDID,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VP: PresentationSubject{
			Context:              []string{"https://www.w3.org/2018/credentials/v1"},
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{vcJWT},
			Holder:               holderDID,
		},
	}

	// Sign VP
	vpJWT, err := SignVP(vpClaims, holderPrivateKey, holderDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VP: %v", err)
	}

	// Validate VP
	validator := NewJWTValidator(resolver)
	validatedVP, err := validator.ValidateVP(vpJWT, nonce, audience)
	if err != nil {
		t.Fatalf("Failed to validate VP: %v", err)
	}

	// Verify VP claims
	if validatedVP.Subject != holderDID {
		t.Errorf("Holder mismatch: got %s, want %s", validatedVP.Subject, holderDID)
	}

	if validatedVP.ID != nonce {
		t.Errorf("Nonce mismatch: got %s, want %s", validatedVP.ID, nonce)
	}

	// Verify embedded VC
	if len(validatedVP.VP.VerifiableCredential) != 1 {
		t.Fatalf("Expected 1 VC, got %d", len(validatedVP.VP.VerifiableCredential))
	}

	// Validate the embedded VC
	embeddedVC := validatedVP.VP.VerifiableCredential[0]
	validatedVC, err := validator.ValidateVC(embeddedVC)
	if err != nil {
		t.Fatalf("Failed to validate embedded VC: %v", err)
	}

	if validatedVC.Issuer != issuerDID {
		t.Errorf("VC Issuer mismatch: got %s, want %s", validatedVC.Issuer, issuerDID)
	}
}

func TestValidateVP_NonceMismatch(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create DID resolver with local key
	resolver := NewDIDResolver()
	holderDID := "did:example:holder456"
	resolver.RegisterLocalKey(holderDID, &privateKey.PublicKey)

	// Create VP with one nonce
	vpClaims := &VPClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "nonce-12345",
			Subject:   holderDID,
			Audience:  jwt.ClaimStrings{"did:example:verifier789"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VP: PresentationSubject{
			Context:              []string{"https://www.w3.org/2018/credentials/v1"},
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{},
			Holder:               holderDID,
		},
	}

	vpJWT, err := SignVP(vpClaims, privateKey, holderDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VP: %v", err)
	}

	// Validate with different nonce - should fail
	validator := NewJWTValidator(resolver)
	_, err = validator.ValidateVP(vpJWT, "different-nonce", "did:example:verifier789")
	if err == nil {
		t.Error("Expected validation to fail for nonce mismatch")
	}
}

func TestValidateVP_AudienceMismatch(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create DID resolver with local key
	resolver := NewDIDResolver()
	holderDID := "did:example:holder456"
	resolver.RegisterLocalKey(holderDID, &privateKey.PublicKey)

	// Create VP with one audience
	vpClaims := &VPClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "nonce-12345",
			Subject:   holderDID,
			Audience:  jwt.ClaimStrings{"did:example:verifier789"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VP: PresentationSubject{
			Context:              []string{"https://www.w3.org/2018/credentials/v1"},
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{},
			Holder:               holderDID,
		},
	}

	vpJWT, err := SignVP(vpClaims, privateKey, holderDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VP: %v", err)
	}

	// Validate with different audience - should fail
	validator := NewJWTValidator(resolver)
	_, err = validator.ValidateVP(vpJWT, "nonce-12345", "did:example:different-verifier")
	if err == nil {
		t.Error("Expected validation to fail for audience mismatch")
	}
}

func TestExtractDIDFromJWT(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	issuerDID := "did:example:issuer123"

	// Create VC
	vcClaims := &VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  issuerDID,
			Subject: "did:example:holder456",
		},
	}

	vcJWT, err := SignVC(vcClaims, privateKey, issuerDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VC: %v", err)
	}

	// Extract issuer DID
	extractedDID, err := ExtractDIDFromJWT(vcJWT, "iss")
	if err != nil {
		t.Fatalf("Failed to extract DID: %v", err)
	}

	if extractedDID != issuerDID {
		t.Errorf("DID mismatch: got %s, want %s", extractedDID, issuerDID)
	}

	// Extract subject DID
	subjectDID, err := ExtractDIDFromJWT(vcJWT, "sub")
	if err != nil {
		t.Fatalf("Failed to extract subject DID: %v", err)
	}

	if subjectDID != "did:example:holder456" {
		t.Errorf("Subject DID mismatch: got %s, want %s", subjectDID, "did:example:holder456")
	}
}
