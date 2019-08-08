package password

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/rangetable"
)

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
		t1 := intersectRangeTables(unstridifyRangeTable(tabs[0]), tabs[1])
		t2 := naiveIntersectRangeTables(tabs[0], tabs[1])

		runes1 = runes1[:0]
		rangetable.Visit(t1, func(r rune) { runes1 = append(runes1, r) })

		runes2 = runes2[:0]
		rangetable.Visit(t2, func(r rune) { runes2 = append(runes2, r) })

		require.Equal(t, runes2, runes1)
	}
}

func BenchmarkIntersectRangeTables(b *testing.B) {
	t1 := rangetable.Merge(
		unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit,
	)
	t1u := unstridifyRangeTable(t1)

	t2 := rangetable.Merge(allowedRanges...)

	b.Run("naive", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			naiveIntersectRangeTables(t1, t2)
		}
	})
	b.Run("efficient", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			intersectRangeTables(t1u, t2)
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

func unstridifyRangeTable(tab *unicode.RangeTable) *unicode.RangeTable {
	rt := &unicode.RangeTable{
		R16: tab.R16[:len(tab.R16):len(tab.R16)],
		R32: tab.R32[:len(tab.R32):len(tab.R32)],
	}

	for i := 0; i < len(rt.R16); i++ {
		if r16 := rt.R16[i]; r16.Stride != 1 {
			size := int((r16.Hi-r16.Lo)/r16.Stride) + 1
			rt.R16 = append(rt.R16, make([]unicode.Range16, size-1)...)
			copy(rt.R16[i+size:], rt.R16[i+1:])

			for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
				if r <= unicode.MaxLatin1 {
					rt.LatinOffset++
				}

				rt.R16[i] = unicode.Range16{Lo: uint16(r), Hi: uint16(r), Stride: 1}
				i++
			}
			i--
		} else if r16.Hi <= unicode.MaxLatin1 {
			rt.LatinOffset++
		}
	}

	for i := 0; i < len(rt.R32); i++ {
		if r32 := rt.R32[i]; r32.Stride != 1 {
			size := int((r32.Hi-r32.Lo)/r32.Stride) + 1
			rt.R32 = append(rt.R32, make([]unicode.Range32, size-1)...)
			copy(rt.R32[i+size:], rt.R32[i+1:])

			for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
				rt.R32[i] = unicode.Range32{Lo: uint32(r), Hi: uint32(r), Stride: 1}
				i++
			}
			i--
		}
	}

	return rt
}
