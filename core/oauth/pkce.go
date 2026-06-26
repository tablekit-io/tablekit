package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

// verifyPKCE reports whether codeVerifier matches codeChallenge under the S256
// method: BASE64URL(SHA256(verifier)) == challenge. Constant-time compare.
func verifyPKCE(codeVerifier, codeChallenge string) bool {
	sum := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(codeChallenge)) == 1
}
