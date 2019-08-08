package password

import (
	"unicode"
)

//go:generate go run unicode_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 11.0.0

func notAllowed(r rune) bool {
	if r <= 0x7e { // Fast path for ASCII.
		return r < 0x20
	}

	return !unicode.Is(allowedRangeTable, r)
}

func intersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	var rt unicode.RangeTable

	/*if r0.Stride != 1 {
		panic("strongroom/password: unicode.RangeTable has entry with Stride > 1")
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
