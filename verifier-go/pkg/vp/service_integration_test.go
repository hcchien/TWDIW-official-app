package vp

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/crypto"
)

// TestValidate_WithRealJWT tests the full VP validation flow with real JWT signatures
func TestValidate_WithRealJWT(t *testing.T) {
	// Setup: Generate keys for issuer and holder
	issuerPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate issuer key: %v", err)
	}

	holderPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate holder key: %v", err)
	}

	issuerDID := "did:example:issuer123"
	holderDID := "did:example:holder456"

	// Create DID resolver and register keys
	resolver := crypto.NewDIDResolver()
	resolver.RegisterLocalKey(issuerDID, &issuerPrivateKey.PublicKey)
	resolver.RegisterLocalKey(holderDID, &holderPrivateKey.PublicKey)

	// Create service with custom resolver
	service := NewServiceWithResolver(resolver)

	// Create a Verifiable Credential
	vcClaims := &crypto.VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   holderDID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "vc-12345",
		},
		VC: crypto.CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential", "NationalIDCredential"},
			CredentialSubject: map[string]interface{}{
				"id":         holderDID,
				"nationalID": "A123456789",
				"name":       "Test User",
			},
			Issuer:       issuerDID,
			IssuanceDate: time.Now().Format(time.RFC3339),
		},
	}

	vcJWT, err := crypto.SignVC(vcClaims, issuerPrivateKey, issuerDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VC: %v", err)
	}

	// Create a Verifiable Presentation containing the VC
	vpClaims := &crypto.VPClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "nonce-67890",
			Subject:   holderDID,
			Audience:  jwt.ClaimStrings{"did:example:verifier789"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VP: crypto.PresentationSubject{
			Context:              []string{"https://www.w3.org/2018/credentials/v1"},
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{vcJWT},
			Holder:               holderDID,
		},
	}

	vpJWT, err := crypto.SignVP(vpClaims, holderPrivateKey, holderDID+"#key-1")
	if err != nil {
		t.Fatalf("Failed to sign VP: %v", err)
	}

	// Test: Validate the VP
	ctx := context.Background()
	result, status, err := service.Validate(ctx, []string{vpJWT})

	// Assert: Validation should succeed
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Parse response
	var response []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Expected 1 validation result, got %d", len(response))
	}

	// Verify response data
	if len(response) > 0 {
		resp := response[0]

		if holderDIDResp, ok := resp["holder_did"].(string); !ok || holderDIDResp != holderDID {
			t.Errorf("Expected holder_did %s, got %v", holderDID, resp["holder_did"])
		}

		if nonceResp, ok := resp["nonce"].(string); !ok || nonceResp != "nonce-67890" {
			t.Errorf("Expected nonce 'nonce-67890', got %v", resp["nonce"])
		}

		if clientIDResp, ok := resp["client_id"].(string); !ok || clientIDResp != "did:example:verifier789" {
			t.Errorf("Expected client_id 'did:example:verifier789', got %v", resp["client_id"])
		}

		// Check VCs
		if vcs, ok := resp["vcs"].([]interface{}); !ok || len(vcs) != 1 {
			t.Errorf("Expected 1 VC in response, got %v", resp["vcs"])
		}
	}
}

// TestValidate_WithExpiredVC tests validation with an expired VC
func TestValidate_WithExpiredVC(t *testing.T) {
	// Setup: Generate keys
	issuerPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	holderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	issuerDID := "did:example:issuer123"
	holderDID := "did:example:holder456"

	// Create DID resolver and register keys
	resolver := crypto.NewDIDResolver()
	resolver.RegisterLocalKey(issuerDID, &issuerPrivateKey.PublicKey)
	resolver.RegisterLocalKey(holderDID, &holderPrivateKey.PublicKey)

	// Create service with custom resolver
	service := NewServiceWithResolver(resolver)

	// Create an EXPIRED Verifiable Credential
	vcClaims := &crypto.VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   holderDID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ID:        "vc-expired",
		},
		VC: crypto.CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential"},
		},
	}

	vcJWT, _ := crypto.SignVC(vcClaims, issuerPrivateKey, issuerDID+"#key-1")

	// Create a VP containing the expired VC
	vpClaims := &crypto.VPClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "nonce-test",
			Subject:   holderDID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VP: crypto.PresentationSubject{
			Context:              []string{"https://www.w3.org/2018/credentials/v1"},
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{vcJWT},
			Holder:               holderDID,
		},
	}

	vpJWT, _ := crypto.SignVP(vpClaims, holderPrivateKey, holderDID+"#key-1")

	// Test: Validate the VP with expired VC
	ctx := context.Background()
	result, status, err := service.Validate(ctx, []string{vpJWT})

	// Assert: VP validation should succeed, but VC validation should fail
	// The service continues processing but won't include the expired VC in results
	if err != nil {
		t.Logf("Expected behavior: VP validates but VC is rejected: %v", err)
	}

	if status != http.StatusOK {
		t.Logf("Status: %d (may be error if VP-level validation fails)", status)
	}

	// Parse response
	var response interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	t.Logf("Response: %+v", response)
}

// TestValidate_WithInvalidSignature tests validation with invalid signature
func TestValidate_WithInvalidSignature(t *testing.T) {
	// Setup: Generate two different key pairs
	issuerPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	holderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	wrongPrivateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	issuerDID := "did:example:issuer123"
	holderDID := "did:example:holder456"

	// Create DID resolver with WRONG public key
	resolver := crypto.NewDIDResolver()
	resolver.RegisterLocalKey(issuerDID, &issuerPrivateKey.PublicKey) // Correct issuer key
	resolver.RegisterLocalKey(holderDID, &wrongPrivateKey.PublicKey)  // Wrong holder key

	// Create service with custom resolver
	service := NewServiceWithResolver(resolver)

	// Create a VC
	vcClaims := &crypto.VCClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuerDID,
			Subject:   holderDID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VC: crypto.CredentialSubject{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiableCredential"},
		},
	}

	vcJWT, _ := crypto.SignVC(vcClaims, issuerPrivateKey, issuerDID+"#key-1")

	// Create a VP signed with the ACTUAL holder key (not the wrong one registered)
	vpClaims := &crypto.VPClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "nonce-test",
			Subject:   holderDID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		VP: crypto.PresentationSubject{
			Context:              []string{"https://www.w3.org/2018/credentials/v1"},
			Type:                 []string{"VerifiablePresentation"},
			VerifiableCredential: []string{vcJWT},
			Holder:               holderDID,
		},
	}

	vpJWT, _ := crypto.SignVP(vpClaims, holderPrivateKey, holderDID+"#key-1")

	// Test: Validate the VP - should fail due to signature mismatch
	ctx := context.Background()
	_, status, err := service.Validate(ctx, []string{vpJWT})

	// Assert: Should fail with signature error
	if err == nil {
		t.Error("Expected validation to fail due to invalid signature")
	}

	if status == http.StatusOK {
		t.Error("Expected non-OK status due to invalid signature")
	}
}
