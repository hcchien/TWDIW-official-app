# TWDIW Verifier - Go Implementation

Go implementation of the Taiwan Digital Wallet Verifier services, rewritten from Java.

---

## ⚠️ SECURITY WARNING - NOT PRODUCTION READY ⚠️

**THIS CODE REQUIRES ADDITIONAL SECURITY HARDENING**

🚨 **DO NOT USE IN PRODUCTION WITHOUT COMPLETING THE CHECKLIST** 🚨

### Security Status

This implementation includes:

✅ **Implemented:**
- Full JWT signature verification for VCs and VPs
- DID resolution and public key validation
- Input size limits and DoS protection
- Nonce and audience verification
- Expiration and timestamp validation

❌ **Still Required for Production:**
- Authentication and authorization for API endpoints
- Rate limiting and throttling
- Production-grade error handling (no information leakage)
- Audit logging and monitoring
- Credential revocation checking
- Database security and access controls

**See [SECURITY.md](SECURITY.md) for complete security audit and required fixes.**

### What This Code Does

This implementation provides:
- ✅ Full cryptographic validation of VCs and VPs
- ✅ DID resolution with caching
- ✅ Excellent test coverage (40+ tests, 100% passing)
- ✅ Clean architecture and code structure
- ✅ Compatible error codes with Java implementation
- ⚠️ Partial security controls (cryptographic layer complete)
- ❌ **NO authentication/authorization layer**

### Before Production Use

You MUST implement:
- Authentication and authorization middleware
- Rate limiting
- Proper error handling that doesn't leak sensitive information
- Audit logging
- Credential revocation checking
- Security monitoring and alerts

**Read the complete [Pre-Production Checklist](SECURITY.md#pre-production-checklist) before deployment.**

---

## Overview

This project provides verification services for:
- **VP (Verifiable Presentation)** validation
- **OID4VP (OpenID for Verifiable Presentations)** verification
- Error handling matching Java implementation
- Comprehensive test coverage

## Project Structure

```
verifier-go/
├── pkg/
│   ├── crypto/           # Cryptographic validation
│   │   ├── jwt.go              # JWT signing and validation
│   │   ├── jwt_test.go         # JWT tests
│   │   ├── did_resolver.go     # DID resolution
│   │   ├── did_resolver_test.go # DID tests
│   │   └── README.md           # Crypto package docs
│   ├── errors/           # Error codes and error handling
│   │   ├── errors.go
│   │   └── errors_test.go
│   ├── models/           # Data models and DTOs
│   │   └── models.go
│   ├── vp/               # VP validation service
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── service_integration_test.go
│   └── oidvp/            # OID4VP verification service
│       ├── service.go
│       └── service_test.go
├── cmd/
│   └── api-server/       # HTTP server
│       └── main.go
├── internal/
│   ├── config/           # Configuration (future)
│   └── middleware/       # HTTP middleware (future)
├── go.mod
├── go.sum
└── README.md
```

## Features

### Cryptographic Validation (`pkg/crypto`)

Full JWT-based cryptographic validation:

- **JWT Validation** - Complete signature verification for VCs and VPs
- **DID Resolution** - Resolve DIDs to public keys with caching
- **Multiple Algorithms** - ECDSA (P-256/384/521), RSA, EdDSA support
- **Security Checks** - Expiration, nonce, audience, key consistency validation
- **did:web Support** - Fetch DID documents from web endpoints
- **Testing Support** - Local key registration for testing

See [pkg/crypto/README.md](pkg/crypto/README.md) for detailed documentation.

### VP Validation Service (`pkg/vp`)

Equivalent to Java's `PresentationServiceAsync`:

- **Validate()** - Validates a list of verifiable presentations with full cryptographic verification
- **Cryptographic Validation** - JWT signature verification for all VPs and embedded VCs
- **DID Resolution** - Automatic public key resolution from issuer and holder DIDs
- **Security Checks** - Expiration, signature, nonce, and audience validation
- **DoS Protection** - Input size limits (1MB per presentation, 10MB total, max 100 presentations)
- **Error Handling** - Detailed error responses with proper HTTP status codes
- Handles null/empty/blank presentation lists
- Returns JSON responses with proper HTTP status codes
- Supports multiple presentation validation

### OID4VP Verification Service (`pkg/oidvp`)

Equivalent to Java's `VerifierService`:

- **Verify()** - Verifies OID4VP authorization responses
- **VerifyPresentation()** - Validates VP tokens and presentation submissions
- **GetVerifyResult()** - Retrieves stored verification results
- **ModifyPresentationDefinitionData()** - Manages presentation definitions

### Error Handling (`pkg/errors`)

Error codes matching Java's `VpException`:

```go
const (
    // Presentation errors (71xxx)
    ErrPresInvalidPresentationValidationRequest = 71001
    ErrPresValidateVPError                      = 71002
    ErrPresValidateVPContentError               = 71003

    // Credential errors (72xxx)
    ErrCredValidateVCContentError = 72001
    ErrCredValidateVCSchemaError  = 72002

    // Database errors (78xxx)
    ErrDBQueryError  = 78001
    ErrDBInsertError = 78002
)
```

## Installation

```bash
# Clone the repository
cd verifier-go

# Download dependencies
go mod tidy

# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -cover
```

## Usage

### Cryptographic Validation

```go
package main

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/moda-gov-tw/twdiw-verifier-go/pkg/crypto"
)

func main() {
    // Generate keys
    privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

    // Create and sign a VC
    vcClaims := &crypto.VCClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "did:web:issuer.example.com",
            Subject:   "did:web:holder.example.com",
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ID:        "vc-12345",
        },
        VC: crypto.CredentialSubject{
            Context: []string{"https://www.w3.org/2018/credentials/v1"},
            Type:    []string{"VerifiableCredential", "NationalIDCredential"},
            CredentialSubject: map[string]interface{}{
                "id": "did:web:holder.example.com",
                "nationalID": "A123456789",
            },
        },
    }

    vcJWT, _ := crypto.SignVC(vcClaims, privateKey, "did:web:issuer.example.com#key-1")

    // Validate the VC
    resolver := crypto.NewDIDResolver()
    validator := crypto.NewJWTValidator(resolver)

    validatedClaims, err := validator.ValidateVC(vcJWT)
    if err != nil {
        // Handle validation error
        panic(err)
    }

    fmt.Printf("Valid VC from: %s\n", validatedClaims.Issuer)
}
```

### VP Validation Service

```go
package main

import (
    "context"
    "fmt"
    "github.com/moda-gov-tw/twdiw-verifier-go/pkg/vp"
)

func main() {
    // Create service (includes cryptographic validation)
    service := vp.NewService()

    // Validate presentations with real JWT signatures
    presentations := []string{
        "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkaWQ6d2ViOmlzc3Vlci5jb20iLCAic3ViIjoiZGlkOndlYjpob2xkZXIuY29tIiwgInZwIjp7fX0.signature",
    }

    result, status, err := service.Validate(context.Background(), presentations)
    if err != nil {
        fmt.Printf("Error: %v (HTTP %d)\n", err, status)
        return
    }

    fmt.Printf("Result: %s\n", result)
}
```

### OID4VP Verification Service

```go
package main

import (
    "context"
    "fmt"
    "github.com/moda-gov-tw/twdiw-verifier-go/pkg/oidvp"
    "github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
)

func main() {
    // Create service
    service := oidvp.NewVerifierService("http://vp-validator:8080/verify")

    // Prepare authorization response
    authzResponse := &models.OIDVPAuthorizationResponse{
        VPToken:               "eyJhbGciOiJFUzI1NiJ9.vp.signature",
        PresentationSubmission: `{"id":"ps1","definition_id":"pd1"}`,
    }

    // Verify
    result, err := service.Verify(
        context.Background(),
        authzResponse,
        "test-nonce",
        "test-client-id",
        `{"id":"pd1","input_descriptors":[]}`,
    )

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Verification result: %v\n", result.VerifyResult)
    fmt.Printf("Holder DID: %s\n", result.HolderDID)
}
```

## Testing

### Run All Tests

```bash
go test ./... -v
```

### Run Specific Package Tests

```bash
# Cryptographic validation tests
go test ./pkg/crypto -v

# VP service tests
go test ./pkg/vp -v

# OID4VP service tests
go test ./pkg/oidvp -v

# Error handling tests
go test ./pkg/errors -v

# Integration tests with real cryptographic validation
go test ./pkg/vp -v -run "TestValidate_WithRealJWT"
```

### Test Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

## Test Results

```
✅ pkg/crypto:  PASS (16 tests) - JWT signing, validation, DID resolution
✅ pkg/errors:  PASS (5 tests)  - Error handling and codes
✅ pkg/oidvp:   PASS (11 tests) - OID4VP verification
✅ pkg/vp:      PASS (14 tests) - VP validation (9 unit + 3 integration + 2 helpers)

Total: 46 tests passing
Coverage: High (all critical paths tested including cryptographic validation)
```

## Comparison with Java Implementation

| Feature | Java | Go | Status |
|---------|------|-----|--------|
| VP Validation | PresentationServiceAsync | vp.Service | ✅ Implemented |
| OID4VP Verification | VerifierService | oidvp.VerifierService | ✅ Implemented |
| JWT Validation | JWT libraries | crypto.JWTValidator | ✅ Implemented |
| DID Resolution | DID resolver | crypto.DIDResolver | ✅ Implemented |
| Signature Verification | ECDSA/RSA | ECDSA/RSA/EdDSA | ✅ Implemented |
| Error Codes | VpException | errors.VPError | ✅ Matching |
| Data Models | DTOs | models package | ✅ Implemented |
| Test Coverage | JUnit 5 + Mockito | Go testing | ✅ Comprehensive (46 tests) |
| HTTP Status Mapping | toHttpStatus() | HTTPStatus() | ✅ Matching |

## Migration Notes

### Key Differences from Java

1. **Error Handling**
   - Java: Exceptions with try-catch
   - Go: Error returns with explicit handling

2. **Dependency Injection**
   - Java: Spring @Autowired
   - Go: Constructor injection

3. **Async/Concurrency**
   - Java: CompletableFuture
   - Go: Goroutines and channels (to be implemented)

4. **JSON Handling**
   - Java: Jackson annotations
   - Go: struct tags

### Maintained Compatibility

- ✅ Error code numbers identical
- ✅ HTTP status code mapping identical
- ✅ Response JSON structure compatible
- ✅ Method signatures equivalent

## Future Enhancements

- [x] Implement JWT VP/VC parsing and validation
- [x] Implement DID resolution
- [x] Add cryptographic signature verification
- [ ] Add credential revocation checking (status list)
- [ ] Add HTTP REST API server with authentication
- [ ] Add database integration
- [ ] Implement presentation definition evaluation
- [ ] Add asynchronous VC validation with goroutines
- [ ] Add logging and tracing
- [ ] Add metrics and monitoring
- [ ] Add Docker containerization
- [ ] Add API documentation (Swagger/OpenAPI)
- [ ] Support did:key and did:ion methods
- [ ] Add selective disclosure (SD-JWT) support

## Development

### Prerequisites

- Go 1.21 or higher
- Git

### Building

```bash
# Build the project
go build ./...

# Build with optimizations
go build -ldflags="-s -w" ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...
```

## Contributing

1. Follow Go best practices
2. Maintain error code compatibility with Java implementation
3. Write tests for all new features
4. Keep test coverage above 80%
5. Document public APIs

## License

See LICENSE.txt in the project root.

## Support

For questions or issues, please refer to the main TWDIW project documentation.
