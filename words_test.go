package passit

import (
	"math/rand"
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
		{"       ", mustWords(t)},
		{"to to to to to to to to", mustWords(t, "to")},
		{"or or and and or or and or", mustWords(t, "and", "or")},
		{"υγεία υγεία ευτυχία ελπίδα υγεία αιώνια ελπίδα αιώνια", mustWords(t, "ελπίδα", "υγεία", "ευτυχία", "αιώνια")},
	} {
		testRand := rand.New(rand.NewSource(0))

		pass, err := Repeat(tc.tmpl, " ", 8).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
