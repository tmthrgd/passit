package passit

import (
	"math/rand"
	"regexp"
	"regexp/syntax"
	"testing"

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
		allRunesAllowed(t, pass)
	}
}

func TestRegexpUnicodeAny(t *testing.T) {
	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`

	var p RegexpParser
	p.SetUnicodeAny()

	tmpl, err := p.Parse(pattern, syntax.Perl)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"x90822236719\U00017e98\uaade\U000245df\ufe14yyyyzzzzyy=",
		"x7977150zzyyyyzzzzzzyyzzzzyyyy<",
		"x14404\uc5a2\U00022a0e\U0002102e\u8a35yyyyzzyyyyyyyyyyyyyyzz#",
		"acd0saaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeee\U00021028\U0002d791\uc206\U0001b1dc\U000209bc\U000259fb\U00011113\U000243b9\u6d15\U00024d65\U0002106a\U0002a80b\U0001b230\u6664\U00025ce7\u2db9\U00020c6e",
		"acd2haaaaaaabbbbbbbccccdddddeeeeeeeeeeeeeeeee\U0001bc5f\u674b\U0002b17e\U0002ac44\U0002297a\uac5a\U000283b2\U0002db0e\U00017100\u49f6\U000217b9\U000293b7\U0002cc55\uca25\U00025270\U0002ab5d\U00025d5c",
		"acd3labbbbbbbbbbccccddddddeeeeeef\u59b8\U0002076a\U0002b5d2\U000265cd\u3780\U00023e1a\U0002d85e\U000294e1\u630b\U00027c99\u8ed2\ub80d\U0001d74b\u3c3b\U000211a9\U0002dd7a\U00023aaf",
		"abd4iaaaaaaaabccccddddddeeeeeeeeeeeeeef\u38f7\u3fca\u9e41\u7432",
		"x18372266066416\U0002ea3e\U00018a29yyyy?",
		"acd2saaaaaaaaaaaaaabbbbbbbbbbbbbbbccccddddddeeeeeeeeeeeeeeeeegggggghhhhhh\U00027b8d\u37c8\U0002acaa\U0002bef5\U00028739\u24db\ufb56\U0002dbdb\U0001d093\U00022972\U00023f66\ud7d9\U00029ac1\u4ef2\u4e29\u6f67\U000211e4\U00026269",
		"abd3Paaaaaabbbbbbbbbccccddddddeeeeeeeeeeeeeeef\U00023859\U00010f41\u4c52\U00029f1c\U0002d219\U0001682f\U0002b6a8\U00029eab\U000223d7",
		"acd8Kaabbbbbbccccddddddeeeeee\U00023fa1\u132e\U0002a79e\u3625\U00027281\U00026d0c\U00014561\U0001340b\u8219\ua4e8\u539b\U00024f60\u9cdc\U00027efa\U00021d20\u47d2\ud43b",
		"abd8sbbbbbbbccccddddeeeeeeeeef\U0002716a\U0002c5d7\u9884\u7fcf\uffc4\u3369\U0002b838\ufb8c\u0c1d\U0001d36f",
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
		allRunesAllowed(t, pass)
	}
}

func TestRegexpSpecialCaptures(t *testing.T) {
	var p RegexpParser
	p.SetSpecialCapture("word", func(*syntax.Regexp) (Template, error) {
		return EFFLargeWordlist(1), nil
	})

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
		allRunesAllowed(t, pass)
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
