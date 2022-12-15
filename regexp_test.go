package passit

import (
	"regexp"
	"regexp/syntax"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexp(t *testing.T) {
	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`

	gen, err := ParseRegexp(pattern, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"acd7faaaaaaaabbbbbbbbbbbccccdddddeeeeeeeeeeeeeeee0u.=_K?L#",
		"x141s8zzyyzzyyzzyyyyyy[",
		"acd3labbbbbbbbbbccccddddddeeeeeeeeeeeeeeeeefadM[EV Y4V,",
		"x8100jnzzzzzzyyyyyyyyyyzzzzyyyyyy|",
		"x06314366735700yyzz'",
		"x29323870\\nl]zzzzzzyyzzzzzzyy$",
		"abd5Baaaaaaaaabbbbbbbbbbbbbccccdddeeeeeeeeeeeeeeeeeeef]pRACY!>S5RG)>",
		"x22581669274S.FERzz#",
		"acd3Vaaaabbbbbbccccddddeeeeeeeeeeeeeeeeee6\\9)+|.T/ =/:",
		"abd2raaaaaaaaaaaabbccccdddeeeeeeeeeeeeeeeeeeu(&78)\"~#",
		"x3508122602218~S~MXyyyyyyyyzzzzzzzzyyyyzzyyyyyyzz,",
		"x0560855439%i Jyyyyyyyyyyzzzzzzzzyyyyzzzz-",
		"abd8Maabccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhh)MOC4P`K~C>/W}",
	} {
		pass, err := gen.Password(tr)
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

	gen, err := p.Parse(pattern, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"acd7faaaaaaaabbbbbbbbbbbccccdddddeeeeeeeeeeeeeeee1ᾒ𝉁-Ἆᴦδώ𝈩",
		"x141Ἦ𐅄zzyyzzyyzzyyyyyy[",
		"acd3labbbbbbbbbbccccddddddeeeeeeeeeeeeeeeeef𝈟ῼἡὒ𐅌ἐυὧ𝈏ᾨῧ",
		"x100ήᾣzzzzzzyyyyyyyyyyzzzzyyyyyy|",
		"x06314366735700yyzz'",
		"x29323870Ἤoὦδzzzzzzyyzzzzzzyy$",
		"abd5BaaaaaaaaabbbbbbbbbbbbbccccdddeeeeeeeeeeeeeeeeeeefἽΙῺϿ𝈑ᾐΏἣϲ𝈢Ὃ𐅯η𐅸",
		"x22581669274Ὼϓᾧᾜὼzz#",
		"acd3Vaaaabbbbbbccccddddeeeeeeeeeeeeeeeeeeᾉᵪᾋ𐅌ᵩϸᾃϠ𐅾ᵧὤ𐅒ᾨ",
		"abd2raaaaaaaaaaaabbccccdddeeeeeeeeeeeeeeeeeeO῀ἠἇᾤἙἴςε",
		"x3508122602218Α𐅏Ὴψ𐅮yyyyyyyyzzzzzzzzyyyyzzyyyyyyzz,",
		"x0560855439ύ+TὟyyyyyyyyyyzzzzzzzzyyyyzzzz-",
		"abd8Maabccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhh𝉄῭ᵡᵟΎῄB𝈰Ω῀ᾛ϶ϳΌ",
	} {
		pass, err := gen.Password(tr)
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

	gen, err := p.Parse(`((?P<word>) ){6}[[:upper:]][[:digit:]][[:punct:]]`, syntax.PerlX)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"reprint wool pantry unworried mummify veneering U9]",
		"steep cresting dastardly cubical thriving procreate V9_",
		"acetone stroller frantic catapult tipping wildland P6*",
		"consumer phantom handclasp blast broadside spleen E4[",
	} {
		pass, err := gen.Password(tr)
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
		{"(?P<word>)", "reprint"},
		{"(?P<words>)", "reprint"},
		{"(?P<words>1)", "reprint"},
		{"(?P<words>2)", "reprint wool"},
		{"(?P<words>03)", "reprint wool pantry"},
	} {
		tr := newTestRand()

		gen, err := p.Parse(tc.pattern, syntax.PerlX)
		if !assert.NoError(t, err, tc.pattern) {
			continue
		}

		pass, err := gen.Password(tr)
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

func BenchmarkRegexpPassword(b *testing.B) {
	const pattern = `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`
	gen, err := ParseRegexp(pattern, syntax.Perl)
	if err != nil {
		require.NoError(b, err)
	}
	tr := newTestRand()

	for n := 0; n < b.N; n++ {
		_, err := gen.Password(tr)
		if err != nil {
			require.NoError(b, err)
		}
	}
}
