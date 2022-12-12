package passit

import (
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEFFLargeWordlist(t *testing.T) {
	const size = 8

	testRand := rand.New(rand.NewSource(0))

	pass, err := Repeat(EFFLargeWordlist, " ", size).Password(testRand)
	require.NoError(t, err)

	assert.Equal(t, "timothy hubcap partner frigidly usage probiotic yodel playback", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(EFFLargeWordlist.(*embeddedList).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEFFShortWordlist1(t *testing.T) {
	const size = 8

	testRand := rand.New(rand.NewSource(0))

	pass, err := Repeat(EFFShortWordlist1, " ", size).Password(testRand)
	require.NoError(t, err)

	assert.Equal(t, "dice marry lake cash shade sift woozy plank", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(EFFShortWordlist1.(*embeddedList).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEFFShortWordlist2(t *testing.T) {
	const size = 8

	testRand := rand.New(rand.NewSource(0))

	pass, err := Repeat(EFFShortWordlist2, " ", size).Password(testRand)
	require.NoError(t, err)

	assert.Equal(t, "downstairs lakefront jersey breath reimburse robe wreckage narrator", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(EFFShortWordlist2.(*embeddedList).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEmoji13(t *testing.T) {
	const size = 25

	for i, expect := range []string{
		"🇳🇷👀👨🏻\u200d🦰🧚🏻\u200d♀️✴️🇵🇸🍣👫🏼📚🇪🇷👩\u200d🍼🇼🇫🩺🧝🏼\u200d♂️👆🏾🎧🧏🏿\u200d♀️🕴🏽🧜🏿\u200d♀️🖋️🎁🇬🇭🪙🇨🇺⏺️",
		"🤷🏻🏥📑🚣🏻\u200d♂️🕣🧑🏻👮🏿\u200d♀️🧘🏽\u200d♂️🌁🚻♈👧🏻👴🏽🙆🏾👨🏻\u200d🔬⏏️👩🏽\u200d🦽🥭🅿️👝⚾✋🏾👩🏻\u200d🎨🏄🏽🤍",
		"🏹👩🏾\u200d🦰🦹🏻🥿🚣🅾️🧑🏽\u200d🤝\u200d🧑🏾🍐🔂🧝🏽\u200d♂️💂🏿\u200d♂️🧑🏽\u200d🚀👴🏿👩🏾\u200d🏭👩🏾\u200d✈️🧿🐒💁🏿\u200d♂️🌴🧑\u200d🍼🚴🏽🇦🇪🦘🥓🇱🇧",
		"🕵️🧑🏻\u200d🎄🔎🧑🏼\u200d💻🧕🏼🤤🛸🈶☝🏿👨🏼\u200d⚖️👊🏾🏊🏻\u200d♂️🧎🧍\u200d♂️🗓️🏃🏾\u200d♀️🗽🐬👩🏾\u200d💼🎸☕🚜🤾🏼\u200d♂️🐯🚴🏽",
	} {
		testRand := rand.New(rand.NewSource(int64(i)))

		pass, err := Repeat(Emoji13, "", size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, expect, pass)
		assert.Equal(t, size, countEmojiInString(Emoji13.(*embeddedList).list, pass),
			"countEmojiInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}

func TestEmojiValid(t *testing.T) {
	for _, emoji := range strings.Split(emoji13List, "\n") {
		assert.Truef(t, utf8.ValidString(emoji),
			"utf8.ValidString(%q)", emoji)
	}
}

func TestEmojiCounts(t *testing.T) {
	// Expected count is taken from https://www.unicode.org/emoji/charts-M.N/emoji-counts.html.
	assert.Equal(t, 3304, strings.Count(emoji13List, "\n")+1, "Unicode 13.0")
}

func countEmojiInString(list []string, s string) int {
	var count int
outer:
	for len(s) > 0 {
		for i := len(list) - 1; i >= 0; i-- {
			emoji := list[i]
			if strings.HasPrefix(s, emoji) {
				count++
				s = s[len(emoji):]
				continue outer
			}
		}

		return -1
	}

	return count
}
