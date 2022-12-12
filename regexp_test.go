package passit

import (
	"math/rand"
	"regexp"
	"regexp/syntax"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexp(t *testing.T) {
	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`

	tmpl, err := ParseRegexp(pattern, syntax.Perl)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"x24130549434343dj;u6zzzz*",
		"abd1oaaaaaaaabbbbbbbccccdddddeeeeebJG",
		"abd5WaaaaaaaaaaabbbbbbccccdddddeeeeeeebB8A@L:=\"",
		"x795e&yyyyyyzzyyyyzzyyzzzzyyzzyyzzyy+",
		"x19yyyyyyzzyy.",
		"x33345yyzzyyyyyyzzyyyyzzzzzz`",
		"acd2XaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeefC NJ% }5/\"'0SNK\"4",
		"abd7Vaaaabbbbbbbbbbbbbccccddddeeeeeeeeb4Q(&4I{2",
		"x99{Ayyzzzzyyyyyyyyzzyyyyyyyyyyzz(",
		"abd3naaaaaaaaaaabbbbbbccccdddddeeeeeeeeeeefggggggghhhhhhhhhhhhhh ?W",
		"abd8PaaaaabbbbbbbbccccdddeeeeeeeeeeegBRIKII~IFE",
		"x0921UmGujzz=",
		"x4941983495469641gv; -zzzz;",
	} {
		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)

		matchPattern := "^(?:" + pattern + ")$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", matchPattern, pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

// asciiGreek13RangeTable is rangeTableASCII + unicode.Greek @ 13.0.0.
var asciiGreek13RangeTable = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x0020, Hi: 0x007e, Stride: 1},
		{Lo: 0x0370, Hi: 0x0373, Stride: 1},
		{Lo: 0x0375, Hi: 0x0377, Stride: 1},
		{Lo: 0x037a, Hi: 0x037d, Stride: 1},
		{Lo: 0x037f, Hi: 0x0384, Stride: 5},
		{Lo: 0x0386, Hi: 0x0388, Stride: 2},
		{Lo: 0x0389, Hi: 0x038a, Stride: 1},
		{Lo: 0x038c, Hi: 0x038e, Stride: 2},
		{Lo: 0x038f, Hi: 0x03a1, Stride: 1},
		{Lo: 0x03a3, Hi: 0x03e1, Stride: 1},
		{Lo: 0x03f0, Hi: 0x03ff, Stride: 1},
		{Lo: 0x1d26, Hi: 0x1d2a, Stride: 1},
		{Lo: 0x1d5d, Hi: 0x1d61, Stride: 1},
		{Lo: 0x1d66, Hi: 0x1d6a, Stride: 1},
		{Lo: 0x1dbf, Hi: 0x1f00, Stride: 321},
		{Lo: 0x1f01, Hi: 0x1f15, Stride: 1},
		{Lo: 0x1f18, Hi: 0x1f1d, Stride: 1},
		{Lo: 0x1f20, Hi: 0x1f45, Stride: 1},
		{Lo: 0x1f48, Hi: 0x1f4d, Stride: 1},
		{Lo: 0x1f50, Hi: 0x1f57, Stride: 1},
		{Lo: 0x1f59, Hi: 0x1f5f, Stride: 2},
		{Lo: 0x1f60, Hi: 0x1f7d, Stride: 1},
		{Lo: 0x1f80, Hi: 0x1fb4, Stride: 1},
		{Lo: 0x1fb6, Hi: 0x1fc4, Stride: 1},
		{Lo: 0x1fc6, Hi: 0x1fd3, Stride: 1},
		{Lo: 0x1fd6, Hi: 0x1fdb, Stride: 1},
		{Lo: 0x1fdd, Hi: 0x1fef, Stride: 1},
		{Lo: 0x1ff2, Hi: 0x1ff4, Stride: 1},
		{Lo: 0x1ff6, Hi: 0x1ffe, Stride: 1},
		{Lo: 0x2126, Hi: 0xab65, Stride: 35391},
	},
	R32: []unicode.Range32{
		{Lo: 0x10140, Hi: 0x1018e, Stride: 1},
		{Lo: 0x101a0, Hi: 0x1d200, Stride: 53344},
		{Lo: 0x1d201, Hi: 0x1d245, Stride: 1},
	},
	LatinOffset: 1,
}

func TestRegexpUnicodeAny(t *testing.T) {
	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`

	var p RegexpParser
	p.SetAnyRangeTable(asciiGreek13RangeTable)

	tmpl, err := p.Parse(pattern, syntax.Perl)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"x24130549434343`𐆀𐅞ῧϴzzzz*",
		"abd1oaaaaaaaabbbbbbbccccdddddeeeee4@ό",
		"abd5Waaaaaaaaaaabbbbbbccccdddddeeeeeeeϋϰῠ𐆅ψᵦὐΉὦ",
		"x795ὛΗyyyyyyzzyyyyzzyyzzzzyyzzyyzzyy+",
		"x19yyyyyyzzyy.",
		"x33345yyzzyyyyyyzzyyyyzzzzzz`",
		"acd2Xaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeefβᵡϊ𐅴𝈥ᵡῄ𐅙𐅺ᵦ𐅘ΏὊ=8Wᾪ",
		"acd4naaaaaaaaaaaabbbbbbbbbbccccddddddeeeeeeeeeeegghhhhhhh$𐅇ἳ`",
		"abd3SaaaaaaabbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeeeeggggggggggggghhhhhhhhhhhhhhΧΏ͵UP𐅎ῧ",
		"abd0SaaaaaaabccccdddddeeeeeeeegggggghhhhhhhᾤὖίYᾄ𐅾𝈉ῚὌJͱᾺ",
		"x42886446y*ἅ\"ᾲyy$",
		"x92178ͳςΉῗyyzzzzyyzzyyzzzzyyzz-",
		"x46964136170313͵ὣCzzzz:",
	} {
		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		if !assert.Equal(t, expect, pass) {
			t.Logf("%+q", pass)
		}

		matchPattern := "^(?:" + pattern + ")$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%+q)", matchPattern, pass)
		allRunesAllowed(t, asciiGreek13RangeTable, pass)
	}
}

func TestRegexpSpecialCaptures(t *testing.T) {
	var p RegexpParser
	p.SetSpecialCapture("word", SpecialCaptureBasic(EFFLargeWordlist))

	tmpl, err := p.Parse(`((?P<word>) ){6}[[:upper:]][[:digit:]][[:punct:]]`, syntax.PerlX)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"timothy hubcap partner frigidly usage probiotic E5/",
		"configure drool tainted heading mama synthesis Z3)",
		"gusty judicial expansive groin widely vocalist F4-",
		"refutable velocity synergy phoenix wand tipper C2?",
	} {
		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

func TestRegexpSpecialCaptureFactories(t *testing.T) {
	var p RegexpParser
	p.SetSpecialCapture("word", SpecialCaptureBasic(EFFLargeWordlist))
	p.SetSpecialCapture("words", SpecialCaptureWithRepeat(EFFLargeWordlist, " "))

	for _, tc := range []struct {
		pattern, expect string
	}{
		{"(?P<word>)", "duplicity"},
		{"(?P<words>)", "duplicity"},
		{"(?P<words>1)", "duplicity"},
		{"(?P<words>2)", "duplicity employee"},
		{"(?P<words>03)", "duplicity employee praising"},
	} {
		testRand := rand.New(rand.NewSource(1))

		tmpl, err := p.Parse(tc.pattern, syntax.PerlX)
		if !assert.NoError(t, err, tc.pattern) {
			continue
		}

		pass, err := tmpl.Password(testRand)
		if assert.NoError(t, err, tc.pattern) {
			assert.Equal(t, tc.expect, pass, tc.pattern)
			allRunesAllowed(t, rangeTableASCII, pass)
		}
	}

	for _, tc := range []struct {
		pattern, errString string
	}{
		{"(?P<word>1)", "passit: unsupported capture"},
		{"(?P<word> )", "passit: unsupported capture"},
		{"(?P<word>[0-9])", "passit: unsupported capture"},
		{"(?P<words>[0-9])", "passit: unsupported capture"},
		{`(?P<words>\+2)`, "passit: failed to parse capture: strconv.ParseUint: parsing \"+2\": invalid syntax"},
		{"(?P<words>-3)", "passit: failed to parse capture: strconv.ParseUint: parsing \"-3\": invalid syntax"},
		{"(?P<words>0x12)", "passit: failed to parse capture: strconv.ParseUint: parsing \"0x12\": invalid syntax"},
		{"(?P<words>-0x12)", "passit: failed to parse capture: strconv.ParseUint: parsing \"-0x12\": invalid syntax"},
		{"(?P<words>4a)", "passit: failed to parse capture: strconv.ParseUint: parsing \"4a\": invalid syntax"},
	} {
		_, err := p.Parse(tc.pattern, syntax.PerlX)
		assert.EqualError(t, err, tc.errString, tc.pattern)
	}
}

func BenchmarkRegexpParse(b *testing.B) {
	const pattern = `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`
	var p RegexpParser

	for n := 0; n < b.N; n++ {
		_, err := p.Parse(pattern, syntax.Perl)
		if err != nil {
			require.NoError(b, err)
		}
	}
}
