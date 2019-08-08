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
		"acd0saaaaaabbbbbbbbbbbbbbbbccccddddeeeeeeeeeeeeeeeeegggggggghhhhhhhhhhhhhhh\U00028617\U00024a94\U00017db0\U00027fb3\U00029c0f\u4fdc\U0001f5cc",
		"abd9Xaaaaaaaaabbbbbbbbbbbbbbccccddddeeeeeeeeeggggggghhhhhhhhhhhhhhh\U000131bd\U00027390\U0002a346\u32ba",
		"x30\u230a\U0002357e\ua219\U00026243\U00017e4bzzzzyyzzyyzzzzyyzzzz(",
		"x67927673\U0002aaec\U0002400d\U000273c0\u4580yyyyyyyyyyyyyyyyzzyyyyzz^",
		"x21903591837\u69fc\u31f9\U0002b611\U00011684yyyyzzyyyyyyyy<",
		"abd6Faaabbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeeeeeef\U000288eb\U0002b76b\u8c00\u41c1\U00016b5e\U0002bd02\U00026d44\u35af",
		"x7055\U0001d728\U0001740f\U0002aa29zzyyyyyyzzyyyyyyzzyyzz?",
		"acd8Iabbbbbbbbbbbbccccddddddeeeeeeeee\u9c38\u241d\U000187ca\U0002d4fa\U0002f860",
		"abd8Vaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccdddddeeeeeeeeeeeeeeggggggggggggghhhhhhhhhhhhhhhh\U000121b5\U00028fb7\ucf9f\U000228e5\ud367\U0001d803\U00026306\u3cd6\ua88e\U0001335d\U0002598a\ua3ad\U00018841\ub51b",
		"acd1Yaaaaaaaaaaabbbbbbbbbbbbbbbccccdddddeeeeeeeeeeef\U00016974\U00021933\U0002b008\U0001207d\U0002464b\u361e\u540e\U0002dfab\U00024f6a",
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
	allowed := rangetable.Merge(allowedRanges...)
	rangeTableLatin1 := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: unicode.MaxLatin1, Stride: 1}},
		LatinOffset: 1,
	}
	stridedR16 := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: 128, Stride: 2}},
		LatinOffset: 1,
	}
	stridedR32 := &unicode.RangeTable{
		R32:         []unicode.Range32{{Lo: 1 << 16, Hi: 1<<16 + 128, Stride: 2}},
		LatinOffset: 1,
	}
	stridedBoth := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: 128, Stride: 2}},
		R32:         []unicode.Range32{{Lo: 1 << 16, Hi: 1<<16 + 128, Stride: 2}},
		LatinOffset: 1,
	}

	var runes1, runes2 []rune
	for _, tabs := range [][2]*unicode.RangeTable{
		{rangeTableASCII, allowed},
		{rangeTableLatin1, allowed},
		{allowed, rangeTableASCII},
		{allowed, rangeTableLatin1},
		{stridedR16, allowed},
		{allowed, stridedR16},
		{stridedR32, allowed},
		{allowed, stridedR32},
		{stridedBoth, allowed},
		{allowed, stridedBoth},
		{stridedR16, rangeTableASCII},
		{rangeTableASCII, stridedR16},
		{stridedR32, rangeTableASCII},
		{rangeTableASCII, stridedR32},
		{stridedBoth, rangeTableASCII},
		{rangeTableASCII, stridedBoth},
		{rangetable.Merge(unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit), allowed},
		{unicode.Latin, unicode.C},
		{unicode.Sc, unicode.S},
		{unicode.L, unicode.Lo},
		{
			&unicode.RangeTable{
				R16: []unicode.Range16{{Lo: 0, Hi: 1<<16 - 1, Stride: 1}},
				R32: []unicode.Range32{{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1}},
			},
			allowed,
		},
		{
			&unicode.RangeTable{
				R16: []unicode.Range16{
					{Lo: 0, Hi: 'a' - 1, Stride: 1},
					{Lo: 'z' + 1, Hi: 1<<16 - 1, Stride: 1},
				},
				R32: []unicode.Range32{
					{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1},
				},
				LatinOffset: 1,
			},
			allowed,
		},
	} {
		t1 := intersectRangeTables(unstridifyRangeTable(tabs[0]), unstridifyRangeTable(tabs[1]))
		t2 := naiveIntersectRangeTables(tabs[0], tabs[1])

		runes1 = runes1[:0]
		rangetable.Visit(t1, func(r rune) { runes1 = append(runes1, r) })

		runes2 = runes2[:0]
		rangetable.Visit(t2, func(r rune) { runes2 = append(runes2, r) })

		require.Equal(t, runes2, runes1)
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

			if r <= unicode.MaxLatin1 {
				rt.LatinOffset++
			}
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
