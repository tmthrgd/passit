package passit

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	maxUint16   = (1 << 16) - 1
	maxReadIntN = maxUint16
)

func wrapReadError(err error) error {
	return fmt.Errorf("passit: failed to read entropy: %w", err)
}

func readUint16Buffer(r io.Reader) (uint16, error) {
	// buf escapes into io.Reader and is thus heap allocated.
	var buf [2]byte

	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, wrapReadError(err)
	}

	return binary.LittleEndian.Uint16(buf[:]), nil
}

func readUint16Byte(br io.ByteReader) (uint16, error) {
	b0, err := br.ReadByte()
	if err != nil {
		return 0, wrapReadError(err)
	}

	b1, err := br.ReadByte()
	if err != nil {
		return 0, wrapReadError(err)
	}

	// Assemble a little-endian uint16.
	return uint16(b0) | uint16(b1)<<8, nil
}

// readUint16n is a helper function that should only be called by readIntN.
func readUint16n(r io.Reader, n uint16) (v uint16, err error) {
	// This is based on golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/rand.go#L91-L109.

	br, brOK := r.(io.ByteReader)
	if brOK {
		v, err = readUint16Byte(br)
	} else {
		v, err = readUint16Buffer(r)
	}
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
			if brOK {
				v, err = readUint16Byte(br)
			} else {
				v, err = readUint16Buffer(r)
			}
			if err != nil {
				return 0, err
			}
		}
	}

	return v % n, nil
}

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
