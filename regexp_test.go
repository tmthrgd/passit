package password

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
		"acd0saaaaaabbbbbbbbbbbbbbbbccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhhhhhhhhhhy;>;,2E",
		"abd9Xaaaaaaaaabbbbbbbbbbbbbbccccddddeeeeeeeeeggggggghhhhhhhhhhhhhhh.L]Q",
		"x30J%PnHzzzzyyzzyyzzzzyyzzzz(",
		"x67927673}Oc)yyyyyyyyyyyyyyyyzzyyyyzz^",
		"x21903591837hb?*yyyyzzyyyyyyyy<",
		"abd6FaaabbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeeefL}VR1[\\I",
		"x7055,~Fzzyyyyyyzzyyyyyyzzyyzz?",
		"acd8IabbbbbbbbbbbbccccddddddeeeeeeeeesML[L",
		"abd8VaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeggggggggggggghhhhhhhhhhhhhhhhR<YB[EGN[^~+.@",
		"acd1Yaaaaaaaaaaabbbbbbbbbbbbbbbccccdddddeeeeeeeeeeef{*`\"8$S%H",
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
		"x90822236719\U0002c37c\uad54\u7839\u8bfbyyyyzzzzyy=",
		"x7977150zzyyyyzzzzzzyyzzzzyyyy<",
		"x14404\U00021afa\U0002e207\U0001d6b5\U00020972yyyyzzyyyyyyyyyyyyyyzz#",
		"acd0saaaaaabbbbbbbbbbbbbbbbccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhhhhhhhhhh\U00017ff4\U0001847f\U00023fad\U00027948\u509e\ub6e7\u5e00",
		"abd9Xaaaaaaaaabbbbbbbbbbbbbbccccddddeeeeeeeeeggggggghhhhhhhhhhhhhhh\u14a8\ud493\u9f66\U00022b99",
		"x30\U00025696\uadac\U0002749c\U0002ca0e\ub694zzzzyyzzyyzzzzyyzzzz(",
		"x67927673\U00029c0b\u9e58\U00026af0\U0001815cyyyyyyyyyyyyyyyyzzyyyyzz^",
		"x21903591837\ud211\u8c73\U000210b8\ua663yyyyzzyyyyyyyy<",
		"abd6Faaabbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeeef\u2922\U0002e7cc\U0002c000\U0002b98b\u616a\u6349\U0002e5d0\U00022476",
		"x7055\u201d\u0662\U0002b32czzyyyyyyzzyyyyyyzzyyzz?",
		"acd8Iabbbbbbbbbbbbccccddddddeeeeeeeee\U0002bcc1\U000122aa\u29ff\U00027d31\U00028606",
		"abd8Vaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeggggggggggggghhhhhhhhhhhhhhhh\u7872\U000100e6\u8c9f\u74f4\ud1a7\U0001817b\U0002ebb9\u66c5\u96da\U00017186\U00026144\u2b0b\u041d\u993c",
		"acd1Yaaaaaaaaaaabbbbbbbbbbbbbbbccccdddddeeeeeeeeeeef\u228c\U0002c3c1\ub291\u5af7\U0001f022\u226a\U0002ad94\U00017c37\u8c40",
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
		return DefaultWords(1), nil
	})

	tmpl, err := p.Parse(`((?P<word>) ){6}[[:upper:]][[:digit:]][[:punct:]]`, syntax.PerlX)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"remover dismay vocation sepia backtalk think S3)",
		"finance obscure dusk rigor hemlock dusk U8;",
		"zoning say shrug actress swirl cross L9~",
		"stiffen deduct amigo outmatch viral shrimp M4\\",
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
