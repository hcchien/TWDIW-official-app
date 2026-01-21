package credential

import (
	"strings"
	"testing"
)

func TestGenerateOpaqueIDSeed(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	tests := []struct {
		name           string
		holderUID      string
		credentialType string
		wantErr        bool
	}{
		{
			name:           "Valid holder and type",
			holderUID:      "holder123",
			credentialType: "age_verification",
			wantErr:        false,
		},
		{
			name:           "Empty holder UID",
			holderUID:      "",
			credentialType: "age_verification",
			wantErr:        false, // Should still work, just uses empty string as key
		},
		{
			name:           "Empty credential type",
			holderUID:      "holder123",
			credentialType: "",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, err := registry.GenerateOpaqueIDSeed(tt.holderUID, tt.credentialType)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateOpaqueIDSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify seed format
				if len(seed) != 43 {
					t.Errorf("GenerateOpaqueIDSeed() seed length = %d, want 43", len(seed))
				}

				// Verify it's base64url encoded (should not contain + or / or =)
				if strings.ContainsAny(seed, "+/=") {
					t.Errorf("GenerateOpaqueIDSeed() seed contains non-base64url characters: %s", seed)
				}

				// Verify validation passes
				if err := ValidateOpaqueIDSeed(seed); err != nil {
					t.Errorf("ValidateOpaqueIDSeed() failed for generated seed: %v", err)
				}
			}
		})
	}
}

func TestGenerateOpaqueIDSeed_Stability(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	holderUID := "holder123"
	credentialType := "age_verification"

	// Generate seed first time
	seed1, err := registry.GenerateOpaqueIDSeed(holderUID, credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() first call error = %v", err)
	}

	// Generate seed second time (should be same)
	seed2, err := registry.GenerateOpaqueIDSeed(holderUID, credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() second call error = %v", err)
	}

	if seed1 != seed2 {
		t.Errorf("GenerateOpaqueIDSeed() not stable: first=%s, second=%s", seed1, seed2)
	}
}

func TestGenerateOpaqueIDSeed_UniquenessAcrossHolders(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	credentialType := "age_verification"

	seed1, err := registry.GenerateOpaqueIDSeed("holder1", credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() holder1 error = %v", err)
	}

	seed2, err := registry.GenerateOpaqueIDSeed("holder2", credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() holder2 error = %v", err)
	}

	if seed1 == seed2 {
		t.Errorf("GenerateOpaqueIDSeed() generated same seed for different holders")
	}
}

func TestGenerateOpaqueIDSeed_UniquenessAcrossTypes(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	holderUID := "holder123"

	seed1, err := registry.GenerateOpaqueIDSeed(holderUID, "age_verification")
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() type1 error = %v", err)
	}

	seed2, err := registry.GenerateOpaqueIDSeed(holderUID, "driver_license")
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() type2 error = %v", err)
	}

	if seed1 == seed2 {
		t.Errorf("GenerateOpaqueIDSeed() generated same seed for different credential types")
	}
}

func TestGetSeed(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	holderUID := "holder123"
	credentialType := "age_verification"

	// Seed doesn't exist initially
	_, exists := registry.GetSeed(holderUID, credentialType)
	if exists {
		t.Error("GetSeed() found seed that shouldn't exist")
	}

	// Generate seed
	expectedSeed, err := registry.GenerateOpaqueIDSeed(holderUID, credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() error = %v", err)
	}

	// Now it should exist
	retrievedSeed, exists := registry.GetSeed(holderUID, credentialType)
	if !exists {
		t.Error("GetSeed() didn't find seed that should exist")
	}

	if retrievedSeed != expectedSeed {
		t.Errorf("GetSeed() = %s, want %s", retrievedSeed, expectedSeed)
	}
}

func TestRevokeSeed(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	holderUID := "holder123"
	credentialType := "age_verification"

	// Generate seed
	_, err := registry.GenerateOpaqueIDSeed(holderUID, credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() error = %v", err)
	}

	// Verify it exists
	_, exists := registry.GetSeed(holderUID, credentialType)
	if !exists {
		t.Error("GetSeed() didn't find seed before revocation")
	}

	// Revoke seed
	registry.RevokeSeed(holderUID, credentialType)

	// Verify it no longer exists
	_, exists = registry.GetSeed(holderUID, credentialType)
	if exists {
		t.Error("GetSeed() still found seed after revocation")
	}
}

func TestInjectOpaqueIDSeed(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	tests := []struct {
		name               string
		credentialSubject  map[string]interface{}
		holderUID          string
		credentialType     string
		wantErr            bool
		wantFieldCount     int
	}{
		{
			name: "Inject into empty subject",
			credentialSubject: map[string]interface{}{},
			holderUID:     "holder123",
			credentialType: "age_verification",
			wantErr:       false,
			wantFieldCount: 1, // Only opaque_id_seed
		},
		{
			name: "Inject into subject with existing fields",
			credentialSubject: map[string]interface{}{
				"over_18": true,
				"over_21": false,
			},
			holderUID:      "holder123",
			credentialType: "age_verification",
			wantErr:        false,
			wantFieldCount: 3, // over_18, over_21, opaque_id_seed
		},
		{
			name: "Inject with empty holder UID",
			credentialSubject: map[string]interface{}{
				"over_18": true,
			},
			holderUID:      "",
			credentialType: "age_verification",
			wantErr:        false,
			wantFieldCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := registry.InjectOpaqueIDSeed(
				tt.credentialSubject,
				tt.holderUID,
				tt.credentialType,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("InjectOpaqueIDSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify opaque_id_seed was added
				seed, ok := result["opaque_id_seed"]
				if !ok {
					t.Error("InjectOpaqueIDSeed() didn't add opaque_id_seed field")
					return
				}

				seedStr, ok := seed.(string)
				if !ok {
					t.Errorf("InjectOpaqueIDSeed() opaque_id_seed is not string, got type %T", seed)
					return
				}

				// Verify seed format
				if err := ValidateOpaqueIDSeed(seedStr); err != nil {
					t.Errorf("InjectOpaqueIDSeed() generated invalid seed: %v", err)
				}

				// Verify field count
				if len(result) != tt.wantFieldCount {
					t.Errorf("InjectOpaqueIDSeed() result has %d fields, want %d", len(result), tt.wantFieldCount)
				}

				// Verify original fields are preserved
				for key, value := range tt.credentialSubject {
					if result[key] != value {
						t.Errorf("InjectOpaqueIDSeed() didn't preserve field %s", key)
					}
				}

				// Verify original map wasn't modified
				if _, exists := tt.credentialSubject["opaque_id_seed"]; exists {
					t.Error("InjectOpaqueIDSeed() modified input map")
				}
			}
		})
	}
}

func TestValidateOpaqueIDSeed(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	// Generate valid seed
	validSeed, err := registry.GenerateOpaqueIDSeed("holder123", "age_verification")
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() error = %v", err)
	}

	tests := []struct {
		name    string
		seed    string
		wantErr bool
	}{
		{
			name:    "Valid seed",
			seed:    validSeed,
			wantErr: false,
		},
		{
			name:    "Invalid base64url",
			seed:    "invalid!@#$%^&*()",
			wantErr: true,
		},
		{
			name:    "Wrong length (too short)",
			seed:    "YWJjZGVm", // Only 6 bytes
			wantErr: true,
		},
		{
			name:    "Empty seed",
			seed:    "",
			wantErr: true,
		},
		{
			name:    "Base64 with padding (not base64url)",
			seed:    "a3d7f9c8b2e1a4f6d8c9b7e2a5f8d3c1b4e7a9f2d6c8b3e5a7f9d2c4b6e8a1f3==",
			wantErr: true, // Padding should fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOpaqueIDSeed(tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOpaqueIDSeed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpaqueIDSeedRegistry_Concurrent(t *testing.T) {
	registry := NewOpaqueIDSeedRegistry()

	holderUID := "holder123"
	credentialType := "age_verification"

	// Generate seed once
	expectedSeed, err := registry.GenerateOpaqueIDSeed(holderUID, credentialType)
	if err != nil {
		t.Fatalf("GenerateOpaqueIDSeed() error = %v", err)
	}

	// Concurrent access
	const numGoroutines = 100
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			seed, err := registry.GenerateOpaqueIDSeed(holderUID, credentialType)
			if err != nil {
				errors <- err
				return
			}
			results <- seed
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent GenerateOpaqueIDSeed() error: %v", err)
		case seed := <-results:
			if seed != expectedSeed {
				t.Errorf("Concurrent GenerateOpaqueIDSeed() got different seed: %s, want %s", seed, expectedSeed)
			}
		}
	}
}
