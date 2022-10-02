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
		"x90822236719&17:yyyyzzzzyy=",
		"x7977150zzyyyyzzzzzzyyzzzzyyyy<",
		"x14404'5\"Lyyyyzzyyyyyyyyyyyyyyzz#",
		"acd0saaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeee?s4;RM>;,2EI/1B~S",
		"acd2haaaaaaabbbbbbbccccdddddeeeeeeeeeeeeeeeeegG($|A=}6Q('3)7]U",
		"acd3labbbbbbbbbbccccddddddeeeeeef\"..$~['`<_\\*:&MS2",
		"abd4iaaaaaaaabccccddddddeeeeeeeeeeeeeefY&TV",
		"x18372266066416nDyyyy?",
		"acd2saaaaaaaaaaaaaabbbbbbbbbbbbbbbccccddddddeeeeeeeeeeeeeeeeegggggghhhhhhy;[\\I.P1P;6_F_N)=6",
		"abd3Paaaaaabbbbbbbbbccccddddddeeeeeeeeeeeeeeefl]-3;V/8H",
		"acd8KaabbbbbbccccddddddeeeeeeZm7 _AJL~M\"]H}YB[",
		"abd8sbbbbbbbccccddddeeeeeeeeefTSX#Z>V+(I",
		"x49868zzzzzzyyyyyyzzzzyyzzyyzz{",
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
		"x90822236719\u1f8cX\u03c4\u1f63yyyyzzzzyy=",
		"x7977150zzyyyyzzzzzzyyzzzzyyyy<",
		"x14404ZbS\u1f94yyyyzzyyyyyyyyyyyyyyzz#",
		"acd0saaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeee\u1f75\u1f24\u03d7\U0001d23b\u1f78\u03ddPI\u03c9\u1fb2\u03a8\U00010173\U0001018d\u1fce\u03d1\u1f1a\u1fe8",
		"acd2haaaaaaabbbbbbbccccdddddeeeeeeeeeeeeeeeee\u0398L3\u1fa0\U00010153\u0370\u1f2f\U00010159\u1d67\u1fcd\u1d5f\U00010160\U0001017b\u03f5\U00010188\U0001d22c_",
		"acd3labbbbbbbbbbccccddddddeeeeeef\U0001016e\u1f31&\u1d61\u1fa1I\U000101765\u1fd7\u1f18\u03d2\u1f59\U0001d21b\U00010145\u1f6e\u1faf\U00010161",
		"abd4iaaaaaaaabccccddddddeeeeeeeeeeeeeefm\u1fd7\u1f82\u1f2a",
		"x18372266066416\U0001d21b\U0001018eyyyy?",
		"acd2saaaaaaaaaaaaaabbbbbbbbbbbbbbbccccddddddeeeeeeeeeeeeeeeeegggggghhhhhh\u1f86\U00010161\U0001d203\U0001d205*V'\u1f66\u03ce\u1d27\u1f7c\u1f22\u03cbJ\u1fee\u1fd3SR",
		"abd3Paaaaaabbbbbbbbbccccddddddeeeeeeeeeeeeeeef\u1f57\u1f4a)\u1fac\u03fe\u1f41\u1f26\u1fbdQ",
		"acd8Kaabbbbbbccccddddddeeeeee\U0001d23e\u1f20\U0001d206\u1fcb\u1f9d\u1fab\u1f0c\u1ffd\u03d1\u03c5\U0001d209\u0386\U00010141\U0001017e\u03d0\u1fb8\u1f4b",
		"abd8sbbbbbbbccccddddeeeeeeeeef@\u1f1c\U00010175\U00010166\u03c5\U0001d23e\U00010157GX\u03b2",
		"x49868zzzzzzyyyyyyzzzzyyzzyyzz{",
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
	p.SetSpecialCapture("word", SpecialCaptureBasic(EFFLargeWordlist(1)))

	tmpl, err := p.Parse(`((?P<word>) ){6}[[:upper:]][[:digit:]][[:punct:]]`, syntax.PerlX)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"native remover dismay vocation sepia backtalk E2`",
		"hemlock exit finance obscure dusk rigor A8}",
		"gone spouse hungrily zoning say shrug Q5[",
		"crispness bannister pauper silica stiffen deduct S6+",
	} {
		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

func TestRegexpSpecialCaptureFactories(t *testing.T) {
	var p RegexpParser
	p.SetSpecialCapture("word", SpecialCaptureBasic(EFFLargeWordlist(1)))
	p.SetSpecialCapture("words", SpecialCaptureWithRepeat(EFFLargeWordlist(1), " "))

	for _, tc := range []struct {
		pattern, expect string
	}{
		{"(?P<word>)", "clanking"},
		{"(?P<words>)", "clanking"},
		{"(?P<words>1)", "clanking"},
		{"(?P<words>2)", "clanking avalanche"},
		{"(?P<words>03)", "clanking avalanche cursor"},
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
