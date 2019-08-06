package password

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/rangetable"
)

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
