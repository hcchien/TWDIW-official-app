# Taiwan Digital Wallet API Server

HTTP REST API server for Taiwan Digital Wallet Verifier and Issuer services.

## ⚠️ SECURITY WARNING

**THIS IS A TEST/DEMO SERVER FOR THE FRAMEWORK IMPLEMENTATION**

- No cryptographic operations are implemented
- No authentication or authorization
- No database persistence
- NOT suitable for production use

See [../../SECURITY.md](../../SECURITY.md) for complete security information.

---

## Features

This API server provides HTTP REST endpoints for:

### Credential Issuance
- `POST /api/credential` - Generate a new credential
- `GET /api/credential/query?cid=...` - Query credential by ID
- `PUT /api/credential/revoke?cid=...` - Revoke a credential
- `PUT /api/credential/suspend?cid=...` - Suspend a credential
- `PUT /api/credential/recover?cid=...` - Recover a suspended credential

### VP Validation
- `POST /api/presentation/validation` - Validate verifiable presentations

### OID4VP Verification
- `POST /api/oidvp/verify` - Verify OID4VP authorization response
- `GET /api/oidvp/result?client_id=...&nonce=...` - Get verification result

### Health Check
- `GET /api/health` - Server health status

### Web Interface
- `GET /` - Web-based test interface

---

## Quick Start

### Build

```bash
cd /path/to/verifier-go
go build -o bin/api-server ./cmd/api-server
```

### Run

```bash
./bin/api-server
```

The server will start on port 8080 by default.

### Custom Port

```bash
PORT=3000 ./bin/api-server
```

### Access

- Web Interface: http://localhost:8080
- API Base URL: http://localhost:8080/api
- Health Check: http://localhost:8080/api/health

---

## API Documentation

### Generate Credential

**Endpoint:** `POST /api/credential`

**Request Body:**
```json
{
  "credential_type": "IdentityCredential",
  "credential_subject": {
    "name": "John Doe",
    "birthDate": "1990-01-01",
    "nationalID": "A123456789"
  }
}
```

**Response (201 Created):**
```json
{
  "cid": "generated-credential-id",
  "credential": "eyJhbGci... (placeholder JWT)",
  "nonce": "nonce-value"
}
```

**Response (400 Bad Request):**
```json
{
  "code": 61001,
  "message": "Invalid credential generation request"
}
```

---

### Query Credential

**Endpoint:** `GET /api/credential/query?cid=<credential-id>`

**Response (404 Not Found):**
```json
{
  "code": 61010,
  "message": "Credential not found"
}
```

Note: Always returns 404 in framework implementation (no database).

---

### Revoke Credential

**Endpoint:** `PUT /api/credential/revoke?cid=<credential-id>`

**Response (200 OK):**
```json
{
  "cid": "credential-id",
  "status": "REVOKED"
}
```

---

### Suspend Credential

**Endpoint:** `PUT /api/credential/suspend?cid=<credential-id>`

**Response (200 OK):**
```json
{
  "cid": "credential-id",
  "status": "SUSPENDED"
}
```

---

### Recover Credential

**Endpoint:** `PUT /api/credential/recover?cid=<credential-id>`

**Response (200 OK):**
```json
{
  "cid": "credential-id",
  "status": "ACTIVE"
}
```

---

### Validate Presentations

**Endpoint:** `POST /api/presentation/validation`

**Request Body:**
```json
[
  "eyJhbGciOiJFUzI1NiJ9.presentation1.signature",
  "eyJhbGciOiJFUzI1NiJ9.presentation2.signature"
]
```

**Response (200 OK):**
```json
[
  {
    "client_id": "test-client-id",
    "nonce": "test-nonce",
    "holder_did": "did:example:holder",
    "verifiable_credentials": []
  }
]
```

**Response (400 Bad Request):**
```json
{
  "code": 71001,
  "message": "too many presentations: maximum 100 allowed"
}
```

---

### Verify OID4VP

**Endpoint:** `POST /api/oidvp/verify`

**Request Body:**
```json
{
  "vp_token": "eyJhbGciOiJFUzI1NiJ9.vp.signature",
  "presentation_submission": "{\"id\":\"ps1\",\"definition_id\":\"pd1\"}",
  "nonce": "test-nonce",
  "client_id": "test-client-id",
  "presentation_definition": "{\"id\":\"pd1\",\"input_descriptors\":[]}"
}
```

**Response (200 OK):**
```json
{
  "verify_result": true,
  "holder_did": "did:example:holder",
  "error": null
}
```

---

### Health Check

**Endpoint:** `GET /api/health`

**Response (200 OK):**
```json
{
  "status": "healthy",
  "time": "2026-01-20T10:30:00Z",
  "services": {
    "vp": "ready",
    "oidvp": "ready",
    "credential": "ready"
  }
}
```

---

## Web Interface

The server includes a web-based test interface accessible at `http://localhost:8080`.

Features:
- **Credential Issuer** tab - Test credential generation and management
- **VP Verifier** tab - Test VP validation
- **OID4VP** tab - Test OID4VP verification

The interface provides:
- Pre-filled example data
- JSON syntax validation
- Pretty-printed responses
- Color-coded success/error states

---

## Architecture

```
cmd/api-server/
├── main.go              # HTTP server and routing
└── README.md           # This file

web/
└── index.html          # Web interface

Dependencies:
├── verifier-go/pkg/vp          # VP validation service
├── verifier-go/pkg/oidvp       # OID4VP verification service
└── issuer-go/pkg/credential    # Credential issuance service
```

### Server Structure

The server is built using Go's standard library `net/http` with no external dependencies:

- **HTTP Server**: `http.Server` with configurable timeouts
- **Routing**: `http.ServeMux` for URL routing
- **Middleware**: CORS and logging middleware
- **Graceful Shutdown**: Signal handling for clean shutdown

### Middleware

**CORS Middleware**:
- Allows all origins (*)
- Supports GET, POST, PUT, DELETE, OPTIONS
- Allows Content-Type and Authorization headers

**Logging Middleware**:
- Logs all HTTP requests
- Records method, URL, and duration

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |

### Defaults (in code)

```go
const (
    DefaultPort       = "8080"
    DefaultIssuerDID  = "did:example:issuer"
    DefaultIssuerKey  = "issuer-key-placeholder"
    DefaultVPVerifyURI = "http://localhost:8080/api/vp/validate"
)
```

---

## Error Codes

### Credential Service (61xxx, 62xxx, 63xxx, 68xxx, 69xxx)

| Code | HTTP Status | Description |
|------|-------------|-------------|
| 61001 | 400 | Invalid credential generation request |
| 61006 | 400 | Invalid credential ID |
| 61010 | 404 | Credential not found |
| 69004 | 500 | Issuer has not registered a DID |

### VP Service (71xxx, 72xxx, 78xxx)

| Code | HTTP Status | Description |
|------|-------------|-------------|
| 71001 | 400 | Invalid presentation validation request |
| 71002 | 500 | VP validation error |
| 72001 | 500 | VC content validation error |

See [../../SECURITY.md](../../SECURITY.md) for complete error code reference.

---

## Testing with cURL

### Generate Credential

```bash
curl -X POST http://localhost:8080/api/credential \
  -H "Content-Type: application/json" \
  -d '{
    "credential_type": "IdentityCredential",
    "credential_subject": {
      "name": "John Doe",
      "age": 30
    }
  }'
```

### Validate Presentations

```bash
curl -X POST http://localhost:8080/api/presentation/validation \
  -H "Content-Type: application/json" \
  -d '["eyJhbGciOiJFUzI1NiJ9.test1.sig", "eyJhbGciOiJFUzI1NiJ9.test2.sig"]'
```

### Query Credential

```bash
curl http://localhost:8080/api/credential/query?cid=test-id
```

### Revoke Credential

```bash
curl -X PUT http://localhost:8080/api/credential/revoke?cid=test-id
```

### Health Check

```bash
curl http://localhost:8080/api/health
```

---

## Input Validation

The server implements DoS protection through input validation:

**VP Validation Limits**:
- Maximum 100 presentations per request
- Maximum 1MB per presentation
- Maximum 10MB total payload

**Credential Limits**:
- Maximum 1000 fields in credential subject
- Maximum 1MB per string field
- Maximum 10 levels of map nesting

Requests exceeding these limits return HTTP 400 Bad Request.

---

## Graceful Shutdown

The server handles SIGINT and SIGTERM signals for graceful shutdown:

1. Signal received (Ctrl+C or kill)
2. Stop accepting new connections
3. Wait up to 10 seconds for active requests to complete
4. Shutdown complete

---

## Development

### Run in Development Mode

```bash
# From verifier-go directory
go run ./cmd/api-server
```

### Build for Production

```bash
# Build with optimizations
go build -ldflags="-s -w" -o bin/api-server ./cmd/api-server

# Run
./bin/api-server
```

### Run Tests

The API server uses the existing service layer tests:

```bash
# Test all services
go test ./...

# Test with coverage
go test ./... -cover

# Test specific package
go test ./pkg/vp -v
go test ./pkg/credential -v
```

---

## Limitations

This is a **framework/skeleton implementation** with the following limitations:

### NOT Implemented (CRITICAL)
- ❌ JWT parsing and signature verification
- ❌ Cryptographic operations (all stubbed)
- ❌ Database persistence
- ❌ Authentication
- ❌ Authorization
- ❌ Rate limiting
- ❌ Audit logging
- ❌ DID resolution
- ❌ Credential schema validation
- ❌ Status list generation

### Implemented (Framework Features)
- ✅ HTTP REST API endpoints
- ✅ Input validation and DoS protection
- ✅ Error handling with proper status codes
- ✅ CORS support
- ✅ Request logging
- ✅ Graceful shutdown
- ✅ Web test interface
- ✅ Comprehensive tests (51 tests, 100% passing)

---

## Production Deployment (NOT RECOMMENDED)

**DO NOT deploy this server to production** until ALL critical issues in [../../SECURITY.md](../../SECURITY.md) are resolved.

Required before production:
1. Implement JWT signing and verification
2. Add database integration
3. Implement authentication/authorization
4. Add rate limiting
5. Implement audit logging
6. Add TLS/HTTPS support
7. Implement proper DID resolution
8. Add credential schema validation

---

## Support

For issues or questions:
- See [../../SECURITY.md](../../SECURITY.md) for security concerns
- See [../../README.md](../../README.md) for general information
- Refer to main TWDIW project documentation

---

## License

See LICENSE.txt in the project root.
