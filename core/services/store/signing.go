package store

import (
	"encoding/base64"
	"errors"
	"fmt"
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
