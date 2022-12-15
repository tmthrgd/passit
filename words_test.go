package passit

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWords(t *testing.T) {
	mustWords := func(t *testing.T, list ...string) Generator {
		t.Helper()

		gen, err := FromWords(list...)
		require.NoError(t, err)
		return gen
	}

	for _, tc := range []struct {
		expect string
		gen    Generator
	}{
		{"       ", mustWords(t)},
		{"to to to to to to to to", mustWords(t, "to")},
		{"and or or and and and and or", mustWords(t, "and", "or")},
		{"ευτυχία αιώνια αιώνια ελπίδα ελπίδα ευτυχία ευτυχία αιώνια", mustWords(t, "ελπίδα", "υγεία", "ευτυχία", "αιώνια")},
	} {
		tr := newTestRand()

		pass, err := Repeat(tc.gen, " ", 8).Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
