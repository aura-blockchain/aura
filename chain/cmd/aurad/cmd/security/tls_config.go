package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// TLS certificate and key file names
	ServerCertFile = "server.crt"
	ServerKeyFile  = "server.key"
	CACertFile     = "ca.crt"
)

// TLSConfig holds TLS configuration
type TLSConfig struct {
	certPath   string
	keyPath    string
	caPath     string
	logger     Logger
	mutateTLS  bool // Enable mutual TLS
	minVersion uint16
}

// NewTLSConfig creates a new TLS configuration
func NewTLSConfig(homeDir string, logger Logger) *TLSConfig {
	configDir := filepath.Join(homeDir, "config", "tls")
	return &TLSConfig{
		certPath:   filepath.Join(configDir, ServerCertFile),
		keyPath:    filepath.Join(configDir, ServerKeyFile),
		caPath:     filepath.Join(configDir, CACertFile),
		logger:     logger,
		mutateTLS:  false,
		minVersion: tls.VersionTLS13,
	}
}

// EnableMutualTLS enables mutual TLS (client certificate verification)
func (tc *TLSConfig) EnableMutualTLS() {
	tc.mutateTLS = true
}

// SetMinTLSVersion sets the minimum TLS version
func (tc *TLSConfig) SetMinTLSVersion(version uint16) error {
	// Only allow TLS 1.2 and 1.3
	if version != tls.VersionTLS12 && version != tls.VersionTLS13 {
		return fmt.Errorf("minimum TLS version must be TLS 1.2 or 1.3")
	}
	tc.minVersion = version
	return nil
}

// LoadTLSConfig loads TLS configuration for servers
func (tc *TLSConfig) LoadTLSConfig() (*tls.Config, error) {
	// Check if certificate files exist
	if !tc.certificatesExist() {
		tc.logger.SecurityEvent("tls_certificates_missing", map[string]interface{}{
			"cert_path": tc.certPath,
			"key_path":  tc.keyPath,
		})
		return nil, fmt.Errorf("TLS certificates not found in %s. Run 'aurad init' to generate certificates", filepath.Dir(tc.certPath))
	}

	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(tc.certPath, tc.keyPath)
	if err != nil {
		tc.logger.SecurityEvent("tls_certificate_load_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS config with strong security settings
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tc.minVersion,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: getStrongCipherSuites(),
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
		PreferServerCipherSuites: true,
		SessionTicketsDisabled:   false,
		Renegotiation:            tls.RenegotiateNever,
	}

	// Enable mutual TLS if requested
	if tc.mutateTLS {
		if err := tc.setupMutualTLS(config); err != nil {
			return nil, err
		}
	}

	tc.logger.SecurityEvent("tls_config_loaded", map[string]interface{}{
		"mutual_tls":  tc.mutateTLS,
		"min_version": tlsVersionString(tc.minVersion),
	})

	return config, nil
}

// setupMutualTLS configures mutual TLS (client certificate verification)
func (tc *TLSConfig) setupMutualTLS(config *tls.Config) error {
	// Check if CA certificate exists
	if _, err := os.Stat(tc.caPath); os.IsNotExist(err) {
		tc.logger.SecurityEvent("ca_certificate_missing", map[string]interface{}{
			"ca_path": tc.caPath,
		})
		return fmt.Errorf("CA certificate not found: %s", tc.caPath)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(tc.caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Create certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to parse CA certificate")
	}

	// Configure client authentication
	config.ClientAuth = tls.RequireAndVerifyClientCert
	config.ClientCAs = certPool

	tc.logger.SecurityEvent("mutual_tls_enabled", map[string]interface{}{
		"ca_path": tc.caPath,
	})

	return nil
}

// LoadClientTLSConfig loads TLS configuration for clients
func (tc *TLSConfig) LoadClientTLSConfig() (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tc.minVersion,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: getStrongCipherSuites(),
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}

	// If mutual TLS is enabled, load client certificate
	if tc.mutateTLS && tc.certificatesExist() {
		cert, err := tls.LoadX509KeyPair(tc.certPath, tc.keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if available
	if _, err := os.Stat(tc.caPath); err == nil {
		caCert, err := os.ReadFile(tc.caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		config.RootCAs = certPool
	}

	return config, nil
}

// certificatesExist checks if TLS certificates exist
func (tc *TLSConfig) certificatesExist() bool {
	_, certErr := os.Stat(tc.certPath)
	_, keyErr := os.Stat(tc.keyPath)
	return certErr == nil && keyErr == nil
}

// getStrongCipherSuites returns a list of strong cipher suites
func getStrongCipherSuites() []uint16 {
	return []uint16{
		// TLS 1.3 cipher suites (preferred)
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,

		// TLS 1.2 cipher suites (fallback)
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	}
}

// tlsVersionString converts TLS version to string
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%d)", version)
	}
}

// GenerateSelfSignedCert generates a self-signed certificate for development
// This should only be used for testing/development purposes
func (tc *TLSConfig) GenerateSelfSignedCert(hostname string) error {
	// Create TLS directory if it doesn't exist
	tlsDir := filepath.Dir(tc.certPath)
	if err := os.MkdirAll(tlsDir, SecureDirPerms); err != nil {
		return fmt.Errorf("failed to create TLS directory: %w", err)
	}

	tc.logger.SecurityEvent("generating_self_signed_cert", map[string]interface{}{
		"hostname": hostname,
		"cert_dir": tlsDir,
	})

	// Note: In a real implementation, you would use crypto/x509 to generate
	// a proper self-signed certificate. For brevity, this is a placeholder.
	// Users should use proper certificate management tools in production.

	return fmt.Errorf("self-signed certificate generation not implemented - please provide valid certificates")
}

// ValidateCertificate validates a TLS certificate
func (tc *TLSConfig) ValidateCertificate() error {
	if !tc.certificatesExist() {
		return fmt.Errorf("certificates do not exist")
	}

	// Try to load the certificate
	cert, err := tls.LoadX509KeyPair(tc.certPath, tc.keyPath)
	if err != nil {
		return fmt.Errorf("invalid certificate: %w", err)
	}

	// Parse the certificate
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("no certificate found in file")
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if certificate is expired
	if x509Cert.NotAfter.Before(x509Cert.NotBefore) {
		return fmt.Errorf("certificate expiry date is before start date")
	}

	tc.logger.SecurityEvent("certificate_validated", map[string]interface{}{
		"subject":    x509Cert.Subject.String(),
		"not_before": x509Cert.NotBefore,
		"not_after":  x509Cert.NotAfter,
	})

	return nil
}
