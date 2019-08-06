package password

import (
	"math/rand"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinTemplates(t *testing.T) {
	tmpl := JoinTemplates(
		FixedString("abc"),
		Space,
		FixedString("def"),
	)

	testRand := rand.New(rand.NewSource(0))

	pass, err := tmpl.Password(testRand)
	require.NoError(t, err)

	assert.Equal(t, "abc def", pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)
}
