package password

import (
	"math/rand"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestCharset(t *testing.T) {
	for _, tc := range []struct{ expect, template string }{
		{"1010000010111010000001100", "01"},
		{"1690822236719012868805980", "0123456789"},
		{"lwrqmcesfypbvqzagueycldeq", "abcdefghijklmnopqrstuvwxyz"},
		{"LWRQMCESFYPBVQZAGUEYCLDEQ", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"lWRQmCESfYpBVqZAGuEYcLDeq", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"},
		{"VwJIgWOitmbpXelQQossSXxYe", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"},
		{"lj$#+pr%%yc!iqmathelp_dr#", "abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+"},
		{"ρΘΙδΚΙσΧεΣσΕΒωπΠΟΠδΑΕλΚΥΦ", "ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω"},
		{"🍧🛰🍳🔱🚱👒🎩👒🍉🌴💻🍧🍳🐊🍧🎩🚱🛰💅💅🔱👗🚋🚱🐊", "🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋"},
	} {
		const size = 25

		testRand := rand.New(rand.NewSource(0))

		tmpl, err := NewCharset(tc.template)
		if !assert.NoError(t, err) {
			continue
		}

		pass, err := tmpl(size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
