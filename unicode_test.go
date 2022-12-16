package passit

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"golang.org/x/text/unicode/rangetable"
)

func allRunesAllowed(t *testing.T, allowed any, str string) {
	var notAllowed func(r rune) bool
	switch allowed := allowed.(type) {
	case string:
		notAllowed = func(r rune) bool {
			return !strings.ContainsRune(allowed, r)
		}
	case *unicode.RangeTable:
		notAllowed = func(r rune) bool {
			return !unicode.Is(allowed, r)
		}
	default:
		panic("passit: unsupported allowed argument type")
	}

	if idx := strings.IndexFunc(str, notAllowed); idx >= 0 {
		t.Helper()

		r, _ := utf8.DecodeRuneInString(str[idx:])
		t.Errorf("string contains prohibited rune %U: %+q", r, str)
	}
}

func TestCountRunesInTable(t *testing.T) {
	for _, tabs := range []map[string]*unicode.RangeTable{
		unicode.Categories, unicode.Properties, unicode.Scripts,
	} {
		for name, tab := range tabs {
			got := countRunesInTable(tab)

			var expect int
			rangetable.Visit(tab, func(rune) { expect++ })

			assert.Equal(t, expect, got, name)
		}
	}
}

func TestGetRuneInTable(t *testing.T) {
	for _, tabs := range []map[string]*unicode.RangeTable{
		unicode.Categories, unicode.Properties, unicode.Scripts,
	} {
		for name, tab := range tabs {
			var got []rune
			for i, c := 0, countRunesInTable(tab); i < c; i++ {
				got = append(got, getRuneInTable(tab, i))
			}

			var expect []rune
			rangetable.Visit(tab, func(r rune) {
				expect = append(expect, r)
			})

			require.Equal(t, expect, got, name)
		}
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
	multiRangeTable := rangetable.Merge(
		unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
		unicode.N,
		unicode.P,
		unicode.Sm, unicode.Sc, unicode.So,
		rangeTableASCII,
	)

	asciiSpace := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: ' ', Hi: ' ', Stride: 1}},
		LatinOffset: 1,
	}
	tab := intersectRangeTables(asciiSpace, unicode.Z)
	assert.Equal(t, asciiSpace, tab)

	var runes1, runes2, runes3 []rune
	for _, tabs := range [][2]*unicode.RangeTable{
		{rangeTableASCII, multiRangeTable},
		{rangeTableLatin1, multiRangeTable},
		{stridedR16, multiRangeTable},
		{stridedR32, multiRangeTable},
		{stridedBoth, multiRangeTable},
		{stridedR16, rangeTableASCII},
		{stridedR32, rangeTableASCII},
		{stridedBoth, rangeTableASCII},
		{rangetable.Merge(unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit), multiRangeTable},
		{unicode.Latin, unicode.C},
		{unicode.Sc, unicode.S},
		{unicode.L, unicode.Lo},
		{
			&unicode.RangeTable{
				R16: []unicode.Range16{{Lo: 0, Hi: 1<<16 - 1, Stride: 1}},
				R32: []unicode.Range32{{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1}},
			},
			multiRangeTable,
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
			multiRangeTable,
		},
		{unicode.M, multiRangeTable},
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

func BenchmarkIntersectRangeTables(b *testing.B) {
	t1 := rangetable.Merge(
		unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.ASCII_Hex_Digit,
	)
	t1u := unstridifyRangeTable(t1)

	multiRangeTable := rangetable.Merge(
		unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
		unicode.N,
		unicode.P,
		unicode.Sm, unicode.Sc, unicode.So,
		rangeTableASCII,
	)

	b.Run("naive", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			naiveIntersectRangeTables(t1, multiRangeTable)
		}
	})
	b.Run("efficient", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			intersectRangeTables(t1u, multiRangeTable)
		}
	})
}

func naiveIntersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	// Iterate over the smaller table.
	if countRunesInTable(a) > countRunesInTable(b) {
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
		R16: slices.Clip(tab.R16),
		R32: slices.Clip(tab.R32),
	}

	for i := 0; i < len(rt.R16); i++ {
		if r16 := rt.R16[i]; r16.Stride != 1 {
			add := make([]unicode.Range16, 0, (r16.Hi-r16.Lo)/r16.Stride+1)
			for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
				if r <= unicode.MaxLatin1 {
					rt.LatinOffset++
				}

				add = append(add, unicode.Range16{Lo: uint16(r), Hi: uint16(r), Stride: 1})
			}

			rt.R16 = slices.Replace(rt.R16, i, i+1, add...)
			i += len(add) - 1
		} else if r16.Hi <= unicode.MaxLatin1 {
			rt.LatinOffset++
		}
	}

	for i := 0; i < len(rt.R32); i++ {
		if r32 := rt.R32[i]; r32.Stride != 1 {
			add := make([]unicode.Range32, 0, (r32.Hi-r32.Lo)/r32.Stride+1)
			for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
				add = append(add, unicode.Range32{Lo: uint32(r), Hi: uint32(r), Stride: 1})
			}

			rt.R32 = slices.Replace(rt.R32, i, i+1, add...)
			i += len(add) - 1
		}
	}

	return rt
}
