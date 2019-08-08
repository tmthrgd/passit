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
	if !hasUnicode10 {
		t.Skip("skipping without unicode 10.0.0")
	}

	pattern := `a[bc]d[0-9][^\x00-AZ-az-\x{10FFFF}]a*b+c{4}d{3,6}e{5,}f?(g+h+)?.{2}[^a-z]+|x[0-9]+?.{0,5}(?:yy|zz)+(?P<punct>[[:punct:]])`

	var p RegexpParser
	p.SetUnicodeAny()

	tmpl, err := p.Parse(pattern, syntax.Perl)
	require.NoError(t, err)

	testRand := rand.New(rand.NewSource(0))

	for _, expect := range []string{
		"x90822236719\U000281f1\U00020c7c\u9158\U0002beaeyyyyzzzzyy=",
		"x7977150zzyyyyzzzzzzyyzzzzyyyy<",
		"x14404\U0002af07\U000225f8\uc9de\u925dyyyyzzyyyyyyyyyyyyyyzz#",
		"acd0saaaaaabbbbbbbbbbbbbbbbccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhhhhhhhhhh\u89bb\U00017007\U00010b45\U0002084f\uaf59\U0002abb0\U0001b04a",
		"abd9Xaaaaaaaaabbbbbbbbbbbbbbccccddddeeeeeeeeeggggggghhhhhhhhhhhhhhh\u231f\U0002b3aa\U0002d1c6\U0002a316",
		"x30\U000168ba\u6d97\uae28\U00026d49\u41ebzzzzyyzzyyzzzzyyzzzz(",
		"x67927673\u6c23\u3c57\ua1ad\U0002cc4eyyyyyyyyyyyyyyyyzzyyyyzz^",
		"x21903591837\u5674\u0f6a\U00029475\u698cyyyyzzyyyyyyyy<",
		"abd6Faaabbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeeef\ua98f\U000295f8\U00021d16\U000205ad\U00021c08\U000274b4\U0001072b\ud269",
		"x7055\U0001f7b4\U000207b5\U00024a4ezzyyyyyyzzyyyyyyzzyyzz?",
		"acd8Iabbbbbbbbbbbbccccddddddeeeeeeeee\U00020f75\u6abd\uae8c\u6b30\u3dcd",
		"abd8Vaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeggggggggggggghhhhhhhhhhhhhhhh\u13ac\u6679\U000216eb\U0002e96b\U0001854a\ub42d\U0002bb27\U0002deb7\uce48\u0761\U00023f28\U0002dd20\u8beb\U00012452",
		"acd1Yaaaaaaaaaaabbbbbbbbbbbbbbbccccdddddeeeeeeeeeeef\u77d9\U000189db\u55c0\U0002ddba\ua695\U000271fb\U00028c40\U000214f4\U0002346e",
	} {
		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)

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
