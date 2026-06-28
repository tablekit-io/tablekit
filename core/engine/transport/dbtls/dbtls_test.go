package dbtls

import (
	"os"
	"path/filepath"
	"testing"

	"core/engine/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfigDisable(t *testing.T) {
	out, err := BuildConfig(&config.TLSSettings{Mode: "disable"}, "db.example")
	require.NoError(t, err)
	assert.Nil(t, out)
}

func TestBuildConfigModes(t *testing.T) {
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
			out, err := BuildConfig(&config.TLSSettings{Mode: tt.mode}, "db.example")
			require.NoError(t, err)
			require.NotNil(t, out)
			assert.Equal(t, "db.example", out.ServerName)
			assert.Equal(t, tt.skipVerify, out.InsecureSkipVerify)
			assert.Equal(t, tt.hasVerifyConn, out.VerifyConnection != nil)
		})
	}
}

func TestBuildConfigNilDefaultsToPrefer(t *testing.T) {
	out, err := BuildConfig(nil, "db.example")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, out.InsecureSkipVerify)
	assert.Equal(t, "db.example", out.ServerName)
}

func TestBuildConfigUnknownMode(t *testing.T) {
	_, err := BuildConfig(&config.TLSSettings{Mode: "bogus"}, "db.example")
	assert.Error(t, err)
}

func TestBuildConfigBadRootCert(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "ca.pem")
	require.NoError(t, os.WriteFile(bad, []byte("not a certificate"), 0o600))
	_, err := BuildConfig(&config.TLSSettings{Mode: "verify-full", RootCertFilePath: bad}, "db.example")
	assert.Error(t, err)
}

func TestBuildConfigMissingRootCert(t *testing.T) {
	_, err := BuildConfig(&config.TLSSettings{Mode: "verify-full", RootCertFilePath: "/no/such/ca.pem"}, "db.example")
	assert.Error(t, err)
}
