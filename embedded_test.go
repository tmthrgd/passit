package passit

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func allWordsValid(t *testing.T, list []string) {
	t.Helper()

	for _, v := range list {
		assert.Truef(t, utf8.ValidString(v), "utf8.ValidString(%q)", v)
	}
}

// These tests are currently the same, but keep them separate in case that changes.
var allEmojiValid = allWordsValid

func TestEmbeddedWordlist(t *testing.T) {
	for _, tc := range []struct {
		name   string
		gen    Generator
		expect string
	}{
		{"OrchardStreetMedium", OrchardStreetMedium, "pavilion extinct stadium furnace shores pirates hospital influenced"},
		{"OrchardStreetLong", OrchardStreetLong, "agreed stopping brilliant elongated richness populous sprung grassland"},
		{"OrchardStreetAlpha", OrchardStreetAlpha, "bees told hymn pride boy scout hum bus"},
		{"OrchardStreetQWERTY", OrchardStreetQWERTY, "bids trio hurry queer buyer sect hull cadres"},
		{"STS10Wordlist", STS10Wordlist, "winner vertigo spurs believed dude runaways poorest tourists"},
		{"EFFLargeWordlist", EFFLargeWordlist, "reprint wool pantry unworried mummify veneering securely munchkin"},
		{"EFFShortWordlist1", EFFShortWordlist1, "bush vapor issue ruby carol sleep hula case"},
		{"EFFShortWordlist2", EFFShortWordlist2, "barracuda vegetable idly podiatrist bossiness satchel hexagon boxlike"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			const size = 8

			tr := newTestRand()
			pass, err := Repeat(tc.gen, " ", size).Password(tr)
			require.NoError(t, err)

			assert.Equal(t, tc.expect, pass)
			assert.Equal(t, size-1, strings.Count(pass, " "),
				`strings.Count(%q, " ")`, pass)
			assert.Truef(t, utf8.ValidString(pass),
				"utf8.ValidString(%q)", pass)

			allWordsValid(t, tc.gen.(*embeddedGenerator).list)
		})
	}
}

func TestEmoji(t *testing.T) {
	for _, tc := range []struct {
		name   string
		gen    Generator
		expect string
	}{
		{"Emoji13", Emoji13, "💙🏂🏽🧑🏽\u200d🦱🤙🏾🧗🏿\u200d♀️🧑🏻\u200d🤝\u200d🧑🏻🐭🚵🏿🚴🏿🏠💇🏾\u200d♂️💁🏻🙍🏾\u200d♂️👩🏾\u200d🦲🧑🏿\u200d🤝\u200d🧑🏼✨🖐🏿🎮🔑🏔️🔹🇩🇲💇🏼\u200d♀️🕶️🧙🏼🫕👳🏽\u200d♀️👩🏻\u200d💻👰🏿\u200d♀️🇲🇻🔃🖖🏻🧛🏿\u200d♂️👩🏼\u200d🤝\u200d👩🏻🧚🏿🇧🇦🇹🇻🇱🇹🍆🧑🏻\u200d🤝\u200d🧑🏾🌲👨🏼\u200d🦼🏌🏻\u200d♂️👨\u200d🚀😸👰🏽\u200d♀️🦖#️⃣👴🏼💂🏻🏌️\u200d♂️💲🐗🥇↘️👰🇨🇷👈🏽🦸🏿\u200d♂️🗺️🇭🇺🇯🇴🚣🏻👷🏽\u200d♂️🧖🏿🇬🇭🤙🏿🥾🤪⛹🏻👩🏻\u200d🌾☸️🧨▶️🐎🧑🏻\u200d🤝\u200d🧑🏾👨🏼\u200d🤝\u200d👨🏿🏆🕙🏆🏃🏻\u200d♂️🤿👨🏿\u200d⚕️🧑🏾🤛🏼🏋🏿🧑🏽\u200d🏭👮🏼\u200d♀️🙅🏾\u200d♂️Ⓜ️🧘🏾☑️⛴️🎙️🚭🦸🏻\u200d♂️🥷🏻📙👨🏾\u200d⚖️🍤"},
		{"Emoji15", Emoji15, "➡️🦸🏼\u200d♂️👩🏾\u200d🦳📱✍🏻🪣👨🏾\u200d🌾🤩🤵🏽\u200d♂️👮🏼🧗🏾\u200d♂️👷🏾\u200d♀️🧝🏾\u200d♂️👔🟨↗️🕵🏽👦🏽🏃🏽\u200d♂️🦶🤾🏿\u200d♂️⛺👮🏿👇🏽👳🏿🌀🦿👈🏽🏄🏽\u200d♀️🧑🏻\u200d🦰🔃🫣🏪🪿🧗🏽👃8️⃣👩🏿\u200d🦰🇹🇦👮🏼\u200d♂️👨🏼\u200d❤️\u200d👨🏿🧑🏿\u200d🦱🤸🏽\u200d♂️🛫👩🏻\u200d🦰👩🏽\u200d❤️\u200d💋\u200d👨🏾🦶🏾㊗️👩🏼\u200d🎤💁🏻\u200d♂️🧑🏿\u200d🤝\u200d🧑🏻🚶🏻\u200d♂️👨🏿\u200d⚖️🔟👨🏿\u200d🤝\u200d👨🏾👨\u200d👩\u200d👦🧝🏽\u200d♀️🔽🙋🏿\u200d♂️🧑🏼👩🏾\u200d🍼💆🏻\u200d♂️👩🏿\u200d🦽🐀💂🏻\u200d♀️🆑🍠🥸🤚🏾🚶🏻\u200d♂️🇦🇲💙👐💪🏾🫁👱🏻🧒🏿🧢🐖👨🏿\u200d✈️🦀🎅👨🏾\u200d❤️\u200d💋\u200d👨🏼👨🏼\u200d🦱🎫🥻🙆🏿\u200d♂️👩🏼\u200d❤️\u200d💋\u200d👨🏼👴🏿💪🏻💂🏿\u200d♂️🛌🌚🏈👩🏻\u200d🤝\u200d👨🏽🛀🏾👋🏾🧑🏼\u200d🤝\u200d🧑🏻🏫✋🏼"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			const size = 100

			tr := newTestRand()
			pass, err := Repeat(tc.gen, "", size).Password(tr)
			require.NoError(t, err)

			assert.Equal(t, tc.expect, pass)
			assert.Equal(t, size, countEmojiInString(tc.gen.(*embeddedGenerator).list, pass),
				"countEmojiInString(%q)", pass)
			assert.Truef(t, utf8.ValidString(pass),
				"utf8.ValidString(%q)", pass)

			allEmojiValid(t, tc.gen.(*embeddedGenerator).list)
		})
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
