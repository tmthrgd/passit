package passit

import (
	"math/rand"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/rangetable"
)

func TestCharset(t *testing.T) {
	for _, tc := range []struct{ expect, template string }{
		{"1010000010111010000001100", "01"},
		{"1690822236719012868805980", "0123456789"},
		{"lwrqmcesfypbvqzagueycldeq", "abcdefghijklmnopqrstuvwxyz"},
		{"LWRQMCESFYPBVQZAGUEYCLDEQ", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"lWRQmCESfYpBVqZAGuEYcLDeq", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"},
		{"VwJIgWOitmbpXelQQossSXxYe", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"},
		{"lj$#+pr%%yc!iqmathelp_dr#", "abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+"},
		{"ρΘΙδΚΙσΧεΣσΕΒωπΠΟΠδΑΕλΚΥΦ", "ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω"},
		{"🍧🛰🍳🔱🚱👒🎩👒🍉🌴💻🍧🍳🐊🍧🎩🚱🛰💅💅🔱👗🚋🚱🐊", "🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋"},
	} {
		const size = 25

		testRand := rand.New(rand.NewSource(0))

		tmpl, err := FromCharset(tc.template)
		if !assert.NoError(t, err) {
			continue
		}

		pass, err := tmpl(size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, pass)
	}
}

func TestFixedCharset(t *testing.T) {
	for _, tc := range []struct {
		expect   string
		template func(int) Template
	}{
		{"lwrqmcesfypbvqzagueycldeq", LatinLower},
		{"LWRQMCESFYPBVQZAGUEYCLDEQ", LatinUpper},
		{"LwrqMcesFyPbvQzagUeyCldEQ", LatinMixed},
		{"1690822236719012868805980", Number},
	} {
		const size = 25

		testRand := rand.New(rand.NewSource(0))

		pass, err := tc.template(size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, pass)
	}
}

func TestRangeTable(t *testing.T) {
	newTable := func(s string) *unicode.RangeTable {
		return rangetable.New([]rune(s)...)
	}

	type testCase struct {
		expect string
		tab    *unicode.RangeTable
	}
	testCases := []testCase{
		{"1010000010111010000001100", newTable("01")},
		{"1690822236719012868805980", newTable("0123456789")},
		{"lwrqmcesfypbvqzagueycldeq", newTable("abcdefghijklmnopqrstuvwxyz")},
		{"LWRQMCESFYPBVQZAGUEYCLDEQ", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ")},
		{"lWRQmCESfYpBVqZAGuEYcLDeq", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")},
		{"Lm98WMEYjcRfNUbGGeiiINnOU", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")},
		{"_@sr~dfttm$p+ea!h*&_dz%fr", newTable("abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+")},
		{"κΟΡΘΤΡμτΚλμΙΓωθηεηΘΑΙΧΤπς", newTable("ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω")},
		{"🍈🐂👒🌴🚱🍉💅🍉🍧🔱🚋🍈👒👗🍈💅🚱🐂🐝🐝🌴💻🛰🚱👗", newTable("🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋")},
		{"7032aEC2b213F2f2eaCecdFc4", unicode.ASCII_Hex_Digit},
	}
	testCasesUni := []testCase{
		{"ƵỖẝȶỶẂᵽɥếⅮḭƒᴥʆṞṕÓꭔǤȪｌⱶꜷṂﬄ", unicode.Latin},
		{"ὦ𐆋𐅼ᾡ𝈉𝈓𐆇ᾶ𐅨Ὺ𝈶ἠῑϸϽῪϸϘΏ𐅵𐅡𝈾ῆϊβ", unicode.Greek},
		{"₥꠸৲߿₰₦฿₲₽₶₫₹₧₮₵₠₠₸₼₼₢₧﷼₨₮", unicode.Sc},
	}

	if unicode.Version == unicodeVersion {
		testCases = append(testCases, testCasesUni...)
	} else {
		t.Logf("skipping %d test cases due to mismatched unicode versions; have %s, want %s", len(testCasesUni), unicode.Version, unicodeVersion)
	}

	for _, tc := range testCases {
		const size = 25

		testRand := rand.New(rand.NewSource(0))

		tmpl, err := FromRangeTable(tc.tab)
		if !assert.NoError(t, err) {
			continue
		}

		pass, err := tmpl(size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, pass)
	}
}

func TestEmoji13(t *testing.T) {
	const size = 25

	for i, expect := range []string{
		"👷\u200d♀️🕸️🏃\u200d♀️🐎🚵🏽\u200d♂️👩🏿\u200d🏫🙋🏻💖👐🏽🧷🔇🌛🙍🏿\u200d♀️👨🏿\u200d🎓🕵🏼👴🧗🏽\u200d♀️💺🇹🇯🔎👳🏽\u200d♂️🤞🏼👩🏻\u200d🦰📦🎂",
		"🏮🤌🏽🧑🏽\u200d🤝\u200d🧑🏾👩🏼\u200d🍳🤽🏽🤳🏃🏾\u200d♂️❕🐣🆚🔧👏🏽🏄🏽💇🏼🥾🤟🏼👨\u200d🚀🦶🏻🧚🏻🛌🏻🚨💒😍🇵🇼🙎🏻",
		"😪🗨️📍☃️🏄🏼\u200d♂️🌑👩🏼\u200d🚒👷🏽\u200d♂️🧙🏾\u200d♀️👌🤹🏿\u200d♀️🈳🧑🏿\u200d🍳🏝️🇷🇸🧑🏼\u200d🎓🧑🏽\u200d⚕️🦻🏽👩\u200d🍼🧑🏿\u200d🏫🇸🇸👲⏺️☺️🦹\u200d♂️",
		"🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f🙆🏽🫑🧘🏽🚄🇸🇧🚶🏾\u200d♀️🤚🏻🦹\u200d♀️👩🏼\u200d🦼👎😵🤷🇻🇬👩🏿\u200d🚀🏊🏻\u200d♀️🙋🏻\u200d♂️👨🏽\u200d🦲🙂👩🏾\u200d🤝\u200d👩🏼🧍🏾\u200d♂️🧖🏾\u200d♀️👩🦸🏽🧜🏽\u200d♀️",
	} {
		testRand := rand.New(rand.NewSource(int64(i)))

		pass, err := Emoji13(size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, expect, pass)
		assert.Equal(t, size, countEmojiInString(pass),
			"countEmojiInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
	}
}

func TestEmojiValid(t *testing.T) {
	for _, emoji := range unicodeEmoji {
		assert.Truef(t, utf8.ValidString(emoji),
			"utf8.ValidString(%q)", emoji)
	}
}

func countEmojiInString(s string) int {
	var count int
outer:
	for len(s) > 0 {
		for i := len(unicodeEmoji) - 1; i >= 0; i-- {
			emoji := unicodeEmoji[i]
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
