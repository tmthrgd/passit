package password

import (
	"math/rand"
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
		{"ṥƩǶⅨᴕȋʡḲⅫỚżƢɔⅴȮẜꝟᶍＫꟅꞡɸɊꞹƙ", unicode.Latin},
		{"ὦ𐆋𐅼ᾡ𝈉𝈓𐆇ᾶ𐅨Ὺ𝈶ἠῑϸϽῪϸϘΏ𐅵𐅡𝈾ῆϊβ", unicode.Greek},
		{"₥꠸৲߿₰₦฿₲₽₶₫₹₧₮₵₠₠₸₼₼₢₧﷼₨₮", unicode.Sc},
	}

	if unicode.Version == "12.0.0" {
		testCases = append(testCases, testCasesUni...)
	} else {
		t.Logf("skipping %d test cases without Unicode 12.0.0", len(testCasesUni))
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
