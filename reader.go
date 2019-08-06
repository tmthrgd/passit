package password

import (
	"encoding/binary"
	"io"
)

const maxUint32 = (1 << 32) - 1

func readUint32(r io.Reader) (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	return binary.LittleEndian.Uint32(buf[:]), err
}

func readUint32n(r io.Reader, n uint32) (uint32, error) {
	// This was modled on golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/rand.go#L91-L109.

	v, err := readUint32(r)
	if err != nil {
		return 0, err
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
