package password

import (
	"math/rand"
	"regexp"
	"regexp/syntax"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/rangetable"
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
		"x90822236719\U00027618\U00016820\U000216df\u4ed6yyyyzzzzyy=",
		"x7977150zzyyyyzzzzzzyyzzzzyyyy<",
		"x14404\U00028f55\u4e68\U0001d734\u096fyyyyzzyyyyyyyyyyyyyyzz#",
		"acd0saaaaaabbbbbbbbbbbbbbbbccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhhhhhhhhhh\U00028617\U00024a948>G;K",
		"abd9Xaaaaaaaaabbbbbbbbbbbbbbccccddddeeeeeeeeeggggggghhhhhhhhhhhhhhh\U000131bd\U00027390Z6",
		"x30\u230a\U0002357e\ua219\U00026243\U00017e4bzzzzyyzzyyzzzzyyzzzz(",
		"x67927673\U0002aaec\U0002400d\U000273c0\u4580yyyyyyyyyyyyyyyyzzyyyyzz^",
		"x21903591837\u69fc\u31f9\U0002b611\U00011684yyyyzzyyyyyyyy<",
		"abd6Faaabbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeeef\U000288eb\U0002b76b/4\"\")@",
		"x7055\U0001d728\U0001740f\U0002aa29zzyyyyyyzzyyyyyyzzyyzz?",
		"acd8Iabbbbbbbbbbbbccccddddddeeeeeeeee\u9c38\u241dI^=",
		"abd8Vaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeggggggggggggghhhhhhhhhhhhhhhh\U000121b5\U00028fb7&Z7<,9{7\U000100057I\U00010002",
		"acd1Yaaaaaaaaaaabbbbbbbbbbbbbbbccccdddddeeeeeeeeeeef\U00016974\U00021933<[PZ\U00010006IE",
	} {
		pass, err := tmpl.Password(testRand)
		require.NoError(t, err)

		assert.Equal(t, expect, pass)

		matchPattern := "^(?:" + pattern + ")$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%+q)", matchPattern, pass)
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
	}
}

func TestUnstridifyRangeTable(t *testing.T) {
	// This will trip the race detector if unicode.C is modified.
	go func() { _ = *unicode.C }()

	oldC := *unicode.C
	rt := unstridifyRangeTable(unicode.C)
	assert.Equal(t, &oldC, unicode.C, "mutated input")

	for _, r16 := range rt.R16 {
		require.Equal(t, uint16(1), r16.Stride)
	}
	for _, r32 := range rt.R32 {
		require.Equal(t, uint32(1), r32.Stride)
	}
}

func TestIntersectRangeTables(t *testing.T) {
	a := unstridifyRangeTable(rangetable.Merge(
		unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit,
	))

	var p RegexpParser
	p.SetUnicodeAny()
	b := p.anyRangeTable()

	t1 := intersectRangeTables(a, b)
	t2 := naiveIntersectRangeTables(a, b)

	var diff1 []rune
	rangetable.Visit(t1, func(r rune) {
		if !unicode.Is(t2, r) {
			diff1 = append(diff1, r)
		}
	})
	assert.Empty(t, diff1, "entries in intersectRangeTables not in naiveIntersectRangeTables")

	var diff2 []rune
	rangetable.Visit(t2, func(r rune) {
		if !unicode.Is(t1, r) {
			diff2 = append(diff2, r)
		}
	})
	assert.Empty(t, diff1, "entries in naiveIntersectRangeTables not in intersectRangeTables")

	for i := 0; i < len(t1.R16)-1; i++ {
		require.True(t, t1.R16[i].Lo <= t1.R16[i].Hi &&
			t1.R16[i].Hi < t1.R16[i+1].Lo,
			"not sorted or overlap")
	}
	for i := 0; i < len(t1.R32)-1; i++ {
		require.True(t, t1.R32[i].Lo <= t1.R32[i].Hi &&
			t1.R32[i].Hi < t1.R32[i+1].Lo,
			"not sorted or overlap")
	}
}

func BenchmarkIntersectRangeTables(b *testing.B) {
	t1 := unstridifyRangeTable(rangetable.Merge(
		unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit,
	))

	var p RegexpParser
	p.SetUnicodeAny()
	t2 := p.anyRangeTable()

	b.Run("naive", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			naiveIntersectRangeTables(t1, t2)
		}
	})
	b.Run("efficient", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			intersectRangeTables(t1, t2)
		}
	})
}

func naiveIntersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	// Iterate over the smaller table.
	if countTableRunes(a) > countTableRunes(b) {
		a, b = b, a
	}

	var rt unicode.RangeTable
	rangetable.Visit(a, func(r rune) {
		if !unicode.Is(b, r) {
			return
		}

		const maxR16 = 1<<16 - 1
		if r <= maxR16 {
			rt.R16 = append(rt.R16, unicode.Range16{
				Lo:     uint16(r),
				Hi:     uint16(r),
				Stride: 1,
			})
		} else {
			rt.R32 = append(rt.R32, unicode.Range32{
				Lo:     uint32(r),
				Hi:     uint32(r),
				Stride: 1,
			})
		}
	})

	return rangetable.Merge(&rt)
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
