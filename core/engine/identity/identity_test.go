package identity_test

import (
	"testing"

	"core/engine/identity"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeReplacesSeparatorsAndSpecials(t *testing.T) {
	assert.Equal(t, "abc-123_._", identity.Sanitize("abc-123_._"), "allowed characters pass through")
	assert.Equal(t, "a_b_c", identity.Sanitize("a b c"), "spaces become underscores")
	assert.Equal(t, "7411_9c", identity.Sanitize("7411:9c"), "colons become underscores")
}

func TestEqualComparesKeys(t *testing.T) {
	a := identity.Identity{Engine: "postgres", Key: "pg-1-2"}
	b := identity.Identity{Engine: "postgres", Key: "pg-1-2", Attributes: map[string]string{"x": "y"}}
	c := identity.Identity{Engine: "postgres", Key: "pg-1-3"}

	assert.True(t, a.Equal(b), "same key is equal regardless of attributes")
	assert.False(t, a.Equal(c), "different key is not equal")
}

func TestEqualEmptyKeyNeverMatches(t *testing.T) {
	empty := identity.Identity{}
	assert.False(t, empty.Equal(identity.Identity{}), "an empty key must not compare equal to another empty key")
}
