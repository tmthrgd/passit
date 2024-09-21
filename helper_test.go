package passit

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
	"math/bits"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"
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
	clear(p)
	return len(p), nil
}

func errTestReader() io.Reader {
	return iotest.ErrReader(errors.New("should not call Read"))
}

func TestJoin(t *testing.T) {
	{
		pattern := regexp.MustCompile(`^([a-z]+ ){5}[A-Z][0-9][~!@#$%^&*()] \+abc-[de]$`)

		gen := Join("",
			Repeat(EFFLargeWordlist, " ", 5),
			Space,
			LatinUpper,
			Digit,
			FromCharset("~!@#$%^&*()"),
			Space,
			String("+abc"),
			Hyphen,
			FromCharset("de"),
		)

		tr := newTestRand()

		pass, err := gen.Password(tr)
		require.NoError(t, err)

		assert.Equal(t, "reprint wool pantry unworried mummify L2* +abc-e", pass)
		assert.Truef(t, pattern.MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", pattern, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}

	{
		gen := Join("@$", LatinUpper, LatinLower, LatinMixed)

		tr := newTestRand()

		pass, err := gen.Password(tr)
		require.NoError(t, err)

		assert.Equal(t, "Y@$z@$x", pass)
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
		"Repeat with count one should return Generator")

	for _, tc := range []struct {
		count  int
		sep    string
		expect string
	}{
		{2, " ", "reprint wool"},
		{4, "", "reprintwoolpantryunworried"},
		{12, "-", "reprint-wool-pantry-unworried-mummify-veneering-securely-munchkin-juiciness-steep-cresting-dastardly"},
	} {
		tr := newTestRand()

		pass, err := Repeat(EFFLargeWordlist, tc.sep, tc.count).Password(tr)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}

func TestRandomRepeat(t *testing.T) {
	const maxInt = 1<<(bits.UintSize-1) - 1

	assert.PanicsWithValue(t, "passit: min argument cannot be greater than max argument", func() {
		RandomRepeat(Hyphen, " ", 10, 7)
	}, "min greater than max")

	assert.PanicsWithValue(t, "passit: min argument must be positive", func() {
		RandomRepeat(Hyphen, " ", -5, 7)
	}, "min negative")

	assert.PanicsWithValue(t, "passit: min argument cannot be greater than max argument", func() {
		RandomRepeat(Hyphen, " ", 5, -7)
	}, "min greater than max; max negative")

	assert.PanicsWithValue(t, "passit: [min,max] range too large", func() {
		RandomRepeat(Hyphen, " ", 0, maxInt)
	}, "out of range: 0, max int")

	gen := RandomRepeat(Hyphen, " ", 0, 0)
	assert.Equal(t, Empty, gen, "min and max equal zero should return Empty")

	gen = RandomRepeat(Hyphen, " ", 1, 1)
	assert.Equal(t, Hyphen, gen, "min and max equal one should return Generator")

	for _, tc := range []int{
		70,
		1<<16 - 1,
		1<<31 - 1,
		maxInt,
	} {
		gen := RandomRepeat(Hyphen, " ", tc, tc)
		assert.IsTypef(t, (*repeatGenerator)(nil), gen,
			"equal min and max should return Repeat(...): %v", tc)
	}

	for _, tc := range []struct {
		min, max int
		sep      string
		expect   string
	}{
		{1, 2, " ", "mascot"},
		{2, 5, "", "mascotultimatumlanternlushly"},
		{4, 7, "-", "mascot-ultimatum-lantern-lushly-recoil-humvee"},
		{10, 20, " ", "mascot ultimatum lantern lushly recoil humvee uncolored phrase spearmint vividness haunt esquire cargo"},
	} {
		tr := newTestRand()

		pass, err := RandomRepeat(EFFLargeWordlist, tc.sep, tc.min, tc.max).Password(tr)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}

func TestAlternate(t *testing.T) {
	assert.Equal(t, Empty, Alternate(),
		"Alternate with no Generators should return Empty")

	assert.Equal(t, Hyphen, Alternate(Hyphen),
		"Alternate with single Generator should return Generator")

	for _, tc := range []struct {
		gens   []Generator
		expect string
	}{
		{[]Generator{}, ""},
		{[]Generator{LatinLower}, "y"},
		{[]Generator{LatinLower, LatinUpper, Digit}, "z"},
		{[]Generator{EFFShortWordlist1, EFFShortWordlist2, EFFLargeWordlist}, "yummy"},
		{[]Generator{
			Repeat(LatinLower, "!", 5),
			Repeat(LatinUpper, "@", 3),
			Repeat(Digit, "#", 7),
		}, "z!x!e!i!s"},
	} {
		tr := newTestRand()

		pass, err := Alternate(tc.gens...).Password(tr)
		if !assert.NoErrorf(t, err, "should not error when generating: %#v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "expected password: %#v", tc)
	}
}

func TestRejectionSample(t *testing.T) {
	rs := RejectionSample(Repeat(LatinMixedDigit, "", 20), func(s string) bool {
		return strings.Contains(s, "A") && strings.Contains(s, "0")
	})
	tr := newTestRand()

	for _, expect := range []string{
		"OVnA1oS7moBq0RUAOucW", // 1
		"N0A7kS2RKGEgim1C0pr7", // 23
		"RNrej47HNFYrK1z23A0D", // 25
		"Y2Za255TBLAQ6Dwb0bXy", // 12
		"57A0CPUu01ig0NI4BqvP", // 20
		"ThD321Av11GE0kvEQhLc", // 11
		"74gna0HQ5jjKALLPxdrF", // 15
		"awrGYutFKLcE3O6W0AE6", // 7
		"mqjrX7JnahA7gSuks0qx", // 1
		"TVz0ods6Ai0otGpdbbIn", // 15
	} {
		pass, err := rs.Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, expect, pass)
	}
}

func TestFromSlice(t *testing.T) {
	for _, tc := range []struct {
		expect string
		gen    Generator
	}{
		{"       ", FromSlice()},
		{"to to to to to to to to", FromSlice("to")},
		{"and or or and or and and or", FromSlice("and", "or")},
		{"ευτυχία υγεία αιώνια ελπίδα αιώνια ευτυχία ελπίδα αιώνια", FromSlice("ελπίδα", "υγεία", "ευτυχία", "αιώνια")},
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

func benchmarkGeneratorPassword(b *testing.B, gen Generator) {
	b.Helper()
	tr := newTestRand()

	for range b.N {
		_, err := gen.Password(tr)
		if err != nil {
			require.NoError(b, err)
		}
	}
}
