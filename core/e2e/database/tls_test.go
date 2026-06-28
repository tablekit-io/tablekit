package database

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"core/e2e/harness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// genServerCert builds a throwaway CA and a server certificate whose SAN is
// dnsName (the DB container's name on the network, which is also the TLS
// ServerName the engine verifies). Returns CA, server cert and server key PEMs.
func genServerCert(t *testing.T, dnsName string) (caPEM, certPEM, keyPEM []byte) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "tablekit-e2e-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	caCert, err := x509.ParseCertificate(caDER)
	require.NoError(t, err)
	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: dnsName},
		DNSNames:     []string{dnsName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	require.NoError(t, err)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(serverKey)
	require.NoError(t, err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	return caPEM, certPEM, keyPEM
}

// writeTLSYaml writes a verify-full database config for a single TLS-enabled DB.
func writeTLSYaml(t *testing.T, engine, host string, port int, database, username, caPath string) string {
	t.Helper()
	yaml := fmt.Sprintf(`databases:
  target:
    type: %s
    details:
      host: %s
      port: %d
      database: %s
      username: %s
      password: pw
    tls:
      mode: verify-full
      rootCertFilePath: %s
`, engine, host, port, database, username, caPath)
	path := filepath.Join(t.TempDir(), "databases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

// runTLSSimpleQuery starts the server against the TLS DB and runs a simple query
// over the verify-full connection.
func runTLSSimpleQuery(t *testing.T, configPath string) {
	t.Helper()
	server := harness.StartServerEnv(t, "DATABASES_FILE="+configPath)
	_, token := harness.GenerateToken(t, server)
	session, err := harness.Connect(t, server.AppURL, harness.BearerClient(token))
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	result, isErr := callRunSQL(t, session, "target", "SELECT 1 AS one")
	require.False(t, isErr, "simple query over TLS should succeed")
	require.Equal(t, 1, result.RowCount)
	assert.Contains(t, result.Columns, "one")
}

// TestTLSPostgres: run_sql reaches a TLS-only postgres with verify-full.
func TestTLSPostgres(t *testing.T) {
	harness.RequireDocker(t)
	harness.EnsureImage(t, "tablekit-e2e-pgtls:latest", filepath.Join(dbDir(), "containers", "pgtls"))

	name := harness.UniqueName("pgtls")
	caPEM, certPEM, keyPEM := genServerCert(t, name)

	harness.RunContainer(t, harness.ContainerSpec{
		Name:  name,
		Image: "tablekit-e2e-pgtls:latest",
		Env: []string{
			"POSTGRES_PASSWORD=pw",
			"POSTGRES_DB=app",
			"TLS_CERT=" + base64.StdEncoding.EncodeToString(certPEM),
			"TLS_KEY=" + base64.StdEncoding.EncodeToString(keyPEM),
		},
		Tmpfs: []string{"/var/lib/postgresql/data"},
	})
	// psql (not pg_isready) so we only proceed once the real TLS server is up
	// with the target db (pg_isready reports ready during the init temp server).
	harness.WaitContainerReady(t, name, 60*time.Second, "psql", "-U", "postgres", "-d", "app", "-c", "SELECT 1")

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0o600))
	configPath := writeTLSYaml(t, "postgres", name, 5432, "app", "postgres", caPath)
	runTLSSimpleQuery(t, configPath)
}

// TestTLSMySQL: run_sql reaches a TLS-required mysql with verify-full.
func TestTLSMySQL(t *testing.T) {
	harness.RequireDocker(t)
	harness.EnsureImage(t, "tablekit-e2e-mysqltls:latest", filepath.Join(dbDir(), "containers", "mysqltls"))

	name := harness.UniqueName("mytls")
	caPEM, certPEM, keyPEM := genServerCert(t, name)

	harness.RunContainer(t, harness.ContainerSpec{
		Name:  name,
		Image: "tablekit-e2e-mysqltls:latest",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=pw",
			"MYSQL_DATABASE=app",
			"TLS_CERT=" + base64.StdEncoding.EncodeToString(certPEM),
			"TLS_KEY=" + base64.StdEncoding.EncodeToString(keyPEM),
		},
		Tmpfs: []string{"/var/lib/mysql"},
	})
	// Probe over TCP with TLS required, not the local socket: under load the
	// socket accepts connections before mysqld is serving TLS over TCP, which is
	// what the engine actually connects on.
	harness.WaitContainerReady(t, name, 90*time.Second,
		"mysql", "-uroot", "-ppw", "-h", "127.0.0.1", "--ssl-mode=REQUIRED", "-e", "SELECT 1")

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0o600))
	configPath := writeTLSYaml(t, "mysql", name, 3306, "app", "root", caPath)
	runTLSSimpleQuery(t, configPath)
}
