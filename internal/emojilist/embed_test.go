package emojilist

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmojiCounts(t *testing.T) {
	// Expected count is taken from https://www.unicode.org/emoji/charts-M.N/emoji-counts.html.
	assert.Equal(t, 3304, strings.Count(Unicode13, "\n")+1, "Unicode 13.0")
}
