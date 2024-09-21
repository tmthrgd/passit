package passit

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
)

func wrapReadError(err error) error {
	return fmt.Errorf("passit: failed to read entropy: %w", err)
}

func readUint64Buffer(r io.Reader, byteLen int) (uint64, error) {
	// buf escapes into io.Reader and is thus heap allocated.
	var buf [8]byte

	if _, err := io.ReadFull(r, buf[:byteLen]); err != nil {
		return 0, wrapReadError(err)
	}

	return binary.LittleEndian.Uint64(buf[:]), nil
}

func readUint64Byte(br io.ByteReader, byteLen int) (uint64, error) {
	var v, n uint64
	for range byteLen {
		b, err := br.ReadByte()
		if err != nil {
			return 0, wrapReadError(err)
		}

		v |= uint64(b) << n
		n += 8
	}

	return v, nil
}

// readUint64n is a helper function that should only be called by readIntN.
func readUint64n(r io.Reader, n uint64, bitLen int) (v uint64, err error) {
	// This is based on golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/rand.go#L91-L109.

	byteLen := bitLen / 8
	br, brOK := r.(io.ByteReader)
	if brOK {
		v, err = readUint64Byte(br, byteLen)
	} else {
		v, err = readUint64Buffer(r, byteLen)
	}
	if err != nil {
		return 0, err
	}

	if n&(n-1) == 0 { // n is power of two, can mask
		return v & (n - 1), nil
	}

	// max is the maximum value read from readUint64* and depends on the number
	// of bytes we read.
	//
	// Note: v <= max && n <= max.
	max := uint64(1)<<bitLen - 1

	// If n does not divide v, to avoid bias we must not use a v that is within
	// max%n of the top of the range.
	if v > max-n { // Fast check.
		ceiling := max - max%n
		for v >= ceiling {
			if brOK {
				v, err = readUint64Byte(br, byteLen)
			} else {
				v, err = readUint64Buffer(r, byteLen)
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
	}

	// Round up to the nearest multiple of 8 (i.e. 8, 16, 24, 32, 40, 48, 56 or 64).
	bitLen := (bits.Len(uint(n)) + 7) &^ 7

	v, err := readUint64n(r, uint64(n), bitLen)
	return int(v), err
}

func readSliceN[T any](r io.Reader, s []T) (T, error) {
	i, err := readIntN(r, len(s))
	if err != nil {
		return *new(T), err
	}

	return s[i], nil
}
