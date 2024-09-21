package emojilist

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmojiCounts(t *testing.T) {
	// Expected count is taken from https://www.unicode.org/Public/emoji/N.M/emoji-test.txt.
	assert.Equal(t, 3295, strings.Count(Unicode13, "\n")+1, "Unicode 13.0")
	assert.Equal(t, 3655, strings.Count(Unicode15, "\n")+1, "Unicode 15.0")
}
