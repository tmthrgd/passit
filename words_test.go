package passit

import (
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
		{"and or or and and and and or", mustWords(t, "and", "or")},
		{"ευτυχία αιώνια αιώνια ελπίδα ελπίδα ευτυχία ευτυχία αιώνια", mustWords(t, "ελπίδα", "υγεία", "ευτυχία", "αιώνια")},
	} {
		testRand := newTestRand()

		pass, err := Repeat(tc.tmpl, " ", 8).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
