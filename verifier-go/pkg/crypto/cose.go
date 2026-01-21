package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/go-cose"
)

// COSEValidator handles COSE signature validation
type COSEValidator struct {
}

// NewCOSEValidator creates a new COSE validator
func NewCOSEValidator() *COSEValidator {
	return &COSEValidator{}
}

// ParseCOSESign1 parses a COSE_Sign1 structure
func (v *COSEValidator) ParseCOSESign1(data []byte) (payload []byte, signature []byte, protected cose.ProtectedHeader, unprotected cose.UnprotectedHeader, err error) {
	var msg cose.Sign1Message

	if err := cbor.Unmarshal(data, &msg); err != nil {
		return nil, nil, cose.ProtectedHeader{}, cose.UnprotectedHeader{}, fmt.Errorf("CBOR unmarshal failed: %w", err)
	}

	return msg.Payload, msg.Signature, msg.Headers.Protected, msg.Headers.Unprotected, nil
}

// VerifySignature verifies COSE signature using public key
func (v *COSEValidator) VerifySignature(coseSign1Data []byte, publicKey interface{}) error {
	var msg cose.Sign1Message

	if err := cbor.Unmarshal(coseSign1Data, &msg); err != nil {
		return fmt.Errorf("CBOR unmarshal failed: %w", err)
	}

	// Determine algorithm from key type
	var alg cose.Algorithm
	switch key := publicKey.(type) {
	case *ecdsa.PublicKey:
		switch key.Curve.Params().BitSize {
		case 256:
			alg = cose.AlgorithmES256
		case 384:
			alg = cose.AlgorithmES384
		case 521:
			alg = cose.AlgorithmES512
		default:
			return fmt.Errorf("unsupported ECDSA curve size: %d", key.Curve.Params().BitSize)
		}
	default:
		return fmt.Errorf("unsupported key type for COSE: %T", publicKey)
	}

	// Create verifier
	verifier, err := cose.NewVerifier(alg, publicKey)
	if err != nil {
		return fmt.Errorf("failed to create verifier: %w", err)
	}

	// Verify the COSE_Sign1 message
	if err := msg.Verify(nil, verifier); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// ExtractCertificateFromCOSE extracts X.509 certificate from COSE protected header
func (v *COSEValidator) ExtractCertificateFromCOSE(protected cose.ProtectedHeader) (*x509.Certificate, error) {
	// COSE uses header label 33 (x5chain) for certificate chain
	// ProtectedHeader is a map[interface{}]interface{}
	x5chainValue, ok := protected[cose.HeaderLabelX5Chain]
	if !ok {
		return nil, fmt.Errorf("no x5chain (label 33) in COSE protected header")
	}

	// x5chain can be a single certificate or an array
	switch chain := x5chainValue.(type) {
	case []byte:
		// Single certificate as byte string
		cert, err := x509.ParseCertificate(chain)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		return cert, nil

	case []interface{}:
		// Array of certificates
		if len(chain) == 0 {
			return nil, fmt.Errorf("empty certificate chain")
		}

		// First cert in chain is the signing certificate
		certDER, ok := chain[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid certificate format in chain")
		}

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		return cert, nil

	default:
		return nil, fmt.Errorf("unexpected x5chain type: %T", chain)
	}
}

// ExtractAlgorithm extracts the algorithm from COSE protected header
func (v *COSEValidator) ExtractAlgorithm(protected cose.ProtectedHeader) (cose.Algorithm, error) {
	algValue, ok := protected[cose.HeaderLabelAlgorithm]
	if !ok {
		return 0, fmt.Errorf("no algorithm in COSE protected header")
	}

	alg, ok := algValue.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid algorithm type: %T", algValue)
	}

	return cose.Algorithm(alg), nil
}

// GetPublicKeyFromCOSEKey extracts a public key from COSE_Key structure
func (v *COSEValidator) GetPublicKeyFromCOSEKey(coseKey map[interface{}]interface{}) (crypto.PublicKey, error) {
	// COSE_Key structure:
	// 1 (kty): Key Type (2 = EC)
	// -1 (crv): Curve (1 = P-256, 2 = P-384, 3 = P-521)
	// -2 (x): x coordinate
	// -3 (y): y coordinate

	ktyValue, ok := coseKey[int64(1)]
	if !ok {
		return nil, fmt.Errorf("missing kty in COSE_Key")
	}

	kty, ok := ktyValue.(int64)
	if !ok {
		return nil, fmt.Errorf("invalid kty type: %T", ktyValue)
	}

	// Only support EC keys for now (kty = 2)
	if kty != 2 {
		return nil, fmt.Errorf("unsupported key type: %d", kty)
	}

	crvValue, ok := coseKey[int64(-1)]
	if !ok {
		return nil, fmt.Errorf("missing crv in COSE_Key")
	}

	crv, ok := crvValue.(int64)
	if !ok {
		return nil, fmt.Errorf("invalid crv type: %T", crvValue)
	}

	xValue, ok := coseKey[int64(-2)]
	if !ok {
		return nil, fmt.Errorf("missing x coordinate in COSE_Key")
	}

	x, ok := xValue.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid x coordinate type: %T", xValue)
	}

	yValue, ok := coseKey[int64(-3)]
	if !ok {
		return nil, fmt.Errorf("missing y coordinate in COSE_Key")
	}

	y, ok := yValue.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid y coordinate type: %T", yValue)
	}

	// Convert COSE curve to Go elliptic curve
	var curve ecdsa.PublicKey
	switch crv {
	case 1: // P-256
		curve.Curve = nil // Will be set by UnmarshalPKIXPublicKey or manually
		curve.X = new(big.Int).SetBytes(x)
		curve.Y = new(big.Int).SetBytes(y)
		curve.Curve = elliptic.P256()
	case 2: // P-384
		curve.X = new(big.Int).SetBytes(x)
		curve.Y = new(big.Int).SetBytes(y)
		curve.Curve = elliptic.P384()
	case 3: // P-521
		curve.X = new(big.Int).SetBytes(x)
		curve.Y = new(big.Int).SetBytes(y)
		curve.Curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %d", crv)
	}

	return &curve, nil
}
