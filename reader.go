package passit

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode"
)

const (
	maxUint32 = (1 << 32) - 1
	maxInt32  = (1 << 31) - 1
)

func readBytes(r io.Reader, buf []byte) (int, error) {
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return n, fmt.Errorf("passit: failed to read entropy: %w", err)
	}
	return n, nil
}

func readUint8(r io.Reader) (uint8, error) {
	var buf [1]byte
	_, err := readBytes(r, buf[:])
	return buf[0], err
}

func readUint32(r io.Reader) (uint32, error) {
	var buf [4]byte
	_, err := readBytes(r, buf[:])
	return binary.LittleEndian.Uint32(buf[:]), err
}

// readUint32n is a helper function that should only be called by readIntN.
func readUint32n(r io.Reader, n uint32) (uint32, error) {
	// This is based on golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/rand.go#L91-L109.

	v, err := readUint32(r)
	if err != nil {
		return 0, err
	}

	if n&(n-1) == 0 { // n is power of two, can mask
		return v & (n - 1), nil
	}

	// If n does not divide v, to avoid bias we must not use
	// a v that is within maxUint32%n of the top of the range.
	if v > maxUint32-n { // Fast check.
		ceiling := maxUint32 - maxUint32%n
		for v >= ceiling {
			v, err = readUint32(r)
			if err != nil {
				return 0, err
			}
		}
	}

	return v % n, nil
}

const maxReadIntN = maxInt32

func readIntN(r io.Reader, n int) (int, error) {
	switch {
	case n <= 0:
		panic("passit: invalid argument to readIntN")
	case n == 1:
		// If n is 1, meaning the result will always be 0, avoid reading
		// anything from r and immediately return 0.
		return 0, nil
	case n <= maxInt32:
		v, err := readUint32n(r, uint32(n))
		return int(v), err
	default:
		panic("passit: invalid argument to readIntN")
	}
}

func countTableRunes(tab *unicode.RangeTable) int {
	var c int
	for _, r16 := range tab.R16 {
		c += int((r16.Hi-r16.Lo)/r16.Stride) + 1
	}
	for _, r32 := range tab.R32 {
		c += int((r32.Hi-r32.Lo)/r32.Stride) + 1
	}

	return c
}

func readRune(r io.Reader, tab *unicode.RangeTable, count int) (rune, error) {
	if count > maxReadIntN {
		panic("passit: unicode.RangeTable is too large")
	}

	v, err := readIntN(r, count)
	if err != nil {
		return 0, err
	}

	for _, r16 := range tab.R16 {
		size := int((r16.Hi-r16.Lo)/r16.Stride) + 1
		if v < size {
			return rune(r16.Lo + uint16(v)*r16.Stride), nil
		}
		v -= size
	}

	for _, r32 := range tab.R32 {
		size := int((r32.Hi-r32.Lo)/r32.Stride) + 1
		if v < size {
			return rune(r32.Lo + uint32(v)*r32.Stride), nil
		}
		v -= size
	}

	panic("passit: internal error: unicode.RangeTable did not contain rune")
}
