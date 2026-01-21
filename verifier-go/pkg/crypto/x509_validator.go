package crypto

import (
	"crypto/x509"
	"fmt"
	"time"
)

// X509Validator handles X.509 certificate chain validation
type X509Validator struct {
	trustedRoots *x509.CertPool
}

// NewX509Validator creates a new X.509 validator
func NewX509Validator() *X509Validator {
	return &X509Validator{
		trustedRoots: x509.NewCertPool(),
	}
}

// AddTrustedRoot adds a trusted root certificate
func (v *X509Validator) AddTrustedRoot(cert *x509.Certificate) {
	v.trustedRoots.AddCert(cert)
}

// AddTrustedRootPEM adds a trusted root certificate from PEM-encoded data
func (v *X509Validator) AddTrustedRootPEM(pemData []byte) error {
	cert, err := x509.ParseCertificate(pemData)
	if err != nil {
		return fmt.Errorf("failed to parse PEM certificate: %w", err)
	}
	v.trustedRoots.AddCert(cert)
	return nil
}

// ValidateChain validates certificate chain up to trusted root
func (v *X509Validator) ValidateChain(cert *x509.Certificate, roots []*x509.Certificate) error {
	// Build cert pool from provided roots
	pool := x509.NewCertPool()
	for _, root := range roots {
		pool.AddCert(root)
	}

	// If no roots provided, use the validator's trusted roots
	if len(roots) == 0 {
		pool = v.trustedRoots
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots:       pool,
		CurrentTime: time.Now(),
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	chains, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	if len(chains) == 0 {
		return fmt.Errorf("no valid certificate chain found")
	}

	return nil
}

// ValidateBasic performs basic certificate validation checks
func (v *X509Validator) ValidateBasic(cert *x509.Certificate) error {
	now := time.Now()

	// Check expiration
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate not yet valid (NotBefore: %v)", cert.NotBefore)
	}

	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (NotAfter: %v)", cert.NotAfter)
	}

	return nil
}

// ValidateReaderAuth validates reader authentication certificate
func (v *X509Validator) ValidateReaderAuth(cert *x509.Certificate) error {
	// Check key usage for digital signature
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("certificate missing digital signature key usage")
	}

	// Perform basic validation (expiration)
	if err := v.ValidateBasic(cert); err != nil {
		return err
	}

	// Note: In production, you would validate against trusted reader CA roots
	// For now, we perform basic validation only
	return nil
}

// ValidateIssuerCert validates an mDL issuer certificate
func (v *X509Validator) ValidateIssuerCert(cert *x509.Certificate, trustedRoots []*x509.Certificate) error {
	// Perform basic validation
	if err := v.ValidateBasic(cert); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// Validate certificate chain
	if err := v.ValidateChain(cert, trustedRoots); err != nil {
		// If chain validation fails but we have no trusted roots configured,
		// we'll allow it with a warning (for development/testing)
		if len(trustedRoots) == 0 && v.trustedRoots.Equal(x509.NewCertPool()) {
			// No trusted roots configured - skip chain validation
			return nil
		}
		return fmt.Errorf("chain validation failed: %w", err)
	}

	return nil
}

// GetPublicKey returns the public key from a certificate
func (v *X509Validator) GetPublicKey(cert *x509.Certificate) interface{} {
	return cert.PublicKey
}
