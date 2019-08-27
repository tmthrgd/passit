package password

import (
	"encoding/binary"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/rangetable"
)

// modReader is used to probe the upper region of uint32 space. It generates values
// sequentially in [maxUint32-15,maxUint32]. With modEdge == 15 and
// maxUint32 == 1<<32-1 == 4294967295, this means that readUint32n(10) will
// repeatedly probe the top range. We thus expect a bias to result unless the
// calculation in readUint32n gets the edge condition right. We test this by calling
// readUint32n 100 times; the results should be perfectly evenly distributed across
// [0,10).
type modReader uint32

const modEdge = 15

func (m *modReader) Read(p []byte) (int, error) {
	if *m > modEdge {
		*m = 0
	}

	binary.LittleEndian.PutUint32(p, uint32(maxUint32-*m))
	*m++
	return 4, nil
}

func TestReadUint64n(t *testing.T) {
	// This test validates that the calculation in readUint32n corrects for
	// possible bias.
	//
	// This test and modReader was taken from golang.org/x/exp/rand:
	// https://github.com/golang/exp/blob/ec7cb31e5a562f5e9e31b300128d2f530f55d127/rand/modulo_test.go.

	var (
		src    modReader
		result [10]int
	)
	for i := 0; i < 100; i++ {
		n, err := readUint32n(&src, 10)
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

type u32Reader uint32

func (u *u32Reader) Read(p []byte) (int, error) {
	binary.LittleEndian.PutUint32(p, uint32(*u))
	*u++
	return 4, nil
}

func TestReadRune(t *testing.T) {
	for _, tab := range []*unicode.RangeTable{
		rangeTableASCII,
		unicode.Lu,
		unicode.Ll,
		unicode.N,
	} {
		count := countTableRunes(tab)

		var ur u32Reader
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
