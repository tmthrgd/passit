package passit

import (
	"encoding/binary"
	"errors"
	"math/bits"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// modReader is used to probe the upper region of uint16 space. It generates values
// sequentially in [maxUint16-15,maxUint16]. With modEdge == 15 and
// maxUint16 == 1<<16-1 == 65535, this means that readUint16n(10) will
// repeatedly probe the top range. We thus expect a bias to result unless the
// calculation in readUint16n gets the edge condition right. We test this by calling
// readUint16n 100 times; the results should be perfectly evenly distributed across
// [0,10).
type modReader uint16

const modEdge = 15

func (m *modReader) Read(p []byte) (int, error) {
	if *m > modEdge {
		*m = 0
	}

	binary.LittleEndian.PutUint16(p, uint16(maxUint16-*m))
	*m++
	return 2, nil
}

func TestReadUint16n(t *testing.T) {
	// This test validates that the calculation in readUint16n corrects for
	// possible bias.
	//
	// This test and modReader was taken from golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/modulo_test.go.

	var (
		src    modReader
		result [10]int
	)
	for i := 0; i < 100; i++ {
		n, err := readUint16n(&src, 10)
		require.NoError(t, err)
		result[n]++
	}

	for _, r := range result {
		require.Equal(t, 10, r, result)
	}
}

func TestReadIntN(t *testing.T) {
	n, err := readIntN(iotest.ErrReader(errors.New("should not call Read")), 1)
	if assert.NoError(t, err, "readIntN with n=1 and error io.Reader: error") {
		assert.Equal(t, 0, n, "readIntN with n=1 and error io.Reader: result")
	}

	const maxInt = 1<<(bits.UintSize-1) - 1
	for _, n := range []int{-10, -1, 0, maxReadIntN + 1, maxInt} {
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
		{maxReadIntN, 19592},
		{maxReadIntN, 23034},
		{100, 14},
		{128, 43},
		{255, 59},
		{maxReadIntN, 52988},
	} {
		got, err := readIntN(tr, tc.N)
		if assert.NoErrorf(t, err, "readIntN with n=%d: error", tc.N) {
			assert.Equalf(t, tc.Expect, got, "readIntN with n=%d: result", tc.N)
		}
	}
}
