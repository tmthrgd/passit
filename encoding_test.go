package passit

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestEncoding(t *testing.T) {
	assert.PanicsWithValue(t, "passit: count must be positive", func() {
		Hex(-1)
	}, "Hex(-1)")
	assert.PanicsWithValue(t, "passit: count must be positive", func() {
		Ascii85(-1)
	}, "Ascii85(-1)")

	for _, tc := range []struct {
		templ  func(int) Template
		expect string
		count  int
	}{
		{Hex, "", 0},
		{Hex, "66e94bd4ef8a2c3b", 8},
		{Hex, "66e94bd4ef8a2c3b884cfa", 11},
		{Hex, "66e94bd4ef8a2c3b884cfa59ca342b2e58", 17},
		{Base32Std, "", 0},
		{Base32Std, "M3UUXVHPRIWDW", 8},
		{Base32Std, "M3UUXVHPRIWDXCCM7I", 11},
		{Base32Std, "M3UUXVHPRIWDXCCM7JM4UNBLFZMA", 17},
		{Base32Hex, "", 0},
		{Base32Hex, "CRKKNL7FH8M3M", 8},
		{Base32Hex, "CRKKNL7FH8M3N22CV8", 11},
		{Base32Hex, "CRKKNL7FH8M3N22CV9CSKD1B5PC0", 17},
		{Base64Std, "", 0},
		{Base64Std, "ZulL1O+KLDs", 8},
		{Base64Std, "ZulL1O+KLDuITPo", 11},
		{Base64Std, "ZulL1O+KLDuITPpZyjQrLlg", 17},
		{Base64URL, "", 0},
		{Base64URL, "ZulL1O-KLDs", 8},
		{Base64URL, "ZulL1O-KLDuITPo", 11},
		{Base64URL, "ZulL1O-KLDuITPpZyjQrLlg", 17},
		{Ascii85, "", 0},
		{Ascii85, "B'Dt<mtrYX", 8},
		{Ascii85, "B'Dt<mtrYXLeRX", 11},
		{Ascii85, "B'Dt<mtrYXLeRYJattV$=9", 17},
	} {
		testRand := newTestRand()

		pass, err := tc.templ(tc.count).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}
