package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DIDDocument represents a W3C DID Document
type DIDDocument struct {
	Context            []string                 `json:"@context"`
	ID                 string                   `json:"id"`
	VerificationMethod []VerificationMethod     `json:"verificationMethod"`
	Authentication     []interface{}            `json:"authentication,omitempty"`
	AssertionMethod    []interface{}            `json:"assertionMethod,omitempty"`
	Service            []Service                `json:"service,omitempty"`
}

// VerificationMethod represents a verification method in a DID Document
type VerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyJwk       *JWK   `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
	PublicKeyBase58    string `json:"publicKeyBase58,omitempty"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type
	Crv string `json:"crv"` // Curve
	X   string `json:"x"`   // X coordinate
	Y   string `json:"y"`   // Y coordinate
	Use string `json:"use,omitempty"`
	Kid string `json:"kid,omitempty"`
}

// Service represents a service in a DID Document
type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// DIDResolver resolves DIDs to public keys
type DIDResolver struct {
	// Cache for resolved keys
	cache map[string]cachedKey
	mu    sync.RWMutex

	// HTTP client for remote resolution
	httpClient *http.Client

	// Local key store for testing
	localKeys map[string]interface{}
}

type cachedKey struct {
	key       interface{}
	expiresAt time.Time
}

// NewDIDResolver creates a new DID resolver
func NewDIDResolver() *DIDResolver {
	return &DIDResolver{
		cache: make(map[string]cachedKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		localKeys: make(map[string]interface{}),
	}
}

// ResolveKey resolves a DID to its public key
func (r *DIDResolver) ResolveKey(did string) (interface{}, error) {
	// Check cache first
	r.mu.RLock()
	if cached, ok := r.cache[did]; ok && time.Now().Before(cached.expiresAt) {
		r.mu.RUnlock()
		return cached.key, nil
	}
	r.mu.RUnlock()

	// Check local keys (for testing) - also cache these
	r.mu.RLock()
	if key, ok := r.localKeys[did]; ok {
		r.mu.RUnlock()
		// Cache the local key
		r.mu.Lock()
		r.cache[did] = cachedKey{
			key:       key,
			expiresAt: time.Now().Add(30 * time.Minute),
		}
		r.mu.Unlock()
		return key, nil
	}
	r.mu.RUnlock()

	// Resolve based on DID method
	var key interface{}
	var err error

	if strings.HasPrefix(did, "did:web:") {
		key, err = r.resolveWebDID(did)
	} else if strings.HasPrefix(did, "did:key:") {
		key, err = r.resolveKeyDID(did)
	} else if strings.HasPrefix(did, "did:example:") {
		// For testing - use a default key
		key, err = r.resolveExampleDID(did)
	} else {
		return nil, fmt.Errorf("unsupported DID method: %s", did)
	}

	if err != nil {
		return nil, err
	}

	// Cache the resolved key (30 minutes)
	r.mu.Lock()
	r.cache[did] = cachedKey{
		key:       key,
		expiresAt: time.Now().Add(30 * time.Minute),
	}
	r.mu.Unlock()

	return key, nil
}

// RegisterLocalKey registers a local key for testing
func (r *DIDResolver) RegisterLocalKey(did string, publicKey interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.localKeys[did] = publicKey
}

// resolveWebDID resolves a did:web DID
func (r *DIDResolver) resolveWebDID(did string) (interface{}, error) {
	// Convert did:web:example.com to https://example.com/.well-known/did.json
	didParts := strings.Split(did, ":")
	if len(didParts) < 3 {
		return nil, fmt.Errorf("invalid did:web format")
	}

	domain := strings.Join(didParts[2:], ":")
	url := fmt.Sprintf("https://%s/.well-known/did.json", domain)

	// Fetch DID document
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DID document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch DID document: status %d", resp.StatusCode)
	}

	var didDoc DIDDocument
	if err := json.NewDecoder(resp.Body).Decode(&didDoc); err != nil {
		return nil, fmt.Errorf("failed to parse DID document: %w", err)
	}

	return r.extractPublicKey(&didDoc)
}

// resolveKeyDID resolves a did:key DID
func (r *DIDResolver) resolveKeyDID(did string) (interface{}, error) {
	// did:key uses multibase encoding
	// For now, return error as full implementation requires multicodec
	return nil, fmt.Errorf("did:key resolution not yet implemented")
}

// resolveExampleDID resolves a did:example DID (for testing)
func (r *DIDResolver) resolveExampleDID(did string) (interface{}, error) {
	// For testing purposes, generate a deterministic key based on DID
	// In production, this should never be used
	// Generate a simple ECDSA P-256 public key for testing
	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult([]byte(did))

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// extractPublicKey extracts the public key from a DID document
func (r *DIDResolver) extractPublicKey(didDoc *DIDDocument) (interface{}, error) {
	if len(didDoc.VerificationMethod) == 0 {
		return nil, fmt.Errorf("no verification methods found")
	}

	// Use the first verification method
	vm := didDoc.VerificationMethod[0]

	// Try JWK format first
	if vm.PublicKeyJwk != nil {
		return r.jwkToPublicKey(vm.PublicKeyJwk)
	}

	// Try multibase format
	if vm.PublicKeyMultibase != "" {
		return r.multibaseToPublicKey(vm.PublicKeyMultibase)
	}

	// Try base58 format
	if vm.PublicKeyBase58 != "" {
		return r.base58ToPublicKey(vm.PublicKeyBase58)
	}

	return nil, fmt.Errorf("no supported public key format found")
}

// jwkToPublicKey converts a JWK to a public key
func (r *DIDResolver) jwkToPublicKey(jwk *JWK) (interface{}, error) {
	if jwk.Kty != "EC" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	// Decode X and Y coordinates
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y coordinate: %w", err)
	}

	// Determine curve
	var curve elliptic.Curve
	switch jwk.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", jwk.Crv)
	}

	// Create public key
	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return pubKey, nil
}

// multibaseToPublicKey converts multibase-encoded key to public key
func (r *DIDResolver) multibaseToPublicKey(multibase string) (interface{}, error) {
	// Full implementation requires multicodec library
	return nil, fmt.Errorf("multibase decoding not yet implemented")
}

// base58ToPublicKey converts base58-encoded key to public key
func (r *DIDResolver) base58ToPublicKey(base58 string) (interface{}, error) {
	// Full implementation requires base58 library
	return nil, fmt.Errorf("base58 decoding not yet implemented")
}

// ClearCache clears the DID resolution cache
func (r *DIDResolver) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]cachedKey)
}

// parseDERPublicKey parses a DER-encoded public key
func parseDERPublicKey(derBytes []byte) (interface{}, error) {
	// Try parsing as PKIX public key
	if key, err := x509.ParsePKIXPublicKey(derBytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse DER public key")
}
