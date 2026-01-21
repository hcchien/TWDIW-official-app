# Deployment Guide

**Taiwan Digital Wallet - Go Implementation**

---

## ⚠️ CRITICAL WARNING

**THIS IS A FRAMEWORK/SKELETON IMPLEMENTATION**

🚨 **DO NOT DEPLOY TO PRODUCTION** 🚨

This code has **CRITICAL security vulnerabilities**. See [SECURITY.md](SECURITY.md) for details.

**Safe for:**
- ✅ Local development and testing
- ✅ GitHub repository (with warnings)
- ✅ Internal demos and proof-of-concept

**NOT safe for:**
- ❌ Production deployment
- ❌ Public internet exposure
- ❌ Processing real credentials or sensitive data

---

## Quick Start (Local Development)

### Prerequisites

- **Go 1.22+** - [Download](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Make** (optional) - For using Makefile commands

### 1. Clone Repository

```bash
git clone <repository-url>
cd verifier-go
```

### 2. Install Dependencies

```bash
make install
# Or manually:
go mod download
go mod tidy
```

### 3. Build

```bash
make build
# Or manually:
go build -o bin/api-server ./cmd/api-server
```

### 4. Run

```bash
make run
# Or manually:
./bin/api-server
```

### 5. Access

- **Web Interface**: http://localhost:8080
- **API**: http://localhost:8080/api
- **Health Check**: http://localhost:8080/api/health

---

## Development Workflow

### Quick Development Cycle

```bash
# Format, build, and run in one command
make quick
```

### Development Mode (No Build)

```bash
# Run directly from source
make dev
# Or:
go run ./cmd/api-server
```

### Run Tests

```bash
# All tests
make test

# Specific services
make test-vp
make test-oidvp
make test-issuer

# With coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Vet code
make vet

# Lint (requires golangci-lint)
make lint

# All checks
make check
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |

### Custom Port

```bash
# Using environment variable
PORT=3000 ./bin/api-server

# Using Makefile
make run-port PORT=3000
```

### Configuration in Code

Edit `cmd/api-server/main.go` to change:

```go
const (
    DefaultPort       = "8080"
    DefaultIssuerDID  = "did:example:issuer"
    DefaultIssuerKey  = "issuer-key-placeholder"
    DefaultVPVerifyURI = "http://localhost:8080/api/vp/validate"
)
```

---

## Project Structure

```
verifier-go/
├── cmd/
│   └── api-server/          # HTTP API server
│       ├── main.go          # Server implementation
│       └── README.md        # API documentation
├── pkg/
│   ├── errors/              # Error handling
│   ├── models/              # Data models
│   ├── vp/                  # VP validation service
│   └── oidvp/               # OID4VP verification service
├── issuer-go/
│   └── pkg/
│       ├── credential/      # Credential issuance service
│       ├── errors/          # Credential error codes
│       └── models/          # Credential data models
├── web/
│   └── index.html          # Web test interface
├── bin/                    # Build artifacts (gitignored)
├── go.mod                  # Go module definition
├── Makefile               # Build automation
├── README.md              # Main documentation
├── SECURITY.md            # Security audit
├── SECURITY-FIXES-APPLIED.md  # Applied security fixes
└── DEPLOYMENT.md          # This file
```

---

## Building

### Development Build

```bash
make build
# Output: bin/api-server (~10MB)
```

### Production Build (Optimized)

```bash
make build-prod
# Output: bin/api-server (~7MB, stripped)
```

### Manual Build

```bash
# Basic build
go build -o bin/api-server ./cmd/api-server

# With optimizations
go build -ldflags="-s -w" -o bin/api-server ./cmd/api-server

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o bin/api-server-linux ./cmd/api-server
```

---

## Running

### Foreground

```bash
./bin/api-server
```

### Background (with nohup)

```bash
nohup ./bin/api-server > server.log 2>&1 &
echo $! > server.pid
```

### Stop Background Server

```bash
kill $(cat server.pid)
rm server.pid
```

### With systemd (Linux)

Create `/etc/systemd/system/twdiw-api.service`:

```ini
[Unit]
Description=Taiwan Digital Wallet API Server
After=network.target

[Service]
Type=simple
User=twdiw
WorkingDirectory=/opt/twdiw
ExecStart=/opt/twdiw/bin/api-server
Restart=on-failure
RestartSec=5s
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable twdiw-api
sudo systemctl start twdiw-api
sudo systemctl status twdiw-api
```

View logs:

```bash
sudo journalctl -u twdiw-api -f
```

---

## Docker Deployment (NOT RECOMMENDED FOR PRODUCTION)

### Create Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy module files
COPY go.mod go.sum ./
COPY issuer-go/go.mod issuer-go/
RUN go mod download

# Copy source
COPY . .

# Build
RUN go build -ldflags="-s -w" -o /app/bin/api-server ./cmd/api-server

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and web files
COPY --from=builder /app/bin/api-server .
COPY --from=builder /app/web ./web

# Expose port
EXPOSE 8080

# Run
CMD ["./api-server"]
```

### Build Docker Image

```bash
docker build -t twdiw-api:latest .
```

### Run Docker Container

```bash
docker run -d \
  --name twdiw-api \
  -p 8080:8080 \
  -e PORT=8080 \
  twdiw-api:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

Run:

```bash
docker-compose up -d
docker-compose logs -f
```

---

## Testing

### Run All Tests

```bash
make test
```

### Test Individual Services

```bash
# VP validation
go test ./pkg/vp -v

# OID4VP verification
go test ./pkg/oidvp -v

# Credential issuance
go test ./issuer-go/pkg/credential -v

# All error handling
go test ./pkg/errors ./issuer-go/pkg/errors -v
```

### Coverage Report

```bash
make test-coverage
# Opens coverage.html in browser
```

### Benchmark Tests

```bash
go test ./pkg/vp -bench=. -benchmem
go test ./issuer-go/pkg/credential -bench=. -benchmem
```

---

## API Testing

### Using cURL

#### Health Check

```bash
curl http://localhost:8080/api/health
```

#### Generate Credential

```bash
curl -X POST http://localhost:8080/api/credential \
  -H "Content-Type: application/json" \
  -d '{
    "credential_type": "IdentityCredential",
    "credential_subject": {
      "name": "John Doe",
      "birthDate": "1990-01-01"
    }
  }'
```

#### Validate VP

```bash
curl -X POST http://localhost:8080/api/presentation/validation \
  -H "Content-Type: application/json" \
  -d '["eyJhbGciOiJFUzI1NiJ9.vp1.sig", "eyJhbGciOiJFUzI1NiJ9.vp2.sig"]'
```

### Using HTTPie

```bash
# Install: brew install httpie

# Health check
http GET :8080/api/health

# Generate credential
http POST :8080/api/credential \
  credential_type=IdentityCredential \
  credential_subject:='{"name":"John Doe"}'

# Validate VP
http POST :8080/api/presentation/validation \
  --json < presentations.json
```

### Using Postman

1. Import collection from `docs/postman_collection.json` (if exists)
2. Or create requests manually:
   - Base URL: `http://localhost:8080/api`
   - Add endpoints as documented in `cmd/api-server/README.md`

---

## Monitoring

### Health Checks

```bash
# Simple health check
curl http://localhost:8080/api/health

# With response details
curl -s http://localhost:8080/api/health | jq .
```

### Logs

The server logs to stdout:

```
2026/01/20 10:30:15 Starting API server on port 8080
2026/01/20 10:30:15 API endpoints:
...
2026/01/20 10:30:20 POST /api/credential 15ms
2026/01/20 10:30:21 GET /api/health 1ms
```

### Metrics (Future)

Currently no metrics endpoint. To add Prometheus metrics:

1. Add `github.com/prometheus/client_golang/prometheus` dependency
2. Create `/metrics` endpoint
3. Export custom metrics (request count, duration, errors)

---

## Performance

### Current Performance

- **Build time**: <1 second
- **Test execution**: <1 second (51 tests)
- **Memory footprint**: ~10MB
- **Binary size**: ~7MB (optimized)
- **Request latency**: <1ms (framework operations)

### Load Testing (with hey)

```bash
# Install: brew install hey

# Test health endpoint
hey -n 10000 -c 100 http://localhost:8080/api/health

# Test credential generation
hey -n 1000 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"credential_type":"Test","credential_subject":{"test":"data"}}' \
  http://localhost:8080/api/credential
```

---

## Troubleshooting

### Server Won't Start

**Error:** `bind: address already in use`

**Solution:**

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use different port
PORT=3000 ./bin/api-server
```

### Build Failures

**Error:** `package not found`

**Solution:**

```bash
go mod download
go mod tidy
go build ./cmd/api-server
```

### Test Failures

**Error:** Tests fail after code changes

**Solution:**

```bash
# Clean and rebuild
make clean
make install
make test
```

### CORS Issues

**Error:** Frontend can't access API

**Solution:** The server has CORS enabled by default (`Access-Control-Allow-Origin: *`)

If still having issues, check browser console and verify request headers.

---

## Security Checklist (Before ANY Deployment)

Before deploying to ANY environment:

### Critical (MUST FIX)
- [ ] Implement JWT signing and verification
- [ ] Add input validation limits (✅ DONE)
- [ ] Sanitize error messages (✅ DONE)
- [ ] Add authentication
- [ ] Add authorization

### High Priority
- [ ] Add rate limiting
- [ ] Implement audit logging
- [ ] Add database integration
- [ ] Configure TLS/HTTPS
- [ ] Add context timeout handling (✅ DONE)

### Medium Priority
- [ ] Add request ID tracking
- [ ] Implement proper logging levels
- [ ] Add metrics/monitoring
- [ ] Configure CORS properly
- [ ] Add health check dependencies

See [SECURITY.md](SECURITY.md) for complete checklist.

---

## Maintenance

### Update Dependencies

```bash
go get -u ./...
go mod tidy
go test ./...
```

### Version Management

Update version in code:

```go
const Version = "0.1.0"
```

Add to health endpoint response.

### Logs Rotation

Using logrotate (`/etc/logrotate.d/twdiw-api`):

```
/var/log/twdiw-api/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    missingok
    sharedscripts
    postrotate
        systemctl reload twdiw-api
    endscript
}
```

---

## Backup and Recovery

### What to Backup (if using database in future)

- Database files/dumps
- Configuration files
- DID key material
- Audit logs

### Current State (Framework)

No backup needed - stateless application with no persistence.

---

## Shutdown

### Graceful Shutdown

The server handles SIGINT/SIGTERM:

```bash
# Send interrupt signal
kill -INT $(cat server.pid)

# Server will:
# 1. Stop accepting new connections
# 2. Wait up to 10s for active requests
# 3. Shut down cleanly
```

### Force Shutdown

```bash
kill -9 $(cat server.pid)
```

---

## Next Steps (Post-Framework)

After implementing security features:

1. **Add Database**
   - PostgreSQL for credentials
   - Redis for caching
   - Database migrations

2. **Add Authentication**
   - OAuth 2.0 / OIDC
   - API keys
   - JWT tokens

3. **Add Monitoring**
   - Prometheus metrics
   - Grafana dashboards
   - Alert manager

4. **Add Production Features**
   - Load balancing
   - Auto-scaling
   - Blue-green deployment
   - CI/CD pipeline

---

## Support

For help:
- **Security Issues**: See [SECURITY.md](SECURITY.md)
- **API Documentation**: See [cmd/api-server/README.md](cmd/api-server/README.md)
- **General Info**: See [README.md](README.md)

---

## License

See LICENSE.txt in project root.
