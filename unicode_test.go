package password

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/rangetable"
)

func allRunesAllowed(t *testing.T, str string) {
	if idx := strings.IndexFunc(str, notAllowed); idx >= 0 {
		t.Helper()

		r, _ := utf8.DecodeRuneInString(str[idx:])
		t.Errorf("string contains prohibitted rune %U: %+q", r, str)
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
	notASCII := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0, Hi: 0x20 - 1, Stride: 1},
			{Lo: 0x7e + 1, Hi: 1<<16 - 1, Stride: 1},
		},
		R32: []unicode.Range32{
			{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1},
		},
		LatinOffset: 1,
	}

	asciiSpace := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: ' ', Hi: ' ', Stride: 1}},
		LatinOffset: 1,
	}
	tab := intersectRangeTables(asciiSpace, unicode.Z)
	assert.Equal(t, asciiSpace, tab)

	var runes1, runes2, runes3 []rune
	for _, tabs := range [][2]*unicode.RangeTable{
		{rangeTableASCII, allowedRangeTable},
		{rangeTableLatin1, allowedRangeTable},
		{stridedR16, allowedRangeTable},
		{stridedR32, allowedRangeTable},
		{stridedBoth, allowedRangeTable},
		{stridedR16, rangeTableASCII},
		{stridedR32, rangeTableASCII},
		{stridedBoth, rangeTableASCII},
		{rangetable.Merge(unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit), allowedRangeTable},
		{unicode.Latin, unicode.C},
		{unicode.Sc, unicode.S},
		{unicode.L, unicode.Lo},
		{
			&unicode.RangeTable{
				R16: []unicode.Range16{{Lo: 0, Hi: 1<<16 - 1, Stride: 1}},
				R32: []unicode.Range32{{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1}},
			},
			allowedRangeTable,
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
			allowedRangeTable,
		},
		{unicode.M, allowedRangeTable},
		{unicode.M, notASCII},
	} {
		t1 := naiveIntersectRangeTables(tabs[0], tabs[1])
		t2 := intersectRangeTables(unstridifyRangeTable(tabs[0]), tabs[1])
		t3 := intersectRangeTables(unstridifyRangeTable(tabs[1]), tabs[0])

		runes1 = runes1[:0]
		rangetable.Visit(t1, func(r rune) { runes1 = append(runes1, r) })

		runes2 = runes2[:0]
		rangetable.Visit(t2, func(r rune) { runes2 = append(runes2, r) })

		runes3 = runes3[:0]
		rangetable.Visit(t3, func(r rune) { runes3 = append(runes3, r) })

		require.Equal(t, runes1, runes2)
		require.Equal(t, runes1, runes3)
	}
}

func TestGeneratedRangeTables(t *testing.T) {
	rangeTableASCIIManual := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0x20, Hi: 0x7e, Stride: 1}},
		LatinOffset: 1,
	}
	assert.Equal(t, rangeTableASCIIManual, rangeTableASCII, "generated rangeTableASCII")

	if unicode.Version != unicodeVersion {
		t.Skipf("skipping rest of test due to mismatched unicode versions; have %s, want %s", unicode.Version, unicodeVersion)
	}

	allowedRangeTableManual := rangetable.Merge(
		unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
		unicode.N,
		unicode.P,
		unicode.Sm, unicode.Sc, unicode.So,
		rangeTableASCIIManual,
	)

	var runes1, runes2 []rune
	rangetable.Visit(allowedRangeTableManual, func(r rune) {
		if !unicode.In(r, unicode.Deprecated,
			unicode.Other_Default_Ignorable_Code_Point) {
			runes1 = append(runes1, r)
		}
	})
	rangetable.Visit(allowedRangeTable, func(r rune) { runes2 = append(runes2, r) })
	assert.Equal(t, runes1, runes2, "generated allowedRangeTable")
}

func TestAllowedRanges(t *testing.T) {
	for _, name := range []string{"C", "Lm", "M", "Sk", "Z"} {
		var runes []rune
		rangetable.Visit(unicode.Categories[name], func(r rune) {
			if unicode.Is(allowedRangeTable, r) && !unicode.Is(rangeTableASCII, r) &&
				// This is a temporary work around as Unicode 11.0.0
				// includes this in M whereas 10.0.0 includes it in P.
				r != 0x111c9 {
				runes = append(runes, r)
			}
		})

		assert.Empty(t, runes, "allowedRanges contains %d unwanted runes from category %s", len(runes), name)
	}

	for _, name := range []string{
		"Bidi_Control",
		"Deprecated",
		"Join_Control",
		"Noncharacter_Code_Point",
		"Other_Grapheme_Extend",
		"Other_Default_Ignorable_Code_Point",
		"Pattern_White_Space",
		"Prepended_Concatenation_Mark",
		"Variation_Selector",
		"White_Space",
	} {
		var runes []rune
		rangetable.Visit(unicode.Properties[name], func(r rune) {
			if unicode.Is(allowedRangeTable, r) && !unicode.Is(rangeTableASCII, r) {
				runes = append(runes, r)
			}
		})

		assert.Empty(t, runes, "allowedRanges contains %d unwanted runes with property %s", len(runes), name)
	}
}

func BenchmarkIntersectRangeTables(b *testing.B) {
	t1 := rangetable.Merge(
		unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit,
	)
	t1u := unstridifyRangeTable(t1)

	b.Run("naive", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			naiveIntersectRangeTables(t1, allowedRangeTable)
		}
	})
	b.Run("efficient", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			intersectRangeTables(t1u, allowedRangeTable)
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
