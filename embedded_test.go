package passit

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSTS10Wordlist(t *testing.T) {
	const size = 8

	tr := newTestRand()

	pass, err := Repeat(STS10Wordlist, " ", size).Password(tr)
	require.NoError(t, err)

	assert.Equal(t, "winner vertigo spurs believed dude runaways poorest tourists", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(STS10Wordlist.(*embeddedGenerator).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEFFLargeWordlist(t *testing.T) {
	const size = 8

	tr := newTestRand()

	pass, err := Repeat(EFFLargeWordlist, " ", size).Password(tr)
	require.NoError(t, err)

	assert.Equal(t, "reprint wool pantry unworried mummify veneering securely munchkin", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(EFFLargeWordlist.(*embeddedGenerator).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEFFShortWordlist1(t *testing.T) {
	const size = 8

	tr := newTestRand()

	pass, err := Repeat(EFFShortWordlist1, " ", size).Password(tr)
	require.NoError(t, err)

	assert.Equal(t, "bush vapor issue ruby carol sleep hula case", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(EFFShortWordlist1.(*embeddedGenerator).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEFFShortWordlist2(t *testing.T) {
	const size = 8

	tr := newTestRand()

	pass, err := Repeat(EFFShortWordlist2, " ", size).Password(tr)
	require.NoError(t, err)

	assert.Equal(t, "barracuda vegetable idly podiatrist bossiness satchel hexagon boxlike", pass)
	assert.Equal(t, size-1, strings.Count(pass, " "),
		`strings.Count(%q, " ")`, pass)
	assert.Truef(t, utf8.ValidString(pass),
		"utf8.ValidString(%q)", pass)

	_, err = FromWords(EFFShortWordlist2.(*embeddedGenerator).list...)
	assert.NoError(t, err, "wordlist valid")
}

func TestEmoji13(t *testing.T) {
	const size = 25

	tr := newTestRand()

	for _, expect := range []string{
		"🏪🇮🇶👩🏾\u200d🎤🚣🏾🧍🏿\u200d♂️👩\u200d👩\u200d👦\u200d👦🐄🙋🏾👰🏽🌲👩🏿\u200d⚕️👩🏾💁🏾\u200d♂️👩🏻\u200d🦲🧘🏽\u200d♀️🧑🏽\u200d🤝\u200d🧑🏽👱🏻🌄💌⛏️🔙🎟️🏋🏾\u200d♂️4️⃣🤷🏾",
		"🦸👨🏾\u200d⚖️👨🏿\u200d🍼🏃🏻\u200d♀️🛰️📼💪🏾🧏🏼\u200d♂️🧏🏿\u200d♂️🤾🏻✋🏾🇰🇳🗒️🌃👩🏾\u200d🤝\u200d👨🏼⚓🤵\u200d♀️🧑🏾\u200d🔬🤽🏼🔏🧑🏽\u200d🏫🛫↙️🇾🇪👫🏾",
		"🧑🏼\u200d🦼🐶🏴🚵🦻👙🈂️🏊🏼🦸🏻\u200d♀️⚗️🏞️🇨🇽💆🏿👨🏼\u200d⚕️🤘🏾🕊️🙏🏻🥸😴⛏️🧗\u200d♀️⏲️🥱🩳🍄",
		"👨🏿\u200d🤝\u200d👨🏽🧑🏻\u200d⚖️🍱📡🍄👩🏿\u200d🎨🚥🧑🏻\u200d🦳🤳🏾💅🏽🏂🏽👩🏾\u200d🎤🧑🏼\u200d🦱👨🏾\u200d✈️🩰🤚🏽⏱️☦️☯️😃🙍🏽\u200d♂️🤌🏽📂🧑🏼\u200d🎓🌘",
	} {
		pass, err := Repeat(Emoji13, "", size).Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, expect, pass)
		assert.Equal(t, size, countEmojiInString(Emoji13.(*embeddedGenerator).list, pass),
			"countEmojiInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}

	// Check that each emoji in Emoji13 is a valid UTF8 string.
	for _, emoji := range Emoji13.(*embeddedGenerator).list {
		assert.Truef(t, utf8.ValidString(emoji),
			"utf8.ValidString(%q)", emoji)
	}
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
