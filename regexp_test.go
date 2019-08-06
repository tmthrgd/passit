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

		matchPattern := "^" + pattern + "$"
		assert.Truef(t, regexp.MustCompile(matchPattern).MatchString(pass),
			"regexp.MustCompile(%q).MatchString(%q)", matchPattern, pass)
	}
}

func TestUnstridifyRangeTable(t *testing.T) {
	rt := unstridifyRangeTable(rangetable.Merge(unicode.C))

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
	b := regexpAnyRangeTable(RegexpUnicodeAny)

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
	t2 := regexpAnyRangeTable(RegexpUnicodeAny)

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
