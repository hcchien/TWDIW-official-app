package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moda-gov-tw/twdiw-issuer-go/pkg/credential"
	"github.com/moda-gov-tw/twdiw-issuer-go/pkg/models"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/oidvp"
	verifierModels "github.com/moda-gov-tw/twdiw-verifier-go/pkg/models"
	"github.com/moda-gov-tw/twdiw-verifier-go/pkg/vp"
)

const (
	DefaultPort       = "8080"
	DefaultIssuerDID  = "did:example:issuer"
	DefaultIssuerKey  = "issuer-key-placeholder"
	DefaultVPVerifyURI = "http://localhost:8080/api/vp/validate"
)

type Server struct {
	// Services
	vpService         *vp.Service
	oidvpService      *oidvp.VerifierService
	credentialService *credential.Service

	// HTTP server
	httpServer *http.Server
}

func NewServer() *Server {
	return &Server{
		vpService:         vp.NewService(),
		oidvpService:      oidvp.NewVerifierService(DefaultVPVerifyURI),
		credentialService: credential.NewService(DefaultIssuerDID, DefaultIssuerKey),
	}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", s.handleHealth)

	// Credential issuance endpoints
	mux.HandleFunc("/api/credential", s.handleCredentialGenerate)           // POST
	mux.HandleFunc("/api/credential/query", s.handleCredentialQuery)        // GET
	mux.HandleFunc("/api/credential/revoke", s.handleCredentialRevoke)      // PUT
	mux.HandleFunc("/api/credential/suspend", s.handleCredentialSuspend)    // PUT
	mux.HandleFunc("/api/credential/recover", s.handleCredentialRecover)    // PUT

	// VP validation endpoints
	mux.HandleFunc("/api/presentation/validation", s.handleVPValidation)    // POST

	// OID4VP verification endpoints
	mux.HandleFunc("/api/oidvp/verify", s.handleOIDVPVerify)               // POST
	mux.HandleFunc("/api/oidvp/result", s.handleOIDVPGetResult)            // GET

	// Static files for frontend
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	// CORS middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting API server on port %s", port)
	log.Printf("API endpoints:")
	log.Printf("  POST   /api/credential               - Generate credential")
	log.Printf("  GET    /api/credential/query?cid=... - Query credential")
	log.Printf("  PUT    /api/credential/revoke?cid=.. - Revoke credential")
	log.Printf("  POST   /api/presentation/validation  - Validate VP")
	log.Printf("  POST   /api/oidvp/verify             - Verify OID4VP")
	log.Printf("  GET    /api/health                   - Health check")
	log.Printf("Web interface: http://localhost:%s", port)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// Health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
		"services": map[string]string{
			"vp":         "ready",
			"oidvp":      "ready",
			"credential": "ready",
		},
	})
}

// Credential generation endpoint
func (s *Server) handleCredentialGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to CredentialRequestDTO
	// For now, pass through as-is since we're using map[string]interface{}
	ctx := r.Context()

	// Extract fields (simplified for demonstration)
	issuerDID, _ := request["issuer_did"].(string)
	credType, _ := request["credential_type"].(string)
	credSubject, _ := request["credential_subject"].(map[string]interface{})

	if issuerDID == "" {
		issuerDID = DefaultIssuerDID
	}

	req := &models.CredentialRequestDTO{
		IssuerDID:         issuerDID,
		CredentialType:    credType,
		CredentialSubject: credSubject,
	}

	result, status, err := s.credentialService.Generate(ctx, req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err != nil {
		w.Write([]byte(result))
	} else {
		w.Write([]byte(result))
	}
}

// Credential query endpoint
func (s *Server) handleCredentialQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cid := r.URL.Query().Get("cid")
	ctx := r.Context()

	result, status, _ := s.credentialService.Query(ctx, cid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(result))
}

// Credential revoke endpoint
func (s *Server) handleCredentialRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cid := r.URL.Query().Get("cid")
	ctx := r.Context()

	result, status, _ := s.credentialService.Revoke(ctx, cid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(result))
}

// Credential suspend endpoint
func (s *Server) handleCredentialSuspend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cid := r.URL.Query().Get("cid")
	ctx := r.Context()

	result, status, _ := s.credentialService.Suspend(ctx, cid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(result))
}

// Credential recover endpoint
func (s *Server) handleCredentialRecover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cid := r.URL.Query().Get("cid")
	ctx := r.Context()

	result, status, _ := s.credentialService.Recover(ctx, cid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(result))
}

// VP validation endpoint
func (s *Server) handleVPValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var presentations []string
	if err := json.NewDecoder(r.Body).Decode(&presentations); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, status, _ := s.vpService.Validate(ctx, presentations)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(result))
}

// OID4VP verify endpoint
func (s *Server) handleOIDVPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		VPToken                string `json:"vp_token"`
		PresentationSubmission string `json:"presentation_submission"`
		Nonce                  string `json:"nonce"`
		ClientID               string `json:"client_id"`
		PresentationDefinition string `json:"presentation_definition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authzResponse := &verifierModels.OIDVPAuthorizationResponse{
		VPToken:                request.VPToken,
		PresentationSubmission: request.PresentationSubmission,
	}

	ctx := r.Context()
	result, err := s.oidvpService.Verify(
		ctx,
		authzResponse,
		request.Nonce,
		request.ClientID,
		request.PresentationDefinition,
	)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// OID4VP get result endpoint
func (s *Server) handleOIDVPGetResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	nonce := r.URL.Query().Get("nonce")

	ctx := r.Context()
	result, err := s.oidvpService.GetVerifyResult(ctx, clientID, nonce)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call next handler
		next.ServeHTTP(w, r)

		// Log request
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	server := NewServer()

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	if err := server.Start(port); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Sprintf("Server failed to start: %v", err))
	}

	log.Println("Server stopped")
}
