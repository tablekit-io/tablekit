package store

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

// signingKeyMinLen is the minimum HS256 key length; shorter keys are zero-padded
// up to it (256-bit keys for HMAC-SHA256).
const signingKeyMinLen = 32

// DecodeSigningKey decodes a base64 signing key (standard or raw, with or
// without `=` padding) and zero-pads it up to signingKeyMinLen. Keys at or above
// the minimum are returned as-is (never truncated). Empty or invalid input is an
// error. Shared by the env-provided key path and the on-disk key file.
func DecodeSigningKey(b64 string) ([]byte, error) {
	raw, err := decodeBase64Tolerant(strings.TrimSpace(b64))
	if err != nil {
		return nil, fmt.Errorf("invalid base64 signing key: %w", err)
	}
	if len(raw) == 0 {
		return nil, errors.New("signing key is empty")
	}
	if len(raw) >= signingKeyMinLen {
		return raw, nil
	}
	padded := make([]byte, signingKeyMinLen)
	copy(padded, raw)
	return padded, nil
}

// decodeBase64Tolerant accepts standard and URL alphabets, with or without
// padding.
func decodeBase64Tolerant(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, errors.New("not valid base64")
}

// writeKeyFile persists raw key bytes as base64 text (0600).
func (s *Store) writeKeyFile(raw []byte) error {
	encoded := base64.StdEncoding.EncodeToString(raw)
	return os.WriteFile(s.path("signing.key"), []byte(encoded), 0o600)
}

// migrateLegacySigningKey runs at boot: a signing.key that is exactly 32 raw
// bytes is the pre-base64 format, so re-encode it in place. A missing file is
// left alone (an env key may supply the secret); anything else is assumed to be
// base64 already.
func (s *Store) migrateLegacySigningKey() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path("signing.key"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) == signingKeyMinLen {
		return s.writeKeyFile(b)
	}
	return nil
}

// SigningKey returns the HS256 secret from signing.key, generating and
// persisting a random 32-byte key (base64-encoded) on first call so tokens
// survive restarts with zero config. Legacy raw files are already normalized to
// base64 by migrateLegacySigningKey at boot, so this only sees base64 or absent.
func (s *Store) SigningKey() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path("signing.key"))
	if err == nil {
		return DecodeSigningKey(string(b))
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	key := make([]byte, signingKeyMinLen)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := s.writeKeyFile(key); err != nil {
		return nil, err
	}
	return key, nil
}
