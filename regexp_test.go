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
	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+[[:punct:]]`

	gen, err := ParseRegexp(pattern, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"acd5VaaaaaaaaaaaaaaabbbbbbbbbbbccccdddeeeeeeeeeeeeeeeeyTN3~YP<VZ=2: ",
		"acd6daaaaaaaaaaaaaabccccdddddeeeeeeeeggggggggghhhz2S+@",
		"x59577893421Azzyyyyyyzzyyyyyyzzzzyyzzzzyyzz<",
		"acd5eaaaaaaaaaaaaaabbbbbbbbbbbbbccccdddeeeeeeeeeeeeeefK>>U<JLD",
		"x925947WQOzzzzyyyyzzyyyyzzyyyyyyzzyyyy_",
		"x2333219060048yyyyzzyyzzyyyyzzyyzzyyyy|",
		"x8138743=yyyyyyzzyyyyyyyyyyzz,",
		"x3eKxmpyyzzzzzzyyzzzzzzyyyyzz@",
		"x84131748(yy*",
		"x3457867#zzzzzzzzyyzzyyzzzzyyzzzzyy.",
		"x088772022554J*9$0yyzzyyyyzzzzzzyyzzzzzzyyzzyyyy%",
		"x4940776138Hq/f}zzyyyyyyyyyyzzzzzzyy^",
		"x2965.yyzzzzyyzzzzyyyyzzzzzzzzyyyyyy,",
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
	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+[[:punct:]]`

	var p RegexpParser
	p.SetAnyRangeTable(asciiGreek13RangeTable)

	gen, err := p.Parse(pattern, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"acd5Vaaaaaaaaaaaaaaabbbbbbbbbbbccccdddeeeeeeeeeeeeeeeeᾋ<ὈϠἕΎ῭𐆆ῐὟὋϹθἣ",
		"acd0Daaaaaaaaabbccccdddddeeeeeeeeeeeeeeeeeeegggggghhhhhhhhhhh𐅆ϐ𐅈𐆋+𝈛",
		"x85ῨΫyΆzz.",
		"acd7xaaaaaaaaabbbbbbbbbbbbbbccccddddddeeeeeeeeeeeeeeeeeee𐅅ύ𝈇ὤέᾕ𝈇Ρ𝈢ξ𝈈Ὄ΄",
		"x5712674230081408Vὦ𐅆𝈈𐆊yyyyyyyyzzyyyy|",
		"acd4haaabbbbbbbbbbbccccddddddeeeeeeeeegggggggggggggggghhhhhhͲἀ΅{ᾐ:𝈰𐅄𐆆𐅺῞ᾔ𐆊ΖϚΉ",
		"x275980syyyyzzzzzzzzyyyy>",
		"abd2caaaaaaaaaaaaabbbbbbbccccddddddeeeeeeeeefggggggggggggggghhhhhhhhhω*𐅉𝈽𐅗ῖΨ𐅊",
		"x10887720225545΅𐆄ἁ.Όzzzzyyzzzzzzyyzzyyyyyyzz^",
		"acd4babbbbbbbbbbccccdddeeeeeeeeeeeeeeeef𐆀Ψ𐅸ᾉἥὶ𝈂ἒᾤΫ𝈙𐅮𐅵",
		"abd9Gaaaabbccccddddddeeeeeeeeeeegggggghhhhhhhhᾧᾜ+ᵞᾚῒSᾞᾲ𐆍ϝ𐆇Ὗ𐅈",
		"abd5maaaaaaaaaaaaabbbbbccccddddddeeeeeeeeeeeeeeefggggggghhhhᾫὰᾃϠ𐅾ᵧὤ𐅒ᾨὃ",
		"abd8kaaaaaaaabbbbbbbbbccccdddeeeeeeeeeeeeeeeeeeeefἋϻὗͻ𐆆ᾶϓ𝈛Ὲ𝈿𝈢𝈶𝈽Ͳϸ῭ῑ",
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
	LatinOffset: 5,
}

func TestRegexpAnyNL(t *testing.T) {
	pattern := `[a-z][0-9]-[\x00-\x7f]{10}-[A-Z][0-9]-.{10}-(?-s:.{10})`

	var p RegexpParser
	p.SetAnyRangeTable(rangeTableReducedNL)

	gen, err := p.Parse(pattern, syntax.Perl|syntax.DotNL)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"a2-1AA1bA-2ab-A1-a-a-ba\n1a\n-bBbAbB12-1",
		"a0-0-0\n2AAAa1-B2-b\n-1bb2022-0--100a2A1",
		"a2-0--1\nb1A\na-B1-AbB\nbb1B-b-1aBB-aB0b0",
		"a1-B10-A2A\nB\n-A2-\n00b00b001-a1BA-10A-2",
		"b0-BBBAAa1aA\n-B2-Aa\n-Bba22A-BbB-B21B-1",
		"b1-10aAa\nab1a-B2-2B1bBaa-bb-a20Ba2bBa1",
		"b0-b2B-0Bb--b-B2-B0--0\n0A\n2--b22ABbbBA",
		"b2--Ab2b2\nB\n--A1-2\n1-B02B-b-00A-A-aABa",
		"a0-BA\nb-0-2b2-A2-a0\n-1B-002-1b--a1BB00",
		"a1-2Ab0BB2BAb-A2-1a2---Ab10-ba10Bbb1A1",
		"b0--ba1--b-10-B1---BAAaBB22--Ab2BA2120",
		"a0-\nbB2022A02-B2--0BA201\nA1-1a0--Ab01A",
		"a2-2a--Ba0a1b-A1-A2aB11\nb1a-1A-01ABB20",
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
		"AbXYaBCzA6A9B2b6TEST0no69",
		"ABwyAbczA1a0b6B6TEsT0no69",
		"aBmyaBCZa9A1b5b5test0no69",
		"AblYABczA5A8B7B2TesT0no69",
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
		"aBXYabcza8A9B6B9TEsTAnoB",
		"aByYabczA8a4b9b4teStAnoB",
		"aBsYabcza6a0B5b8TEstAnob",
		"abXyabcza7a7b8b3TEStanoB",
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

func TestRegexpPotentialOptimisations(t *testing.T) {
	// We could opt to map (Z+)? and (Z*)? to Z*, but we elect not to because
	// while they have the same meaning, they don't have the same probability.
	// ? has a questNoChanceNumerator/questNoChanceDenominator == 1/2 chance to
	// output nothing while * has a 1/maxUnboundedRepeatCount == 1/15 change to
	// output nothing. Also (Z+)? and Z* have a different maximum repeat count.
	//
	// We could optimise (|Z) to Z?, but we elect not to because that allows
	// callers to guarantee they'll get a 50-50 chance even if the
	// questNoChanceNumerator/questNoChanceDenominator == 1/2 value changes.
	pattern := `([a-z]+)?-([a-z]*)?-(|z)`

	gen, err := ParseRegexp(pattern, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"-eishgyluaru--------xdjixrm-kysahqom-z-qto-xljhsjlqg-",
		"al----qarpqt-z--m-z-wggtngiovdmb-jknfncptczbuqov-z-hqcfqgsxekdm--z",
		"oz-useuce-z---z----zsnhmlvkbat--z---",
	} {
		pass, err := Repeat(gen, "-", 5).Password(tr)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)

		matchPattern := "^(?:" + pattern + "-" + pattern + "-" + pattern + "-" + pattern + "-" + pattern + ")$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", matchPattern, pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}
}

type incUint8 uint8

func (m *incUint8) Read(p []byte) (int, error) {
	p[0] = byte(*m)
	(*m)++
	return 1, nil
}

func TestRegexpQuestProbability(t *testing.T) {
	gen, err := ParseRegexp(`Z?`, syntax.Perl)
	require.NoError(t, err)

	var (
		ir    incUint8
		empty int
	)
	for i := 0; i < questNoChanceDenominator; i++ {
		pass, err := gen.Password(&ir)
		require.NoError(t, err)
		empty += 1 - len(pass)
	}

	assert.Equal(t, questNoChanceNumerator, empty, "wrong number of empty passwords")
}

func TestRegexpSpecialCaptures(t *testing.T) {
	var p RegexpParser
	p.SetSpecialCapture("word", SpecialCaptureBasic(EFFLargeWordlist))
	p.SetSpecialCapture("words", SpecialCaptureWithRepeat(EFFLargeWordlist, " "))

	gen, err := p.Parse(`((?P<word>) ){6}[[:upper:]][[:digit:]][[:punct:]]`, syntax.Perl)
	require.NoError(t, err)

	tr := newTestRand()

	for _, expect := range []string{
		"reprint wool pantry unworried mummify veneering U2,",
		"uncolored phrase spearmint vividness haunt esquire M3)",
		"stargazer acetone stroller frantic catapult tipping Q7@",
		"rake linseed consumer phantom handclasp blast R3/",
	} {
		pass, err := gen.Password(tr)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)
		allRunesAllowed(t, rangeTableASCII, pass)
	}

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
		{"(?P<unknown>)", "passit: named capture refers to unknown special capture factory"},
		{"(?P<unknown>inner)", "passit: named capture refers to unknown special capture factory"},
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
	const pattern = `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+[[:punct:]]`
	var p RegexpParser

	for n := 0; n < b.N; n++ {
		_, err := p.Parse(pattern, syntax.Perl)
		if err != nil {
			require.NoError(b, err)
		}
	}
}

func BenchmarkRegexpPassword(b *testing.B) {
	const pattern = `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+[[:punct:]]`
	gen, err := ParseRegexp(pattern, syntax.Perl)
	if err != nil {
		require.NoError(b, err)
	}

	benchmarkGeneratorPassword(b, gen)
}
