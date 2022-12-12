package passit

import (
	"encoding/binary"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/rangetable"
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

func TestCountTableRunes(t *testing.T) {
	for _, tabs := range []map[string]*unicode.RangeTable{
		unicode.Categories, unicode.Properties, unicode.Scripts,
	} {
		for _, tab := range tabs {
			got := countTableRunes(tab)

			var expect int
			rangetable.Visit(tab, func(rune) { expect++ })

			assert.Equal(t, expect, got)
		}
	}
}

type u16Reader uint16

func (u *u16Reader) Read(p []byte) (int, error) {
	binary.LittleEndian.PutUint16(p, uint16(*u))
	*u++
	return 2, nil
}

func TestReadRune(t *testing.T) {
	for _, tab := range []*unicode.RangeTable{
		rangeTableASCII,
		unicode.Lu,
		unicode.Ll,
		unicode.N,
	} {
		count := countTableRunes(tab)

		var ur u16Reader
		seen := make(map[rune]struct{}, count)
		for i := 0; i < count; i++ {
			r, err := readRune(&ur, tab, count)
			require.NoError(t, err)

			_, dup := seen[r]
			seen[r] = struct{}{}

			require.Truef(t, unicode.Is(tab, r), "rune %U not found in table", r)
			require.Falsef(t, dup, "duplicate rune %U returned", r)
		}

		r, err := readRune(&ur, tab, count)
		require.NoError(t, err)

		_, dup := seen[r]
		require.Truef(t, unicode.Is(tab, r), "rune %U not found in table", r)
		require.Truef(t, dup, "expected duplicate rune %U not returned", r)
	}
}
