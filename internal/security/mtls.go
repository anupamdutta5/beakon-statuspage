package security

import (
	"context"
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
	"time"
)

// mTLSConfig represents mTLS configuration
type mTLSConfig struct {
	CAKeyPath      string        `json:"ca_key_path"`
	CACertPath     string        `json:"ca_cert_path"`
	ServerKeyPath  string        `json:"server_key_path"`
	ServerCertPath string        `json:"server_cert_path"`
	ClientKeyPath  string        `json:"client_key_path"`
	ClientCertPath string        `json:"client_cert_path"`
	CertValidity   time.Duration `json:"cert_validity"`
	KeySize        int           `json:"key_size"`
}

// mTLSManager handles mTLS operations
type mTLSManager struct {
	config mTLSConfig
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

// NewmTLSManager creates a new mTLS manager
func NewmTLSManager(config mTLSConfig) (*mTLSManager, error) {
	// Set default values
	if config.CertValidity == 0 {
		config.CertValidity = 365 * 24 * time.Hour // 1 year
	}
	if config.KeySize == 0 {
		config.KeySize = 2048
	}

	manager := &mTLSManager{
		config: config,
	}

	// Load or generate CA
	if config.CAKeyPath != "" && config.CACertPath != "" {
		err := manager.loadCA()
		if err != nil {
			return nil, fmt.Errorf("failed to load CA: %w", err)
		}
	} else {
		err := manager.generateCA()
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA: %w", err)
		}
	}

	return manager, nil
}

// loadCA loads CA certificate and key from files
func (m *mTLSManager) loadCA() error {
	// Load CA certificate
	caCertPEM, err := loadPEMFile(m.config.CACertPath)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	m.caCert = caCert

	// Load CA key
	caKeyPEM, err := loadPEMFile(m.config.CAKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load CA key: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return fmt.Errorf("failed to decode CA key PEM")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA key: %w", err)
	}

	m.caKey = caKey

	return nil
}

// generateCA generates a new CA certificate and key
func (m *mTLSManager) generateCA() error {
	// Generate CA key
	caKey, err := rsa.GenerateKey(rand.Reader, m.config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"StatusPage CA"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(m.config.CertValidity),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// Create CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	m.caCert = caCert
	m.caKey = caKey

	return nil
}

// GenerateServerCert generates a server certificate
func (m *mTLSManager) GenerateServerCert(hostname string) (*tls.Certificate, error) {
	// Generate server key
	serverKey, err := rsa.GenerateKey(rand.Reader, m.config.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key: %w", err)
	}

	// Create server certificate template
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:  []string{"StatusPage Server"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(m.config.CertValidity),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{hostname, "localhost"},
	}

	// Create server certificate
	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, m.caCert, &serverKey.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Create TLS certificate
	tlsCert := &tls.Certificate{
		Certificate: [][]byte{serverCertDER},
		PrivateKey:  serverKey,
	}

	return tlsCert, nil
}

// GenerateClientCert generates a client certificate
func (m *mTLSManager) GenerateClientCert(clientID string) (*tls.Certificate, error) {
	// Generate client key
	clientKey, err := rsa.GenerateKey(rand.Reader, m.config.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client key: %w", err)
	}

	// Create client certificate template
	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization:  []string{"StatusPage Client"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    clientID,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(m.config.CertValidity),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Create client certificate
	clientCertDER, err := x509.CreateCertificate(rand.Reader, &clientTemplate, m.caCert, &clientKey.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate: %w", err)
	}

	// Create TLS certificate
	tlsCert := &tls.Certificate{
		Certificate: [][]byte{clientCertDER},
		PrivateKey:  clientKey,
	}

	return tlsCert, nil
}

// GetServerTLSConfig returns TLS configuration for server
func (m *mTLSManager) GetServerTLSConfig(hostname string) (*tls.Config, error) {
	serverCert, err := m.GenerateServerCert(hostname)
	if err != nil {
		return nil, err
	}

	// Create certificate pool with CA
	certPool := x509.NewCertPool()
	certPool.AddCert(m.caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}

// GetClientTLSConfig returns TLS configuration for client
func (m *mTLSManager) GetClientTLSConfig(clientID string) (*tls.Config, error) {
	clientCert, err := m.GenerateClientCert(clientID)
	if err != nil {
		return nil, err
	}

	// Create certificate pool with CA
	certPool := x509.NewCertPool()
	certPool.AddCert(m.caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{*clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}

// ValidateClientCert validates a client certificate
func (m *mTLSManager) ValidateClientCert(clientCert *x509.Certificate) error {
	// Verify certificate against CA
	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
	}
	opts.Roots.AddCert(m.caCert)

	_, err := clientCert.Verify(opts)
	if err != nil {
		return fmt.Errorf("failed to verify client certificate: %w", err)
	}

	// Check certificate validity
	if time.Now().Before(clientCert.NotBefore) || time.Now().After(clientCert.NotAfter) {
		return fmt.Errorf("client certificate is not valid")
	}

	// Check key usage
	if clientCert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("client certificate missing digital signature key usage")
	}

	return nil
}

// GetCACert returns the CA certificate in PEM format
func (m *mTLSManager) GetCACert() ([]byte, error) {
	if m.caCert == nil {
		return nil, fmt.Errorf("CA certificate not available")
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: m.caCert.Raw,
	}), nil
}

// SaveCACert saves the CA certificate to a file
func (m *mTLSManager) SaveCACert(filename string) error {
	caCertPEM, err := m.GetCACert()
	if err != nil {
		return err
	}

	return savePEMFile(filename, caCertPEM)
}

// SaveCAKey saves the CA key to a file
func (m *mTLSManager) SaveCAKey(filename string) error {
	if m.caKey == nil {
		return fmt.Errorf("CA key not available")
	}

	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(m.caKey),
	})

	return savePEMFile(filename, caKeyPEM)
}

// mTLSMiddleware provides HTTP middleware for mTLS
type mTLSMiddleware struct {
	manager *mTLSManager
}

// NewmTLSMiddleware creates a new mTLS middleware
func NewmTLSMiddleware(manager *mTLSManager) *mTLSMiddleware {
	return &mTLSMiddleware{
		manager: manager,
	}
}

// ClientCertHandler extracts client certificate from request
func (mm *mTLSMiddleware) ClientCertHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if TLS connection exists
		if r.TLS == nil {
			http.Error(w, "TLS connection required", http.StatusBadRequest)
			return
		}

		// Check if client certificate exists
		if len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "Client certificate required", http.StatusUnauthorized)
			return
		}

		// Validate client certificate
		clientCert := r.TLS.PeerCertificates[0]
		if err := mm.manager.ValidateClientCert(clientCert); err != nil {
			http.Error(w, "Invalid client certificate", http.StatusUnauthorized)
			return
		}

		// Add client certificate info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "client_cert", clientCert)
		ctx = context.WithValue(ctx, "client_id", clientCert.Subject.CommonName)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Helper functions for file operations
func loadPEMFile(filename string) ([]byte, error) {
	// In a real implementation, you would read from file
	// For now, return empty bytes
	return []byte{}, nil
}

func savePEMFile(filename string, data []byte) error {
	// In a real implementation, you would write to file
	// For now, just return nil
	return nil
}

// Global mTLS manager instance
var globalmTLSManager *mTLSManager

// InitGlobalmTLS initializes the global mTLS manager
func InitGlobalmTLS(config mTLSConfig) error {
	manager, err := NewmTLSManager(config)
	if err != nil {
		return err
	}
	globalmTLSManager = manager
	return nil
}

// GetGlobalmTLS returns the global mTLS manager
func GetGlobalmTLS() *mTLSManager {
	return globalmTLSManager
}

// GetServerTLSConfig is a convenience function for getting server TLS config
func GetServerTLSConfig(hostname string) (*tls.Config, error) {
	if globalmTLSManager != nil {
		return globalmTLSManager.GetServerTLSConfig(hostname)
	}
	return nil, fmt.Errorf("mTLS manager not initialized")
}

// GetClientTLSConfig is a convenience function for getting client TLS config
func GetClientTLSConfig(clientID string) (*tls.Config, error) {
	if globalmTLSManager != nil {
		return globalmTLSManager.GetClientTLSConfig(clientID)
	}
	return nil, fmt.Errorf("mTLS manager not initialized")
}
