package password

import (
	"math/rand"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCharset(t *testing.T, template string) func(int) Template {
	t.Helper()

	tmpl, err := NewCharset(template)
	require.NoError(t, err)
	return tmpl
}

func TestJoinTemplates(t *testing.T) {
	pattern := regexp.MustCompile(`^([a-z]+ ){5}[A-Z][0-9][~!@#$%^&*()] \+abc[de]$`)

	tmpl := JoinTemplates(
		DefaultWords(5),
		Space,
		mustCharset(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")(1),
		mustCharset(t, "0123456789")(1),
		mustCharset(t, "~!@#$%^&*()")(1),
		Space,
		FixedString("+abc"),
		mustCharset(t, "de")(1),
	)

	testRand := rand.New(rand.NewSource(0))

	pass, err := tmpl.Password(testRand)
	require.NoError(t, err)

	assert.Equal(t, "native remover dismay vocation sepia C2@ +abce", pass)
	assert.Truef(t, pattern.MatchString(pass),
		"regexp.MustCompile(%q).MatchString(%q)", pattern, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)
}
