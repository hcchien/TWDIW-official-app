# Cryptographic Validation Package

This package provides cryptographic validation for Verifiable Credentials (VCs) and Verifiable Presentations (VPs) using JWT signatures and DID-based key resolution.

## Features

- **JWT Validation**: Full JWT signature validation for VCs and VPs
- **DID Resolution**: Resolve DIDs to public keys for signature verification
- **Multiple Signature Algorithms**: Support for ECDSA (P-256, P-384, P-521), RSA, and EdDSA
- **Caching**: Efficient DID resolution with 30-minute cache
- **Security**: Protection against expired credentials, invalid signatures, and nonce/audience mismatches

## Components

### JWT Validator (`jwt.go`)

The JWT validator handles parsing and validating JWT-based VCs and VPs:

```go
import "github.com/moda-gov-tw/twdiw-verifier-go/pkg/crypto"

// Create a DID resolver
resolver := crypto.NewDIDResolver()

// Create a JWT validator
validator := crypto.NewJWTValidator(resolver)

// Validate a VC
vcClaims, err := validator.ValidateVC(vcJWT)
if err != nil {
    // Handle validation error
}

// Validate a VP with nonce and audience check
vpClaims, err := validator.ValidateVP(vpJWT, expectedNonce, expectedAudience)
if err != nil {
    // Handle validation error
}
```

#### VC Validation

The `ValidateVC` method performs the following checks:

1. **JWT Parsing**: Parses the JWT and extracts claims
2. **Issuer Resolution**: Resolves the issuer's DID to get their public key
3. **Signature Verification**: Verifies the JWT signature using the issuer's public key
4. **Expiration Check**: Validates the credential hasn't expired (both `exp` claim and `expirationDate`)
5. **Not-Before Check**: Ensures the credential is currently valid (`nbf` claim)

#### VP Validation

The `ValidateVP` method performs additional checks:

1. **All VC Validation Checks** (above)
2. **Holder Resolution**: Resolves the holder's DID to get their public key
3. **Nonce Verification**: Ensures the nonce matches the expected value (prevents replay attacks)
4. **Audience Verification**: Checks the VP is intended for the expected verifier

### DID Resolver (`did_resolver.go`)

The DID resolver translates DIDs into public keys for signature verification:

```go
resolver := crypto.NewDIDResolver()

// Resolve a DID to its public key
publicKey, err := resolver.ResolveKey("did:web:example.com")
if err != nil {
    // Handle resolution error
}

// For testing: Register local keys
resolver.RegisterLocalKey("did:example:test123", publicKey)

// Clear the cache
resolver.ClearCache()
```

#### Supported DID Methods

- **did:web**: Fetches DID documents from `https://<domain>/.well-known/did.json`
- **did:example**: For testing purposes (generates deterministic keys)
- **Local Keys**: Register keys manually for testing

#### DID Document Format

The resolver expects W3C DID Document format with `verificationMethod`:

```json
{
  "@context": ["https://www.w3.org/ns/did/v1"],
  "id": "did:web:example.com",
  "verificationMethod": [
    {
      "id": "did:web:example.com#key-1",
      "type": "JsonWebKey2020",
      "controller": "did:web:example.com",
      "publicKeyJwk": {
        "kty": "EC",
        "crv": "P-256",
        "x": "...",
        "y": "..."
      }
    }
  ]
}
```

## Signing Credentials and Presentations

The package also provides functions to create signed JWTs:

```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
)

// Generate a key pair
privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

// Create VC claims
vcClaims := &crypto.VCClaims{
    RegisteredClaims: jwt.RegisteredClaims{
        Issuer:    "did:example:issuer",
        Subject:   "did:example:holder",
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        ID:        "vc-12345",
    },
    VC: crypto.CredentialSubject{
        Context: []string{"https://www.w3.org/2018/credentials/v1"},
        Type:    []string{"VerifiableCredential", "NationalIDCredential"},
        CredentialSubject: map[string]interface{}{
            "id": "did:example:holder",
            "nationalID": "A123456789",
        },
    },
}

// Sign the VC
vcJWT, err := crypto.SignVC(vcClaims, privateKey, "did:example:issuer#key-1")

// Sign a VP
vpJWT, err := crypto.SignVP(vpClaims, privateKey, "did:example:holder#key-1")
```

## Security Features

### Expiration Validation

Credentials are checked for expiration in two ways:

1. **JWT `exp` claim**: Standard JWT expiration
2. **VC `expirationDate` field**: W3C VC-specific expiration

Both must be valid for the credential to pass validation.

### Nonce and Audience Protection

VPs include nonce and audience claims to prevent:

- **Replay Attacks**: The nonce ensures each presentation is unique
- **Impersonation**: The audience ensures the VP is intended for the right verifier

```go
// Validate with strict nonce and audience checks
vpClaims, err := validator.ValidateVP(
    vpJWT,
    "expected-nonce-12345",           // Must match jti claim
    "did:example:verifier789",        // Must be in aud claim
)
```

### Key Consistency

When validating VPs, the service ensures:

1. The VP is signed by the holder's private key
2. Each embedded VC's subject matches the VP's holder
3. Each VC is signed by its issuer's private key

This prevents credential theft and ensures the presenter actually owns the credentials.

## Error Handling

The package returns descriptive errors for various failure scenarios:

```go
vcClaims, err := validator.ValidateVC(vcJWT)
if err != nil {
    // Errors can include:
    // - "JWT validation failed": Signature verification failed
    // - "credential has expired": VC is no longer valid
    // - "failed to resolve issuer key": DID resolution failed
    // - "unsupported signing method": Algorithm not supported
}
```

## Caching

DID resolution results are cached for 30 minutes to improve performance:

- Reduces network requests for repeated validations
- Automatic expiration prevents stale keys
- Thread-safe with mutex protection

To bypass the cache:

```go
resolver.ClearCache()
```

## Testing

The package includes comprehensive tests:

```bash
# Run all crypto tests
go test ./pkg/crypto/... -v

# Run specific tests
go test ./pkg/crypto/... -v -run TestSignAndValidateVC
go test ./pkg/crypto/... -v -run TestDIDResolver
```

### Test Coverage

- JWT signing and validation
- Expired credential detection
- Invalid signature detection
- Nonce and audience mismatch detection
- DID resolution and caching
- Multiple key types and curves

## Integration with VP Service

The crypto package integrates with the VP validation service:

```go
// Create a service with default resolver
service := vp.NewService()

// Or with a custom resolver for testing
resolver := crypto.NewDIDResolver()
resolver.RegisterLocalKey("did:example:test", publicKey)
service := vp.NewServiceWithResolver(resolver)

// Validate presentations
result, status, err := service.Validate(ctx, presentations)
```

## Supported Algorithms

### ECDSA

- **ES256**: P-256 curve (recommended for most use cases)
- **ES384**: P-384 curve (higher security)
- **ES512**: P-521 curve (highest security)

### RSA

- **RS256**: RSA with SHA-256
- **RS384**: RSA with SHA-384
- **RS512**: RSA with SHA-512

### EdDSA

- **EdDSA**: Ed25519 curve (compact signatures)

## Future Enhancements

Potential improvements for future versions:

1. **did:key Support**: Full multicodec/multibase decoding
2. **did:ion Support**: Bitcoin-anchored DIDs
3. **Revocation Checking**: Validate credential status lists
4. **Selective Disclosure**: Support for SD-JWT credentials
5. **Zero-Knowledge Proofs**: ZKP-based credential presentations
6. **Hardware Security Modules**: HSM integration for key storage

## References

- [W3C Verifiable Credentials Data Model](https://www.w3.org/TR/vc-data-model/)
- [W3C Decentralized Identifiers (DIDs)](https://www.w3.org/TR/did-core/)
- [JSON Web Token (JWT) RFC 7519](https://tools.ietf.org/html/rfc7519)
- [JSON Web Signature (JWS) RFC 7515](https://tools.ietf.org/html/rfc7515)
- [DID Web Method Specification](https://w3c-ccg.github.io/did-method-web/)
