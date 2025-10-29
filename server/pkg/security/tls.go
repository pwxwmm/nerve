// Package security provides TLS/HTTPS configuration and management functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// TLSServer manages TLS configuration
type TLSServer struct {
	CertFile string
	KeyFile  string
	Config   *tls.Config
}

// NewTLSServer creates a new TLS server configuration
func NewTLSServer(certFile, keyFile string) *TLSServer {
	return &TLSServer{
		CertFile: certFile,
		KeyFile:  keyFile,
	}
}

// LoadTLSConfig loads TLS configuration from files
func (ts *TLSServer) LoadTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(ts.CertFile, ts.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	ts.Config = config
	return config, nil
}

// GenerateSelfSignedCert generates a self-signed certificate for development
func (ts *TLSServer) GenerateSelfSignedCert(host string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Nerve"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{host, "localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Create certificate file
	certOut, err := os.Create(ts.CertFile)
	if err != nil {
		return fmt.Errorf("failed to open cert file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write cert: %v", err)
	}

	// Create private key file
	keyOut, err := os.Create(ts.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to open key file: %v", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return fmt.Errorf("failed to write key: %v", err)
	}

	return nil
}

// SetupTLS sets up TLS configuration
func (ts *TLSServer) SetupTLS() error {
	// Check if certificate files exist
	if _, err := os.Stat(ts.CertFile); os.IsNotExist(err) {
		if _, err := os.Stat(ts.KeyFile); os.IsNotExist(err) {
			// Generate self-signed certificate
			fmt.Printf("Generating self-signed certificate for development...\n")
			if err := ts.GenerateSelfSignedCert("localhost"); err != nil {
				return fmt.Errorf("failed to generate self-signed certificate: %v", err)
			}
		}
	}

	// Load TLS configuration
	_, err := ts.LoadTLSConfig()
	return err
}

// GetTLSConfig returns the TLS configuration
func (ts *TLSServer) GetTLSConfig() *tls.Config {
	return ts.Config
}

// ClientTLSConfig creates a TLS configuration for client connections
func ClientTLSConfig(insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
}

// TLSMiddleware creates a middleware for HTTPS redirect
func TLSMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			c.Next()
			return
		}

		if c.Request.TLS == nil {
			// Redirect HTTP to HTTPS
			httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(http.StatusMovedPermanently, httpsURL)
			return
		}

		c.Next()
	}
}

