package passit

import (
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	var got, expect []rune
	for _, tabs := range []map[string]*unicode.RangeTable{
		unicode.Categories, unicode.Properties, unicode.Scripts,
	} {
		for name, tab := range tabs {
			got = got[:0]
			for i := range countRunesInTable(tab) {
				got = append(got, getRuneInTable(tab, i))
			}

			expect = expect[:0]
			rangetable.Visit(tab, func(r rune) {
				expect = append(expect, r)
			})

			require.Equal(t, expect, got, name)
		}
	}
}

func TestAddIntersectingRunes(t *testing.T) {
	rangeTableLatin1 := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: unicode.MaxLatin1, Stride: 1}},
		LatinOffset: 1,
	}
	stridedR16 := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: 128, Stride: 2}},
		LatinOffset: 1,
	}
	stridedR32 := &unicode.RangeTable{
		R32: []unicode.Range32{{Lo: 1 << 16, Hi: 1<<16 + 128, Stride: 2}},
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
	fullRange := &unicode.RangeTable{
		R16: []unicode.Range16{{Lo: 0, Hi: 1<<16 - 1, Stride: 1}},
		R32: []unicode.Range32{{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1}},
	}
	noLowerAZ := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0, Hi: 'a' - 1, Stride: 1},
			{Lo: 'z' + 1, Hi: 1<<16 - 1, Stride: 1},
		},
		R32: []unicode.Range32{
			{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1},
		},
		LatinOffset: 1,
	}
	multiRangeTable1 := rangetable.Merge(
		unicode.Latin,
		unicode.Greek,
		unicode.Cyrillic,
		unicode.ASCII_Hex_Digit)
	multiRangeTable2 := rangetable.Merge(
		unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
		unicode.N,
		unicode.P,
		unicode.Sm, unicode.Sc, unicode.So,
		rangeTableASCII)

	asciiSpace := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: ' ', Hi: ' ', Stride: 1}},
		LatinOffset: 1,
	}
	tab := intersectRangeTables(asciiSpace, unicode.Z)
	assert.Equal(t, asciiSpace, tab)

	var runes1, runes2, runes3 []rune
	for _, tabs := range [][2]*unicode.RangeTable{
		{rangeTableASCII, multiRangeTable2},
		{rangeTableLatin1, multiRangeTable2},
		{stridedR16, multiRangeTable2},
		{stridedR32, multiRangeTable2},
		{stridedBoth, multiRangeTable2},
		{stridedR16, rangeTableASCII},
		{stridedR32, rangeTableASCII},
		{stridedBoth, rangeTableASCII},
		{multiRangeTable1, multiRangeTable2},
		{unicode.Latin, unicode.C},
		{unicode.Sc, unicode.S},
		{unicode.L, unicode.Lo},
		{fullRange, multiRangeTable2},
		{noLowerAZ, multiRangeTable2},
		{unicode.M, multiRangeTable2},
		{unicode.M, notASCII},
	} {
		t1 := naiveIntersectRangeTables(tabs[0], tabs[1])
		t2 := intersectRangeTables(tabs[0], tabs[1])
		t3 := intersectRangeTables(tabs[1], tabs[0])

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

func TestRemoveNLFromRangeTable(t *testing.T) {
	rangeTableLatin1 := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: unicode.MaxLatin1, Stride: 1}},
		LatinOffset: 1,
	}
	stridedR16 := &unicode.RangeTable{
		R16:         []unicode.Range16{{Lo: 0, Hi: 128, Stride: 2}},
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
	fullRange := &unicode.RangeTable{
		R16: []unicode.Range16{{Lo: 0, Hi: 1<<16 - 1, Stride: 1}},
		R32: []unicode.Range32{{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1}},
	}
	noLowerAZ := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0, Hi: 'a' - 1, Stride: 1},
			{Lo: 'z' + 1, Hi: 1<<16 - 1, Stride: 1},
		},
		R32: []unicode.Range32{
			{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1},
		},
		LatinOffset: 1,
	}
	onlyNL := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: '\n', Hi: '\n', Stride: 1},
		},
		LatinOffset: 1,
	}
	multiRangeTable1 := rangetable.Merge(
		onlyNL,
		unicode.Latin,
		unicode.Greek,
		unicode.Cyrillic,
		unicode.ASCII_Hex_Digit)
	multiRangeTable2 := rangetable.Merge(
		onlyNL,
		unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
		unicode.N,
		unicode.P,
		unicode.Sm, unicode.Sc, unicode.So,
		rangeTableASCII)

	rangeTableNoNL := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0, Hi: '\n' - 1, Stride: 1},
			{Lo: '\n' + 1, Hi: 1<<16 - 1, Stride: 1},
		},
		R32: []unicode.Range32{
			{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1},
		},
		LatinOffset: 1,
	}

	var runes1, runes2 []rune
	for i, tab := range []*unicode.RangeTable{
		rangeTableLatin1,
		stridedR16,
		stridedBoth,
		notASCII,
		fullRange,
		noLowerAZ,
		onlyNL,
		multiRangeTable1,
		multiRangeTable2,
	} {
		if !unicode.Is(tab, '\n') {
			t.Log(i)
			panic("table does not contain newline")
		}

		t1 := intersectRangeTables(tab, rangeTableNoNL)
		t2 := removeNLFromRangeTable(tab)

		ct := countRunesInTable(tab)
		c2 := countRunesInTable(t2)
		require.Equalf(t, ct-1, c2, "table should contain one less rune (i=%d)", i)

		runes1 = runes1[:0]
		rangetable.Visit(t1, func(r rune) { runes1 = append(runes1, r) })

		runes2 = runes2[:0]
		rangetable.Visit(t2, func(r rune) { runes2 = append(runes2, r) })

		require.Equalf(t, runes1, runes2, "tables should contain same runes (i=%d)", i)
	}
}

func BenchmarkAddIntersectingRunes(b *testing.B) {
	tab1 := rangetable.Merge(
		unicode.Latin,
		unicode.Greek,
		unicode.Cyrillic,
		unicode.ASCII_Hex_Digit)
	tab2 := rangetable.Merge(
		unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
		unicode.N,
		unicode.P,
		unicode.Sm, unicode.Sc, unicode.So,
		rangeTableASCII)

	b.Run("naive", func(b *testing.B) {
		for range b.N {
			_ = naiveIntersectRangeTables(tab1, tab2)
		}
	})
	b.Run("efficient", func(b *testing.B) {
		for range b.N {
			_ = intersectRangeTables(tab1, tab2)
		}
	})
}

func BenchmarkRemoveNLFromRangeTable(b *testing.B) {
	onlyNL := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: '\n', Hi: '\n', Stride: 1},
		},
		LatinOffset: 1,
	}
	noNL := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0, Hi: '\n' - 1, Stride: 1},
			{Lo: '\n' + 1, Hi: 1<<16 - 1, Stride: 1},
		},
		R32: []unicode.Range32{
			{Lo: 1 << 16, Hi: unicode.MaxRune, Stride: 1},
		},
		LatinOffset: 1,
	}
	tab := rangetable.Merge(
		onlyNL,
		unicode.Latin,
		unicode.Greek,
		unicode.Cyrillic,
		unicode.ASCII_Hex_Digit)

	b.Run("naive", func(b *testing.B) {
		for range b.N {
			_ = naiveRemoveNLFromRangeTable(tab)
		}
	})
	b.Run("intersect", func(b *testing.B) {
		for range b.N {
			_ = intersectRangeTables(noNL, tab)
		}
	})
	b.Run("efficient", func(b *testing.B) {
		for range b.N {
			_ = removeNLFromRangeTable(tab)
		}
	})
}

func visitRanges(tab *unicode.RangeTable, fn func(lo, hi rune)) {
	for i := range tab.R16 {
		range_ := &tab.R16[i]
		if range_.Stride == 1 {
			fn(rune(range_.Lo), rune(range_.Hi))
			continue
		}
		for r := range_.Lo; range_.Lo <= r && r <= range_.Hi; r += range_.Stride {
			fn(rune(r), rune(r))
		}
	}

	for i := range tab.R32 {
		range_ := &tab.R32[i]
		if range_.Stride == 1 {
			fn(rune(range_.Lo), rune(range_.Hi))
			continue
		}
		for r := range_.Lo; range_.Lo <= r && r <= range_.Hi; r += range_.Stride {
			fn(rune(r), rune(r))
		}
	}
}

func naiveAddIntersectingRunes(runes []rune, lo, hi rune, parent *unicode.RangeTable) []rune {
	runes = slices.Grow(runes, int(hi-lo))
	for r := lo; r <= hi; r++ {
		if unicode.Is(parent, r) {
			runes = append(runes, r)
		}
	}

	return runes
}

func naiveIntersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	var runes []rune
	visitRanges(a, func(lo, hi rune) {
		runes = naiveAddIntersectingRunes(runes, lo, hi, b)
	})
	return rangetable.New(runes...)
}

func intersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	var rt unicode.RangeTable
	visitRanges(a, func(lo, hi rune) {
		addIntersectingRunes(&rt, lo, hi, b)
	})
	setLatinOffset(&rt)
	return &rt
}

func naiveRemoveNLFromRangeTable(tab *unicode.RangeTable) *unicode.RangeTable {
	runes := make([]rune, 0, countRunesInTable(tab))
	rangetable.Visit(tab, func(r rune) {
		if r != '\n' {
			runes = append(runes, r)
		}
	})
	return rangetable.New(runes...)
}
