package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyPKCE(t *testing.T) {
	tests := []struct {
		name      string
		verifier  string
		challenge string
		want      bool
	}{
		{
			// RFC 7636 Appendix B reference vector.
			name:      "rfc7636 vector",
			verifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			challenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			want:      true,
		},
		{
			name:      "wrong verifier",
			verifier:  "not-the-right-verifier",
			challenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			want:      false,
		},
		{
			name:      "empty verifier",
			verifier:  "",
			challenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, VerifyPKCE(tt.verifier, tt.challenge))
		})
	}
}
