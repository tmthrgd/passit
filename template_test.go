package passit

import (
	"math"
	"math/rand"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCharset(t *testing.T, template string) func(int) Template {
	t.Helper()

	tmpl, err := FromCharset(template)
	require.NoError(t, err)
	return tmpl
}

func TestJoin(t *testing.T) {
	{
		pattern := regexp.MustCompile(`^([a-z]+ ){5}[A-Z][0-9][~!@#$%^&*()] \+abc-[de]$`)

		tmpl := Join("",
			EFFLargeWordlist(5),
			Space,
			LatinUpper(1),
			Number(1),
			mustCharset(t, "~!@#$%^&*()")(1),
			Space,
			FixedString("+abc"),
			Hyphen,
			mustCharset(t, "de")(1),
		)

		testRand := rand.New(rand.NewSource(0))

		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, "native remover dismay vocation sepia C2@ +abc-e", pass)
		assert.Truef(t, pattern.MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", pattern, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}

	{
		tmpl := Join("@$", LatinUpper(1), LatinLower(1), LatinMixed(1))

		testRand := rand.New(rand.NewSource(0))

		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, "L@$w@$r", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

func TestRepeat(t *testing.T) {
	assert.PanicsWithValue(t, "passit: count must be positive", func() {
		Repeat(Hyphen, " ", -1)
	})

	assert.Equal(t, FixedString(""), Repeat(Hyphen, " ", 0),
		"Repeat with count zero should return empty FixedString")

	assert.Equal(t, Hyphen, Repeat(Hyphen, " ", 1),
		"Repeat with count one should return Template")

	for _, tc := range []struct {
		count  int
		sep    string
		expect string
	}{
		{2, " ", "native remover"},
		{4, "", "nativeremoverdismayvocation"},
		{15, "-", "native-remover-dismay-vocation-sepia-backtalk-think-conjure-autograph-hemlock-exit-finance-obscure-dusk-rigor"},
	} {
		testRand := rand.New(rand.NewSource(0))

		pass, err := Repeat(EFFLargeWordlist(1), tc.sep, tc.count).Password(testRand)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}

func TestRandomCount(t *testing.T) {
	_, err := RandomCount(EFFLargeWordlist, 10, 7)
	assert.EqualError(t, err, "passit: min argument cannot be greater than max argument",
		"min greater than max")

	for _, tc := range [][2]int{
		{0, maxInt32},
		{-1, maxInt32 - 1},
		{-maxInt32, 0},
		{math.MinInt, 0},
		{0, math.MaxInt},
		{math.MinInt, math.MaxInt},
		{math.MinInt + 1, math.MaxInt},
		{math.MinInt, math.MaxInt - 1},
		{math.MinInt + 1, math.MaxInt - 1},
	} {
		_, err = RandomCount(EFFLargeWordlist, tc[0], tc[1])
		assert.EqualErrorf(t, err, "passit: [min,max] range too large",
			"out of range: %v", tc)
	}

	for _, tc := range [][2]int{
		{0, 0},
		{1, 1},
		{70, 70},
		{-70, -70},
		{maxInt32, maxInt32},
		{-maxInt32, -maxInt32},
		{math.MaxInt, math.MaxInt},
		{math.MinInt, math.MinInt},
	} {
		tmpl, err := RandomCount(EFFLargeWordlist, tc[0], tc[1])
		if !assert.NoErrorf(t, err, "equal min and max should not error: %v", tc) {
			continue
		}
		assert.IsTypef(t, (*embededList)(nil), tmpl,
			"equal min and max should return template: %v", tc)
	}

	for _, tc := range []struct {
		min, max int
		expect   string
	}{
		{1, 2, "remover dismay"},
		{2, 5, "remover dismay vocation"},
		{4, 7, "remover dismay vocation sepia backtalk"},
		{10, 20, "remover dismay vocation sepia backtalk think conjure autograph hemlock exit finance obscure dusk rigor hemlock dusk blouse"},
	} {
		tmpl, err := RandomCount(EFFLargeWordlist, tc.min, tc.max)
		if !assert.NoErrorf(t, err, "valid range should not error: %v", tc) {
			continue
		}

		testRand := rand.New(rand.NewSource(0))

		pass, err := tmpl.Password(testRand)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}
