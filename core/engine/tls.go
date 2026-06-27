package engine

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// buildTLS turns the configured TLS settings into a *tls.Config, or nil when
// TLS is disabled. serverName must be the real database host even when the
// connection is dialled through a local SSH tunnel, so SNI and certificate
// hostname checks target the right name.
//
// sslmode mapping:
//   - disable                  -> no TLS
//   - allow / prefer / require -> TLS on, no verification
//   - verify-ca                -> verify the certificate chain, but not hostname
//   - verify-full              -> verify the chain AND hostname
//
// Note: prefer/allow do not replicate libpq's silent plaintext fallback; they
// require TLS without verification.
func buildTLS(cfg *tlsSettings, serverName string) (*tls.Config, error) {
	mode := "prefer"
	if cfg != nil && cfg.mode != "" {
		mode = cfg.mode
	}
	if mode == "disable" {
		return nil, nil
	}

	out := &tls.Config{ServerName: serverName}
	switch mode {
	case "allow", "prefer", "require":
		out.InsecureSkipVerify = true
	case "verify-full":
		// Default Go verification: chain + hostname against ServerName.
	case "verify-ca":
		// Verify the chain ourselves but skip the hostname check.
		out.InsecureSkipVerify = true
	default:
		return nil, fmt.Errorf("unknown tls mode %q", mode)
	}

	if cfg != nil && cfg.rootCertFilePath != "" {
		pem, err := os.ReadFile(cfg.rootCertFilePath)
		if err != nil {
			return nil, fmt.Errorf("read root cert %q: %w", cfg.rootCertFilePath, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("root cert file contained no valid certificates")
		}
		out.RootCAs = pool
	}

	if cfg != nil && cfg.clientCertFilePath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.clientCertFilePath, cfg.clientKeyFilePath)
		if err != nil {
			return nil, fmt.Errorf("load client cert/key: %w", err)
		}
		out.Certificates = []tls.Certificate{cert}
	}

	if mode == "verify-ca" {
		roots := out.RootCAs
		out.VerifyConnection = func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return errors.New("server presented no certificate")
			}
			intermediates := x509.NewCertPool()
			for _, cert := range state.PeerCertificates[1:] {
				intermediates.AddCert(cert)
			}
			_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{
				Roots:         roots,
				Intermediates: intermediates,
			})
			return err
		}
	}

	return out, nil
}
