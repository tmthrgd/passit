package passit

import (
	"bufio"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

// modReader is used to probe the upper region of uint16 space. It generates values
// sequentially in [maxUint16-15,maxUint16]. With modEdge == 15 and
// maxUint16 == 1<<16-1 == 65535, this means that readUint16n(10) will
// repeatedly probe the top range. We thus expect a bias to result unless the
// calculation in readUint16n gets the edge condition right. We test this by calling
// readUint16n 100 times; the results should be perfectly evenly distributed across
// [0,10).

const modEdge = 15

func modEdgeInc[T constraints.Unsigned](m *T) T {
	if *m > modEdge {
		*m = 0
	}

	v := *m
	*m++
	return v
}

func putUintN[T constraints.Unsigned](b []byte, n int, v T) {
	// Silence a vet warning ("v (may be 8 bits) too small for shift of 8") by
	// always casting v to uint64.
	v64 := uint64(v)

	_ = b[n-1] // early bounds check to guarantee safety of writes below
	for i := 0; i < n; i++ {
		b[i] = byte(v64)
		v64 >>= 8
	}
}

type modReader8 uint8

func (m *modReader8) Read(p []byte) (int, error) {
	putUintN(p, 1, (1<<8-1)-modEdgeInc(m))
	return 1, nil
}

type modReader16 uint16

func (m *modReader16) Read(p []byte) (int, error) {
	putUintN(p, 2, (1<<16-1)-modEdgeInc(m))
	return 2, nil
}

type modReader24 uint32

func (m *modReader24) Read(p []byte) (int, error) {
	putUintN(p, 3, (1<<24-1)-modEdgeInc(m))
	return 3, nil
}

type modReader32 uint32

func (m *modReader32) Read(p []byte) (int, error) {
	putUintN(p, 4, (1<<32-1)-modEdgeInc(m))
	return 4, nil
}

type modReader40 uint64

func (m *modReader40) Read(p []byte) (int, error) {
	putUintN(p, 5, (1<<40-1)-modEdgeInc(m))
	return 5, nil
}

type modReader48 uint64

func (m *modReader48) Read(p []byte) (int, error) {
	putUintN(p, 6, (1<<48-1)-modEdgeInc(m))
	return 6, nil
}

type modReader56 uint64

func (m *modReader56) Read(p []byte) (int, error) {
	putUintN(p, 7, (1<<56-1)-modEdgeInc(m))
	return 7, nil
}

type modReader64 uint64

func (m *modReader64) Read(p []byte) (int, error) {
	putUintN(p, 8, (1<<64-1)-modEdgeInc(m))
	return 8, nil
}

func TestReadUint64nBias(t *testing.T) {
	// This test validates that the calculation in readUintN corrects for
	// possible bias.
	//
	// This test and modReader was taken from golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/modulo_test.go.

	for _, tc := range []struct {
		src     io.Reader
		bitLen  int
		variant string
	}{
		{new(modReader8), 8, "Read"},
		{new(modReader16), 16, "Read"},
		{new(modReader24), 24, "Read"},
		{new(modReader32), 32, "Read"},
		{new(modReader40), 40, "Read"},
		{new(modReader48), 48, "Read"},
		{new(modReader56), 56, "Read"},
		{new(modReader64), 64, "Read"},
		{bufio.NewReader(new(modReader8)), 8, "ReadByte"},
		{bufio.NewReader(new(modReader16)), 16, "ReadByte"},
		{bufio.NewReader(new(modReader24)), 24, "ReadByte"},
		{bufio.NewReader(new(modReader32)), 32, "ReadByte"},
		{bufio.NewReader(new(modReader40)), 40, "ReadByte"},
		{bufio.NewReader(new(modReader48)), 48, "ReadByte"},
		{bufio.NewReader(new(modReader56)), 56, "ReadByte"},
		{bufio.NewReader(new(modReader64)), 64, "ReadByte"},
	} {
		tc := tc
		t.Run(fmt.Sprintf("uint%d_%s", tc.bitLen, tc.variant), func(t *testing.T) {
			var result [10]int
			for i := 0; i < 500; i++ {
				n, _ := readUint64n(tc.src, 10, tc.bitLen)
				result[n]++
			}
			for _, r := range result {
				require.Equal(t, 50, r, "bias")
			}
		})
	}
}

func TestReadIntN(t *testing.T) {
	n, err := readIntN(errTestReader(), 1)
	if assert.NoError(t, err, "readIntN with n=1 and error io.Reader: error") {
		assert.Equal(t, 0, n, "readIntN with n=1 and error io.Reader: result")
	}

	for _, n := range []int{-10, -1, 0} {
		assert.PanicsWithValuef(t, "passit: invalid argument to readIntN", func() {
			readIntN(zeroReader{}, n)
		}, "readIntN with invalid n=%d", n)
	}

	trR := struct{ io.Reader }{newTestRand()}
	trBR := struct {
		io.Reader
		io.ByteReader
	}{
		errTestReader(),
		newTestRand().(io.ByteReader),
	}

	for _, tc := range []struct{ N, Expect int }{
		{1, 0},
		{10, 2},
		{100, 75},
		{128, 84},
		{255, 239},
		{1<<16 - 1, 11402},
		{1<<16 - 1, 34875},
		{100, 76},
		{128, 122},
		{1, 0},
		{1<<8 - 1, 89},
		{1<<16 - 1, 13514},
		{1<<24 - 1, 5778987},
		{1<<32 - 1, 4207869154},
		{1<<48 - 1, 32432210391166},
		{1<<63 - 1, 6523467746500912216},
		{1, 0},
		{1<<8 - 1, 206},
		{1<<16 - 1, 46688},
		{1<<24 - 1, 15962787},
		{1<<32 - 1, 1907999272},
		{1<<48 - 1, 187561078750898},
		{1<<63 - 1, 720003246643169708},
		{1, 0},
		{1<<8 - 1, 148},
		{1<<16 - 1, 49547},
		{1<<24 - 1, 139488},
		{1<<32 - 1, 1934500113},
		{1<<48 - 1, 189899984657044},
		{1<<63 - 1, 1847124326628430800},
	} {
		got, err := readIntN(trR, tc.N)
		if assert.NoErrorf(t, err, "readIntN(io.Reader, %d): error", tc.N) {
			assert.Equalf(t, tc.Expect, got, "readIntN(io.Reader, %d): result", tc.N)
		}

		got, err = readIntN(trBR, tc.N)
		if assert.NoErrorf(t, err, "readIntN(io.ByteReader, %d): error", tc.N) {
			assert.Equalf(t, tc.Expect, got, "readIntN(io.ByteReader, %d): result", tc.N)
		}
	}
}
