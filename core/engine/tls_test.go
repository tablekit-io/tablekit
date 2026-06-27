package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTLSDisable(t *testing.T) {
	out, err := buildTLS(&tlsSettings{mode: "disable"}, "db.example")
	require.NoError(t, err)
	assert.Nil(t, out)
}

func TestBuildTLSModes(t *testing.T) {
	tests := []struct {
		mode          string
		skipVerify    bool
		hasVerifyConn bool
	}{
		{"allow", true, false},
		{"prefer", true, false},
		{"require", true, false},
		{"verify-full", false, false},
		{"verify-ca", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			out, err := buildTLS(&tlsSettings{mode: tt.mode}, "db.example")
			require.NoError(t, err)
			require.NotNil(t, out)
			assert.Equal(t, "db.example", out.ServerName)
			assert.Equal(t, tt.skipVerify, out.InsecureSkipVerify)
			assert.Equal(t, tt.hasVerifyConn, out.VerifyConnection != nil)
		})
	}
}

func TestBuildTLSNilDefaultsToPrefer(t *testing.T) {
	out, err := buildTLS(nil, "db.example")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, out.InsecureSkipVerify)
	assert.Equal(t, "db.example", out.ServerName)
}

func TestBuildTLSUnknownMode(t *testing.T) {
	_, err := buildTLS(&tlsSettings{mode: "bogus"}, "db.example")
	assert.Error(t, err)
}

func TestBuildTLSBadRootCert(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "ca.pem")
	require.NoError(t, os.WriteFile(bad, []byte("not a certificate"), 0o600))
	_, err := buildTLS(&tlsSettings{mode: "verify-full", rootCertFilePath: bad}, "db.example")
	assert.Error(t, err)
}

func TestBuildTLSMissingRootCert(t *testing.T) {
	_, err := buildTLS(&tlsSettings{mode: "verify-full", rootCertFilePath: "/no/such/ca.pem"}, "db.example")
	assert.Error(t, err)
}
