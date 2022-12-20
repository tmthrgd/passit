package passit

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestEncoding(t *testing.T) {
	assert.PanicsWithValue(t, "passit: count must be positive", func() {
		HexLower(-1)
	}, "HexLower(-1)")
	assert.PanicsWithValue(t, "passit: count must be positive", func() {
		Ascii85(-1)
	}, "Ascii85(-1)")

	for _, tc := range []struct {
		gen    func(int) Generator
		expect string
		count  int
	}{
		{HexLower, "", 0},
		{HexLower, "66e94bd4ef8a2c3b", 8},
		{HexLower, "66e94bd4ef8a2c3b884cfa", 11},
		{HexLower, "66e94bd4ef8a2c3b884cfa59ca342b2e58", 17},
		{HexUpper, "", 0},
		{HexUpper, "66E94BD4EF8A2C3B", 8},
		{HexUpper, "66E94BD4EF8A2C3B884CFA", 11},
		{HexUpper, "66E94BD4EF8A2C3B884CFA59CA342B2E58", 17},
		{Base32, "", 0},
		{Base32, "M3UUXVHPRIWDW", 8},
		{Base32, "M3UUXVHPRIWDXCCM7I", 11},
		{Base32, "M3UUXVHPRIWDXCCM7JM4UNBLFZMA", 17},
		{Base32Hex, "", 0},
		{Base32Hex, "CRKKNL7FH8M3M", 8},
		{Base32Hex, "CRKKNL7FH8M3N22CV8", 11},
		{Base32Hex, "CRKKNL7FH8M3N22CV9CSKD1B5PC0", 17},
		{Base64, "", 0},
		{Base64, "ZulL1O+KLDs", 8},
		{Base64, "ZulL1O+KLDuITPo", 11},
		{Base64, "ZulL1O+KLDuITPpZyjQrLlg", 17},
		{Base64URL, "", 0},
		{Base64URL, "ZulL1O-KLDs", 8},
		{Base64URL, "ZulL1O-KLDuITPo", 11},
		{Base64URL, "ZulL1O-KLDuITPpZyjQrLlg", 17},
		{Ascii85, "", 0},
		{Ascii85, "B'Dt<mtrYX", 8},
		{Ascii85, "B'Dt<mtrYXLeRX", 11},
		{Ascii85, "B'Dt<mtrYXLeRYJattV$=9", 17},
	} {
		tr := newTestRand()

		pass, err := tc.gen(tc.count).Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
