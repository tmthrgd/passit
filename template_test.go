package passit

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"math/bits"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRand returns a deterministic CSPRNG for testing use only.
func newTestRand() io.Reader {
	var key [16]byte
	var iv [aes.BlockSize]byte
	block, _ := aes.NewCipher(key[:])
	ctr := cipher.NewCTR(block, iv[:])
	sr := cipher.StreamReader{S: ctr, R: zeroReader{}}
	return bufio.NewReader(sr)
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func mustCharset(t *testing.T, template string) Template {
	t.Helper()

	tmpl, err := FromCharset(template)
	require.NoError(t, err)
	return tmpl
}

func TestJoin(t *testing.T) {
	{
		pattern := regexp.MustCompile(`^([a-z]+ ){5}[A-Z][0-9][~!@#$%^&*()] \+abc-[de]$`)

		tmpl := Join("",
			Repeat(EFFLargeWordlist, " ", 5),
			Space,
			LatinUpper,
			Number,
			mustCharset(t, "~!@#$%^&*()"),
			Space,
			FixedString("+abc"),
			Hyphen,
			mustCharset(t, "de"),
		)

		testRand := newTestRand()

		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, "reprint wool pantry unworried mummify Y4% +abc-d", pass)
		assert.Truef(t, pattern.MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", pattern, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}

	{
		tmpl := Join("@$", LatinUpper, LatinLower, LatinMixed)

		testRand := newTestRand()

		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, "C@$h@$Z", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

func TestRepeat(t *testing.T) {
	assert.PanicsWithValue(t, "passit: count must be positive", func() {
		Repeat(Hyphen, " ", -1)
	})

	assert.Equal(t, Empty, Repeat(Hyphen, " ", 0),
		"Repeat with count zero should return Empty")

	assert.Equal(t, Hyphen, Repeat(Hyphen, " ", 1),
		"Repeat with count one should return Template")

	for _, tc := range []struct {
		count  int
		sep    string
		expect string
	}{
		{2, " ", "reprint wool"},
		{4, "", "reprintwoolpantryunworried"},
		{12, "-", "reprint-wool-pantry-unworried-mummify-veneering-securely-munchkin-juiciness-steep-cresting-dastardly"},
	} {
		testRand := newTestRand()

		pass, err := Repeat(EFFLargeWordlist, tc.sep, tc.count).Password(testRand)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}

func TestRandomRepeat(t *testing.T) {
	const maxInt = 1<<(bits.UintSize-1) - 1

	_, err := RandomRepeat(Hyphen, " ", 10, 7)
	assert.EqualError(t, err, "passit: min argument cannot be greater than max argument",
		"min greater than max")

	_, err = RandomRepeat(Hyphen, " ", -5, 7)
	assert.EqualError(t, err, "passit: min argument must be positive",
		"min negative")

	_, err = RandomRepeat(Hyphen, " ", 5, -7)
	assert.EqualError(t, err, "passit: min argument cannot be greater than max argument",
		"min greater than max; max negative")

	for _, tc := range [][2]int{
		{0, maxReadIntN},
		{0, maxInt},
	} {
		_, err = RandomRepeat(Hyphen, " ", tc[0], tc[1])
		assert.EqualErrorf(t, err, "passit: [min,max] range too large",
			"out of range: %v", tc)
	}

	tmpl, err := RandomRepeat(Hyphen, " ", 0, 0)
	if assert.NoError(t, err, "min and max equal zero should not error") {
		assert.Equal(t, Empty, tmpl,
			"min and max equal zero should return Empty")
	}

	tmpl, err = RandomRepeat(Hyphen, " ", 1, 1)
	if assert.NoError(t, err, "min and max equal one should not error") {
		assert.Equal(t, Hyphen, tmpl,
			"min and max equal one should return template")
	}

	for _, tc := range []int{
		70,
		maxReadIntN,
		maxInt,
	} {
		tmpl, err := RandomRepeat(Hyphen, " ", tc, tc)
		if !assert.NoErrorf(t, err, "equal min and max should not error: %v", tc) {
			continue
		}
		assert.IsTypef(t, (*repeated)(nil), tmpl,
			"equal min and max should return Repeat(...): %v", tc)
	}

	for _, tc := range []struct {
		min, max int
		sep      string
		expect   string
	}{
		{1, 2, " ", "wool"},
		{2, 5, "", "woolpantryunworriedmummify"},
		{4, 7, "-", "wool-pantry-unworried-mummify-veneering-securely"},
		{10, 20, " ", "wool pantry unworried mummify veneering securely munchkin juiciness steep cresting dastardly cubical thriving procreate voice lingo stargazer acetone stroller"},
	} {
		tmpl, err := RandomRepeat(EFFLargeWordlist, tc.sep, tc.min, tc.max)
		if !assert.NoErrorf(t, err, "valid range should not error: %v", tc) {
			continue
		}

		testRand := newTestRand()

		pass, err := tmpl.Password(testRand)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}

func TestAlternate(t *testing.T) {
	assert.Equal(t, Empty, Alternate(),
		"Alternate with no templates should return Empty")

	assert.Equal(t, Hyphen, Alternate(Hyphen),
		"Alternate with single template should return Template")

	for _, tc := range []struct {
		tmpls  []Template
		expect string
	}{
		{[]Template{}, ""},
		{[]Template{LatinLower}, "c"},
		{[]Template{LatinLower, LatinUpper, Number}, "7"},
		{[]Template{EFFShortWordlist1, EFFShortWordlist2, EFFLargeWordlist}, "wool"},
		{[]Template{
			Repeat(LatinLower, "!", 5),
			Repeat(LatinUpper, "@", 3),
			Repeat(Number, "#", 7),
		}, "7#7#8#2#4#4#9"},
	} {
		testRand := newTestRand()

		pass, err := Alternate(tc.tmpls...).Password(testRand)
		if !assert.NoErrorf(t, err, "should not error when generating: %#v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "expected password: %#v", tc)
	}
}

func TestRejectionSample(t *testing.T) {
	rs := RejectionSample(Repeat(LatinMixedNumber, "", 20), func(s string) bool {
		return strings.Contains(s, "A") && strings.Contains(s, "0")
	})
	testRand := newTestRand()

	for _, expect := range []string{
		"l0LXpszA2lAxxyDUjT8o", // 3
		"0ZATYpv8h9K3YpeGjsbA", // 3
		"0ASqickBv1L0WdGukXJ1", // 7
		"aAA1sxVrP0jGibFTVp2T", // 7
		"ETeNP2gjuMyU50DbHtOA", // 18
		"CyglA0KaUPFUvhzRO9DV", // 3
		"izcumHs0xadaksW0cAS9", // 3
		"8dhEXXJLAEq0ZH4va5xC", // 95
		"Mtg00S5dHXBL7ASHEfNd", // 8
		"330aI6KcSbSCoioRAde1", // 4
	} {
		pass, err := rs.Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, expect, pass)
	}
}
