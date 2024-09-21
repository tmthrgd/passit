package passit

import (
	"slices"
	"unicode"
)

func countRunesInTable(tab *unicode.RangeTable) int {
	var c int
	for _, r16 := range tab.R16 {
		c += int((r16.Hi-r16.Lo)/r16.Stride) + 1
	}
	for _, r32 := range tab.R32 {
		c += int((r32.Hi-r32.Lo)/r32.Stride) + 1
	}

	return c
}

func getRuneInTable(tab *unicode.RangeTable, v int) rune {
	for _, r16 := range tab.R16 {
		size := int((r16.Hi-r16.Lo)/r16.Stride) + 1
		if v < size {
			return rune(r16.Lo + uint16(v)*r16.Stride)
		}
		v -= size
	}

	for _, r32 := range tab.R32 {
		size := int((r32.Hi-r32.Lo)/r32.Stride) + 1
		if v < size {
			return rune(r32.Lo + uint32(v)*r32.Stride)
		}
		v -= size
	}

	panic("passit: index out of range of unicode.RangeTable")
}

func addIntersectingRunes(tab *unicode.RangeTable, lo, hi rune, parent *unicode.RangeTable) {
	const maxR16 = 1<<16 - 1
	if lo > maxR16 {
		addIntersectingRunes32(tab, lo, hi, parent)
		return
	}
	if hi > maxR16 {
		addIntersectingRunes32(tab, maxR16+1, hi, parent)
		hi = maxR16
	}
	addIntersectingRunes16(tab, lo, hi, parent)
}

func addIntersectingRunes16(tab *unicode.RangeTable, lo, hi rune, parent *unicode.RangeTable) {
	for i := range parent.R16 {
		range_ := &parent.R16[i]
		if hi < rune(range_.Lo) {
			break
		} else if lo > rune(range_.Hi) {
			continue
		}

		iLo, iHi, stride := intersection(lo, hi, rune(range_.Lo), rune(range_.Hi), rune(range_.Stride))
		if iLo <= iHi {
			tab.R16 = append(tab.R16, unicode.Range16{Lo: uint16(iLo), Hi: uint16(iHi), Stride: uint16(stride)})
		}
	}
}

func addIntersectingRunes32(tab *unicode.RangeTable, lo, hi rune, parent *unicode.RangeTable) {
	for i := range parent.R32 {
		range_ := &parent.R32[i]
		if hi < rune(range_.Lo) {
			break
		} else if lo > rune(range_.Hi) {
			continue
		}

		iLo, iHi, stride := intersection(lo, hi, rune(range_.Lo), rune(range_.Hi), rune(range_.Stride))
		if iLo <= iHi {
			tab.R32 = append(tab.R32, unicode.Range32{Lo: uint32(iLo), Hi: uint32(iHi), Stride: uint32(stride)})
		}
	}
}

func removeNLFromRangeTable(tab *unicode.RangeTable) *unicode.RangeTable {
	// How this works is that we find the unicode.Range16 that contains \n, and
	// then we split it either side of the \n, creating either 0, 1 or 2 new
	// ranges. We can ignore the unicode.Range32 tables as \n will always be in
	// the 16-bit tables.
	//
	// We're only called on unicode.RangeTable's that contain \n so we don't
	// have to check it exists first.

	idx := -1
	for i := range tab.R16 {
		range_ := &tab.R16[i]
		if range_.Lo <= '\n' && '\n' <= range_.Hi {
			idx = i
			break
		}
	}
	range_ := &tab.R16[idx]

	var rt unicode.RangeTable
	rt.R16 = make([]unicode.Range16, idx, len(tab.R16)+1)
	copy(rt.R16, tab.R16)

	lo, hi, stride := intersection(0, '\n'-1, rune(range_.Lo), rune(range_.Hi), rune(range_.Stride))
	if lo <= hi {
		rt.R16 = append(rt.R16, unicode.Range16{Lo: uint16(lo), Hi: uint16(hi), Stride: uint16(stride)})
	}

	lo, hi, stride = intersection('\n'+1, unicode.MaxRune, rune(range_.Lo), rune(range_.Hi), rune(range_.Stride))
	if lo <= hi {
		rt.R16 = append(rt.R16, unicode.Range16{Lo: uint16(lo), Hi: uint16(hi), Stride: uint16(stride)})
	}

	rt.R16 = append(rt.R16, tab.R16[idx+1:]...)
	rt.R32 = slices.Clip(tab.R32)
	setLatinOffset(&rt)
	return &rt
}

func setLatinOffset(tab *unicode.RangeTable) {
	tab.LatinOffset = len(tab.R16)
	for i := range tab.R16 {
		if tab.R16[i].Hi > unicode.MaxLatin1 {
			tab.LatinOffset = i
			break
		}
	}
}

func intersection(lo0, hi0, lo1, hi1, stride1 rune) (lo, hi, stride rune) {
	if stride1 == 1 {
		return max(lo0, lo1), min(hi0, hi1), 1
	}

	lo = lo1
	if lo < lo0 {
		c := lo0 - lo1
		c += stride1 - 1
		c -= c % stride1
		lo += c
	}

	hi = hi1
	if hi > hi0 {
		c := hi1 - hi0
		c += stride1 - 1
		c -= c % stride1
		hi -= c
	}

	if lo == hi {
		return lo, hi, 1
	}
	return lo, hi, stride1
}
