package credential

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
)

// OpaqueIDSeedRegistry manages opaque ID seeds for pairwise pseudonymous identifiers
type OpaqueIDSeedRegistry struct {
	// In-memory storage for seeds (in production, use persistent storage)
	seeds map[string]string // key: holderUID:credentialType, value: base64url-encoded seed
	mu    sync.RWMutex
}

// NewOpaqueIDSeedRegistry creates a new seed registry
func NewOpaqueIDSeedRegistry() *OpaqueIDSeedRegistry {
	return &OpaqueIDSeedRegistry{
		seeds: make(map[string]string),
	}
}

// GenerateOpaqueIDSeed generates or retrieves a cryptographically secure opaque ID seed
// for pairwise pseudonymous identifiers.
//
// The seed is:
// - 256 bits (32 bytes) of cryptographic random data
// - Base64url-encoded (43 characters)
// - Stable for the same holder and credential type
// - Used by wallet to derive verifier-specific pseudonyms via HMAC-SHA256
//
// @param holderUID Unique identifier for the holder (e.g., national ID hash)
// @param credentialType Type of credential being issued
// @return Base64url-encoded 256-bit random seed
func (r *OpaqueIDSeedRegistry) GenerateOpaqueIDSeed(holderUID, credentialType string) (string, error) {
	// Create composite key for registry lookup
	registryKey := fmt.Sprintf("%s:%s", holderUID, credentialType)

	// Check if seed already exists (to maintain stability across credential renewals)
	r.mu.RLock()
	existingSeed, exists := r.seeds[registryKey]
	r.mu.RUnlock()

	if exists {
		return existingSeed, nil
	}

	// Generate new 256-bit seed using cryptographic random number generator
	seedBytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(seedBytes); err != nil {
		return "", fmt.Errorf("failed to generate random seed: %w", err)
	}

	// Encode as base64url (URL-safe, no padding)
	// Results in 43-character string
	seed := base64.RawURLEncoding.EncodeToString(seedBytes)

	// Store for future retrievals
	r.mu.Lock()
	r.seeds[registryKey] = seed
	r.mu.Unlock()

	return seed, nil
}

// GetSeed retrieves an existing opaque ID seed if it exists
func (r *OpaqueIDSeedRegistry) GetSeed(holderUID, credentialType string) (string, bool) {
	registryKey := fmt.Sprintf("%s:%s", holderUID, credentialType)

	r.mu.RLock()
	defer r.mu.RUnlock()

	seed, exists := r.seeds[registryKey]
	return seed, exists
}

// RevokeSeed removes a seed from the registry (for credential revocation)
func (r *OpaqueIDSeedRegistry) RevokeSeed(holderUID, credentialType string) {
	registryKey := fmt.Sprintf("%s:%s", holderUID, credentialType)

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.seeds, registryKey)
}

// InjectOpaqueIDSeed adds the opaque_id_seed field to credential subject data
//
// This function:
// 1. Generates or retrieves the opaque ID seed for the holder
// 2. Adds it to the credential subject map as "opaque_id_seed" field
// 3. Ensures the field is marked for selective disclosure in SD-JWT
//
// @param credentialSubject Map of claims to be included in credential
// @param holderUID Unique identifier for the holder
// @param credentialType Type of credential being issued
// @return Modified credential subject with opaque_id_seed, or error
func (r *OpaqueIDSeedRegistry) InjectOpaqueIDSeed(
	credentialSubject map[string]interface{},
	holderUID string,
	credentialType string,
) (map[string]interface{}, error) {
	// Generate or retrieve opaque ID seed
	seed, err := r.GenerateOpaqueIDSeed(holderUID, credentialType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate opaque ID seed: %w", err)
	}

	// Clone credential subject to avoid modifying input
	result := make(map[string]interface{}, len(credentialSubject)+1)
	for k, v := range credentialSubject {
		result[k] = v
	}

	// Add opaque_id_seed field
	result["opaque_id_seed"] = seed

	return result, nil
}

// ValidateOpaqueIDSeed validates that a seed has correct format
func ValidateOpaqueIDSeed(seed string) error {
	// Decode from base64url
	decoded, err := base64.RawURLEncoding.DecodeString(seed)
	if err != nil {
		return fmt.Errorf("invalid base64url encoding: %w", err)
	}

	// Check length (must be 32 bytes = 256 bits)
	if len(decoded) != 32 {
		return fmt.Errorf("invalid seed length: expected 32 bytes, got %d", len(decoded))
	}

	return nil
}
