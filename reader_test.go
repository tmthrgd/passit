package passit

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

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
	_ = b[n-1] // early bounds check to guarantee safety of writes below
	for i := 0; i < n; i++ {
		b[i] = byte(v)
		v >>= 8
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
	n, err := readIntN(iotest.ErrReader(errors.New("should not call Read")), 1)
	if assert.NoError(t, err, "readIntN with n=1 and error io.Reader: error") {
		assert.Equal(t, 0, n, "readIntN with n=1 and error io.Reader: result")
	}

	for _, n := range []int{-10, -1, 0} {
		assert.PanicsWithValuef(t, "passit: invalid argument to readIntN", func() {
			readIntN(zeroReader{}, n)
		}, "readIntN with invalid n=%d", n)
	}

	tr := newTestRand()

	for _, tc := range []struct{ N, Expect int }{
		{10, 0},
		{100, 47},
		{128, 111},
		{255, 103},
		{1<<16 - 1, 19592},
		{1<<16 - 1, 23034},
		{100, 14},
		{128, 43},
		{1<<8 - 1, 59},
		{1<<16 - 1, 52988},
		{1<<24 - 1, 3178331},
		{1<<32 - 1, 1461550902},
		{1<<48 - 1, 149547980810148},
		{1<<63 - 1, 2950863412694929114},
		{1<<8 - 1, 124},
		{1<<16 - 1, 45681},
		{1<<24 - 1, 16218515},
		{1<<32 - 1, 1263119274},
		{1<<48 - 1, 280968136434521},
		{1<<63 - 1, 2382688018088692628},
		{1<<8 - 1, 193},
		{1<<16 - 1, 55956},
		{1<<24 - 1, 11962828},
		{1<<32 - 1, 3769340880},
		{1<<48 - 1, 156204095655369},
		{1<<63 - 1, 4884642788621909290},
	} {
		got, err := readIntN(tr, tc.N)
		if assert.NoErrorf(t, err, "readIntN with n=%d: error", tc.N) {
			assert.Equalf(t, tc.Expect, got, "readIntN with n=%d: result", tc.N)
		}
	}
}
