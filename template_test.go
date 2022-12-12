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

		testRand := rand.New(rand.NewSource(0))

		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, "timothy hubcap partner frigidly usage B0~ +abc-d", pass)
		assert.Truef(t, pattern.MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", pattern, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}

	{
		tmpl := Join("@$", LatinUpper, LatinLower, LatinMixed)

		testRand := rand.New(rand.NewSource(0))

		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, "H@$x@$k", pass)
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
		{2, " ", "timothy hubcap"},
		{4, "", "timothyhubcappartnerfrigidly"},
		{15, "-", "timothy-hubcap-partner-frigidly-usage-probiotic-yodel-playback-preaching-configure-drool-tainted-heading-mama-synthesis"},
	} {
		testRand := rand.New(rand.NewSource(0))

		pass, err := Repeat(EFFLargeWordlist, tc.sep, tc.count).Password(testRand)
		if !assert.NoErrorf(t, err, "valid range should not error when generating: %v", tc) {
			continue
		}

		assert.Equal(t, tc.expect, pass, "valid range expected password: %v", tc)
	}
}

func TestRandomRepeat(t *testing.T) {
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
		{0, math.MaxInt},
	} {
		_, err = RandomRepeat(Hyphen, " ", tc[0], tc[1])
		assert.EqualErrorf(t, err, "passit: [min,max] range too large",
			"out of range: %v", tc)
	}

	tmpl, err := RandomRepeat(Hyphen, " ", 0, 0)
	if assert.NoError(t, err, "min and max equal zero should not error") {
		assert.Equal(t, fixedString(""), tmpl,
			"min and max equal zero should return empty FixedString")
	}

	tmpl, err = RandomRepeat(Hyphen, " ", 1, 1)
	if assert.NoError(t, err, "min and max equal one should not error") {
		assert.Equal(t, Hyphen, tmpl,
			"min and max equal one should return template")
	}

	for _, tc := range []int{
		70,
		maxReadIntN,
		math.MaxInt,
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
		{1, 2, " ", "hubcap partner"},
		{2, 5, "", "hubcappartnerfrigidly"},
		{4, 7, "-", "hubcap-partner-frigidly-usage-probiotic"},
		{10, 20, " ", "hubcap partner frigidly usage probiotic yodel playback preaching configure drool tainted heading mama synthesis storage"},
	} {
		tmpl, err := RandomRepeat(EFFLargeWordlist, tc.sep, tc.min, tc.max)
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
