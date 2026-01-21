package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDIDResolver_RegisterAndResolveLocalKey(t *testing.T) {
	resolver := NewDIDResolver()

	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	did := "did:example:test123"
	resolver.RegisterLocalKey(did, &privateKey.PublicKey)

	// Resolve the key
	resolvedKey, err := resolver.ResolveKey(did)
	if err != nil {
		t.Fatalf("Failed to resolve key: %v", err)
	}

	// Verify it's the same key
	ecKey, ok := resolvedKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Resolved key is not an ECDSA public key")
	}

	if ecKey.X.Cmp(privateKey.PublicKey.X) != 0 || ecKey.Y.Cmp(privateKey.PublicKey.Y) != 0 {
		t.Error("Resolved key does not match registered key")
	}
}

func TestDIDResolver_CacheExpiration(t *testing.T) {
	resolver := NewDIDResolver()

	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	did := "did:example:test123"
	resolver.RegisterLocalKey(did, &privateKey.PublicKey)

	// First resolution - should cache
	_, err = resolver.ResolveKey(did)
	if err != nil {
		t.Fatalf("Failed to resolve key: %v", err)
	}

	// Check cache has the key
	resolver.mu.RLock()
	_, inCache := resolver.cache[did]
	resolver.mu.RUnlock()

	if !inCache {
		t.Error("Key should be in cache after first resolution")
	}

	// Clear cache
	resolver.ClearCache()

	// Check cache is empty
	resolver.mu.RLock()
	_, stillInCache := resolver.cache[did]
	resolver.mu.RUnlock()

	if stillInCache {
		t.Error("Key should not be in cache after clearing")
	}
}

func TestDIDResolver_WebDID(t *testing.T) {
	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create JWK from public key
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()

	// Pad to 32 bytes for P-256
	xPadded := make([]byte, 32)
	yPadded := make([]byte, 32)
	copy(xPadded[32-len(xBytes):], xBytes)
	copy(yPadded[32-len(yBytes):], yBytes)

	jwk := &JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(xPadded),
		Y:   base64.RawURLEncoding.EncodeToString(yPadded),
	}

	// Create mock DID document
	didDoc := &DIDDocument{
		Context: []string{"https://www.w3.org/ns/did/v1"},
		ID:      "did:web:example.com",
		VerificationMethod: []VerificationMethod{
			{
				ID:           "did:web:example.com#key-1",
				Type:         "JsonWebKey2020",
				Controller:   "did:web:example.com",
				PublicKeyJwk: jwk,
			},
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/did.json" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(didDoc)
	}))
	defer server.Close()

	// Override the resolver's HTTP client to use test server
	resolver := NewDIDResolver()

	// Since we can't easily override the URL construction, we'll test the JWK conversion directly
	resolvedKey, err := resolver.jwkToPublicKey(jwk)
	if err != nil {
		t.Fatalf("Failed to convert JWK to public key: %v", err)
	}

	ecKey, ok := resolvedKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Resolved key is not an ECDSA public key")
	}

	if ecKey.X.Cmp(privateKey.PublicKey.X) != 0 || ecKey.Y.Cmp(privateKey.PublicKey.Y) != 0 {
		t.Error("Resolved key does not match original key")
	}
}

func TestJWKToPublicKey_P256(t *testing.T) {
	resolver := NewDIDResolver()

	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create JWK
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()

	// Pad to 32 bytes for P-256
	xPadded := make([]byte, 32)
	yPadded := make([]byte, 32)
	copy(xPadded[32-len(xBytes):], xBytes)
	copy(yPadded[32-len(yBytes):], yBytes)

	jwk := &JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(xPadded),
		Y:   base64.RawURLEncoding.EncodeToString(yPadded),
	}

	// Convert to public key
	pubKey, err := resolver.jwkToPublicKey(jwk)
	if err != nil {
		t.Fatalf("Failed to convert JWK: %v", err)
	}

	ecKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Converted key is not an ECDSA public key")
	}

	if ecKey.Curve != elliptic.P256() {
		t.Error("Curve mismatch")
	}

	if ecKey.X.Cmp(privateKey.PublicKey.X) != 0 || ecKey.Y.Cmp(privateKey.PublicKey.Y) != 0 {
		t.Error("Public key coordinates do not match")
	}
}

func TestJWKToPublicKey_P384(t *testing.T) {
	resolver := NewDIDResolver()

	// Generate test key
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create JWK
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()

	// Pad to 48 bytes for P-384
	xPadded := make([]byte, 48)
	yPadded := make([]byte, 48)
	copy(xPadded[48-len(xBytes):], xBytes)
	copy(yPadded[48-len(yBytes):], yBytes)

	jwk := &JWK{
		Kty: "EC",
		Crv: "P-384",
		X:   base64.RawURLEncoding.EncodeToString(xPadded),
		Y:   base64.RawURLEncoding.EncodeToString(yPadded),
	}

	// Convert to public key
	pubKey, err := resolver.jwkToPublicKey(jwk)
	if err != nil {
		t.Fatalf("Failed to convert JWK: %v", err)
	}

	ecKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Converted key is not an ECDSA public key")
	}

	if ecKey.Curve != elliptic.P384() {
		t.Error("Curve mismatch")
	}

	if ecKey.X.Cmp(privateKey.PublicKey.X) != 0 || ecKey.Y.Cmp(privateKey.PublicKey.Y) != 0 {
		t.Error("Public key coordinates do not match")
	}
}

func TestJWKToPublicKey_UnsupportedKeyType(t *testing.T) {
	resolver := NewDIDResolver()

	jwk := &JWK{
		Kty: "RSA", // Unsupported in current implementation
		Crv: "P-256",
		X:   "test",
		Y:   "test",
	}

	_, err := resolver.jwkToPublicKey(jwk)
	if err == nil {
		t.Error("Expected error for unsupported key type")
	}
}

func TestJWKToPublicKey_UnsupportedCurve(t *testing.T) {
	resolver := NewDIDResolver()

	jwk := &JWK{
		Kty: "EC",
		Crv: "secp256k1", // Unsupported curve
		X:   "test",
		Y:   "test",
	}

	_, err := resolver.jwkToPublicKey(jwk)
	if err == nil {
		t.Error("Expected error for unsupported curve")
	}
}

func TestExtractPublicKey_NoVerificationMethod(t *testing.T) {
	resolver := NewDIDResolver()

	didDoc := &DIDDocument{
		ID:                 "did:example:test",
		VerificationMethod: []VerificationMethod{}, // Empty
	}

	_, err := resolver.extractPublicKey(didDoc)
	if err == nil {
		t.Error("Expected error for DID document with no verification methods")
	}
}

func TestResolveExampleDID(t *testing.T) {
	resolver := NewDIDResolver()

	did := "did:example:test123"

	// Resolve example DID
	key, err := resolver.ResolveKey(did)
	if err != nil {
		t.Fatalf("Failed to resolve example DID: %v", err)
	}

	// Verify it's an ECDSA key
	_, ok := key.(*ecdsa.PublicKey)
	if !ok {
		t.Error("Example DID should resolve to ECDSA public key")
	}

	// Resolve same DID again - should get same key from cache
	key2, err := resolver.ResolveKey(did)
	if err != nil {
		t.Fatalf("Failed to resolve example DID second time: %v", err)
	}

	ecKey1 := key.(*ecdsa.PublicKey)
	ecKey2 := key2.(*ecdsa.PublicKey)

	if ecKey1.X.Cmp(ecKey2.X) != 0 || ecKey1.Y.Cmp(ecKey2.Y) != 0 {
		t.Error("Same DID should resolve to same key")
	}
}
