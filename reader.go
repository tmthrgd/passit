package passit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode"
)

const maxUint16 = (1 << 16) - 1

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

func readUint16(r io.Reader) (uint16, error) {
	var buf [2]byte
	_, err := readBytes(r, buf[:])
	return binary.LittleEndian.Uint16(buf[:]), err
}

// readUint16n is a helper function that should only be called by readIntN.
func readUint16n(r io.Reader, n uint16) (uint16, error) {
	// This is based on golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/rand.go#L91-L109.

	v, err := readUint16(r)
	if err != nil {
		return 0, err
	}

	if n&(n-1) == 0 { // n is power of two, can mask
		return v & (n - 1), nil
	}

	// If n does not divide v, to avoid bias we must not use
	// a v that is within maxUint16%n of the top of the range.
	if v > maxUint16-n { // Fast check.
		ceiling := maxUint16 - maxUint16%n
		for v >= ceiling {
			v, err = readUint16(r)
			if err != nil {
				return 0, err
			}
		}
	}

	return v % n, nil
}

const maxReadIntN = maxUint16

func readIntN(r io.Reader, n int) (int, error) {
	switch {
	case n <= 0:
		panic("passit: invalid argument to readIntN")
	case n == 1:
		// If n is 1, meaning the result will always be 0, avoid reading
		// anything from r and immediately return 0.
		return 0, nil
	case n <= maxUint16:
		v, err := readUint16n(r, uint16(n))
		return int(v), err
	default:
		panic("passit: invalid argument to readIntN")
	}
}

func readSliceN[T any](r io.Reader, s []T) (T, error) {
	i, err := readIntN(r, len(s))
	if err != nil {
		return *new(T), err
	}

	return s[i], nil
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
		return 0, errors.New("passit: unicode.RangeTable is too large")
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
