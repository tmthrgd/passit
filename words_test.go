package passit

import (
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustWords(t *testing.T, list ...string) Template {
	t.Helper()

	tmpl, err := FromWords(list...)
	require.NoError(t, err)
	return tmpl
}

func TestWords(t *testing.T) {
	for _, tc := range []struct {
		expect string
		tmpl   Template
	}{
		{"or and or and and and and and", mustWords(t, "and", "or")},
		{"υγεία ευτυχία υγεία ελπίδα ευτυχία ευτυχία ελπίδα ευτυχία", mustWords(t, "ελπίδα", "υγεία", "ευτυχία", "αιώνια")},
	} {
		const size = 8

		testRand := rand.New(rand.NewSource(0))

		pass, err := Repeat(tc.tmpl, " ", size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size-1, strings.Count(pass, " "),
			`strings.Count(%q, " ")`, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
