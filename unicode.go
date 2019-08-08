package password

import (
	"sync"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

var rangeTableASCII = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x20, Hi: 0x7e, Stride: 1},
	},
	LatinOffset: 1,
}

// TODO(tmthrgd): Review these ranges. PrintRanges is likely too permissive.
var allowedRanges = append(unicode.PrintRanges, rangeTableASCII)

func notAllowed(r rune) bool {
	if r <= 0x7e { // Fast path for ASCII.
		return r < 0x20
	}

	return !unicode.In(r, allowedRanges...)
}

var allowedRangeTableVal struct {
	tab *unicode.RangeTable
	sync.Once
}

func allowedRangeTable() *unicode.RangeTable {
	allowedRangeTableVal.Do(func() {
		allowedRangeTableVal.tab = rangetable.Merge(allowedRanges...)
	})
	return allowedRangeTableVal.tab
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
