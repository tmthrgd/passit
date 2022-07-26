package passit

import (
	"math/rand"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestEncoding(t *testing.T) {
	for _, tc := range []struct {
		templ  func(int) Template
		expect string
		count  int
	}{
		{Hex, "0194fdc2fa2ffcc0", 8},
		{Hex, "0194fdc2fa2ffcc041d3ff", 11},
		{Hex, "0194fdc2fa2ffcc041d3ff12045b73c86e", 17},
		{Base32Std, "AGKP3QX2F76MA", 8},
		{Base32Std, "AGKP3QX2F76MAQOT74", 11},
		{Base32Std, "AGKP3QX2F76MAQOT74JAIW3TZBXA", 17},
		{Base32Hex, "06AFRGNQ5VUC0", 8},
		{Base32Hex, "06AFRGNQ5VUC0GEJVS", 11},
		{Base32Hex, "06AFRGNQ5VUC0GEJVS908MRJP1N0", 17},
		{Base64Std, "AZT9wvov/MA", 8},
		{Base64Std, "AZT9wvov/MBB0/8", 11},
		{Base64Std, "AZT9wvov/MBB0/8SBFtzyG4", 17},
		{Base64URL, "AZT9wvov_MA", 8},
		{Base64URL, "AZT9wvov_MBB0_8", 11},
		{Base64URL, "AZT9wvov_MBB0_8SBFtzyG4", 17},
		{Ascii85, "!L3Q\"qChc^", 8},
		{Ascii85, "!L3Q\"qChc^6.>i", 11},
		{Ascii85, "!L3Q\"qChc^6.>iH\"C#rgD?", 17},
	} {
		testRand := rand.New(rand.NewSource(0))

		pass, err := tc.templ(tc.count).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, pass)
	}
}