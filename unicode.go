package password

import (
	"io"
	"os"
	"strings"
	"unicode"

	"go.tmthrgd.dev/strongroom/internal"
)

//go:generate go run unicode_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 12.0.0

func notAllowed(r rune) bool {
	if r <= 0x7e { // Fast path for ASCII.
		return r < 0x20
	}

	return !unicode.Is(allowedRangeTable, r)
}

func isTestBinary() bool {
	// This is an approach that was used previously by the standard library in
	// net/http/roundtrip_js.go. See useFakeNetwork in:
	// https://github.com/golang/go/blob/220552f662%5E/src/net/http/roundtrip_js.go#L185-L189.
	return len(os.Args) > 0 && strings.HasSuffix(os.Args[0], ".test")
}

func maybeUnicodeReadByte(r io.Reader) {
	// TODO(tmthrgd): Remove once allowedRangeTable has stabalized.

	if !isTestBinary() {
		internal.MaybeReadByte(r)
	}
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
