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

var rangeTableReducedNL = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: '\n', Hi: '\n', Stride: 1},
		{Lo: '-', Hi: '-', Stride: 1},
		{Lo: '0', Hi: '2', Stride: 1},
		{Lo: 'A', Hi: 'B', Stride: 1},
		{Lo: 'a', Hi: 'b', Stride: 1},
	},
	LatinOffset: 3,
}

func TestRegexpAnyNL(t *testing.T) {
	pattern := `[a-z][0-9]-[\x00-\x7f]{10}-[A-Z][0-9]-.{10}-(?-s:.{10})`

	var p RegexpParser
	p.SetAnyRangeTable(rangeTableReducedNL)

	gen, err := p.Parse(pattern, syntax.Perl|syntax.DotNL)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"a2-b-b1A00Aa2-A2-bBaaA\naaBA-b100b0A0-0",
		"a0-2B2AA-\n-B2-B0-\n\n2\nA-\nB1A-a-b2101-1A",
		"a0-211a-\nA2B2-A2-BB-b\n-00\na-A10BBB--a0",
		"a1-\naB-b1B1AA-A1-aB2AA2B1BB-B2-2-aB-0A",
		"a1-\nabA\nbBbB2-A0-\n0B1B1aBa1-AB1aba2Ab1",
		"b1-1-B-A102Ab-B0-bBbaA22\nA1-bAA0A1B-Aa",
		"a2-AA\n0aB--BA-B0-001-a\n1a21-11Aa1-B1ba",
		"b2-B0b-AB0\n1\n-B2-b\na-2a2A1b-B-aA02-b2a",
		"b2-2bBA10a2\nA-B1-AB2A1-BA0B--0AaBba-12",
		"a2-A1--0b1\n22-B1-a0a-0122aa-AA-a-aBb11",
		"a0-B20-2a\n-2b-A0-b0b101AABB-B010-212b0",
		"a1-0BA--2-1A--B0---B20-2Aab--202-0bAb2",
		"b0-A0\nB2--0b2-B2-B-a2A1\nA00-A-1a10a0a1",
	} {
		pass, err := gen.Password(tr)
		require.NoError(t, err)

		if !assert.Equal(t, expect, pass) {
			t.Logf("%+q", pass)
		}

		matchPattern := "^(?s:" + pattern + ")$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%+q)", matchPattern, pass)
		allRunesAllowed(t, rangeTableReducedNL, pass)
	}
}

func TestRegexpFoldCaseFlag(t *testing.T) {
	pattern := `ab[a-z]y(abc)z(a[0-9]){2}(b[0-9]){2}test0(?-i:no)69`

	gen, err := ParseRegexp(pattern, syntax.Perl|syntax.FoldCase)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"AbzYABCzA8A0B1B9tESt0no69",
		"aBZYaBcza7A7B5B6TESt0no69",
		"aBVyaBcZA3A5b0b2teST0no69",
		"abayABCZA4a5b6B2TEst0no69",
	} {
		pass, err := gen.Password(tr)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)

		matchPattern := "^(?i:" + pattern + ")$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", matchPattern, pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

func TestRegexpFoldCaseCapture(t *testing.T) {
	pattern := `a(?i:b[a-z]y)(abc)z(?i:(a[0-9]){2}(b[0-9]){2})(?i:test)(?i:a(?-i:no)b)`

	gen, err := ParseRegexp(pattern, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"aBHyabczA2A4b4B6TEsTanob",
		"aBsyabcza4a4b6b9teStAnob",
		"aBEYabczA0a3B9b9TeSTanoB",
		"abXYabcza2a9B4b3TeSTAnoB",
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

func TestRegexpSpecialCaptures(t *testing.T) {
	var p RegexpParser
	p.SetSpecialCapture("word", SpecialCaptureBasic(EFFLargeWordlist))

	gen, err := p.Parse(`((?P<word>) ){6}[[:upper:]][[:digit:]][[:punct:]]`, syntax.Perl)
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

		gen, err := p.Parse(tc.pattern, syntax.Perl)
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
		_, err := p.Parse(tc.pattern, syntax.Perl)
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
