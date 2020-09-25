package password

import (
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustWords(t *testing.T, list ...string) func(int) Template {
	t.Helper()

	tmpl, err := FromWords(list...)
	require.NoError(t, err)
	return tmpl
}

func TestWords(t *testing.T) {
	for _, tc := range []struct {
		expect string
		tmpl   func(int) Template
	}{
		{"or and or and and and and and", mustWords(t, "and", "or")},
		{"υγεία ευτυχία υγεία ελπίδα ευτυχία ευτυχία ελπίδα ευτυχία", mustWords(t, "ελπίδα", "υγεία", "ευτυχία", "αιώνια")},
		{"native remover dismay vocation sepia backtalk think conjure", EFFLargeWordlist},
	} {
		const size = 8

		testRand := rand.New(rand.NewSource(0))

		pass, err := tc.tmpl(size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size-1, strings.Count(pass, " "),
			`strings.Count(%q, " ")`, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, pass)
	}
}

func TestDefaultWordsValid(t *testing.T) {
	EFFLargeWordlist(1) // Initialise effLargeWordlistVal.list.

	_, err := FromWords(effLargeWordlistVal.list...)
	assert.NoError(t, err)
}
