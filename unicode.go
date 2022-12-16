package passit

import "unicode"

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

	panic("passit: internal error: unicode.RangeTable did not contain rune")
}

func appendToRangeTable(tab *unicode.RangeTable, lo, hi rune) {
	const maxR16 = 1<<16 - 1

	if lo > hi {
		panic("passit: lo rune is greater than hi rune")
	}
	if lo < 0 {
		panic("passit: lo argument must be positive")
	}

	if lo > maxR16 {
		tab.R32 = append(tab.R32, unicode.Range32{
			Lo:     uint32(lo),
			Hi:     uint32(hi),
			Stride: 1,
		})
		return
	}

	if hi > maxR16 {
		tab.R32 = append(tab.R32, unicode.Range32{
			Lo:     maxR16 + 1,
			Hi:     uint32(hi),
			Stride: 1,
		})
		hi = maxR16
	}

	tab.R16 = append(tab.R16, unicode.Range16{
		Lo:     uint16(lo),
		Hi:     uint16(hi),
		Stride: 1,
	})

	if hi <= unicode.MaxLatin1 {
		tab.LatinOffset++
	}
}

func intersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	var rt unicode.RangeTable

	/*if r0.Stride != 1 {
		panic("passit: unicode.RangeTable has entry with Stride > 1")
	}*/

	for _, r0 := range a.R16 {
		for _, r1 := range b.R16 {
			if r1.Lo > r0.Hi {
				break
			} else if r0.Lo > r1.Hi {
				continue
			}

			lo, hi, stride := intersection(rune(r0.Lo), rune(r0.Hi), rune(r1.Lo), rune(r1.Hi), rune(r1.Stride))
			if lo > hi {
				continue
			}

			if hi <= unicode.MaxLatin1 {
				rt.LatinOffset++
			}

			rt.R16 = append(rt.R16, unicode.Range16{Lo: uint16(lo), Hi: uint16(hi), Stride: uint16(stride)})
		}
	}

	for _, r0 := range a.R32 {
		for _, r1 := range b.R32 {
			if r1.Lo > r0.Hi {
				break
			} else if r0.Lo > r1.Hi {
				continue
			}

			lo, hi, stride := intersection(rune(r0.Lo), rune(r0.Hi), rune(r1.Lo), rune(r1.Hi), rune(r1.Stride))
			if lo > hi {
				continue
			}

			rt.R32 = append(rt.R32, unicode.Range32{Lo: uint32(lo), Hi: uint32(hi), Stride: uint32(stride)})
		}
	}

	return &rt
}

func intersection(lo0, hi0, lo1, hi1, stride1 rune) (lo, hi, stride rune) {
	lo, hi, stride = lo1, hi1, stride1

	if stride1 == 1 {
		if lo < lo0 {
			lo = lo0
		}
		if hi > hi0 {
			hi = hi0
		}
	} else {
		if lo < lo0 {
			c := lo0 - lo1
			c += stride1 - 1
			c -= c % stride1
			lo += c
		}
		if hi > hi0 {
			c := hi1 - hi0
			c += stride1 - 1
			c -= c % stride1
			hi -= c
		}
		if lo == hi {
			stride = 1
		}
	}

	return
}
