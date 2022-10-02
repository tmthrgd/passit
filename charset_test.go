package passit

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

		pass, err := Repeat(tmpl, "", size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, tc.template, pass)
	}
}

func TestFixedCharset(t *testing.T) {
	for _, tc := range []struct {
		expect   string
		template Template
	}{
		{"lwrqmcesfypbvqzagueycldeq", LatinLower},
		{"LWRQMCESFYPBVQZAGUEYCLDEQ", LatinUpper},
		{"LwrqMcesFyPbvQzagUeyCldEQ", LatinMixed},
		{"1690822236719012868805980", Number},
	} {
		const size = 25

		testRand := rand.New(rand.NewSource(0))

		pass, err := Repeat(tc.template, "", size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, tc.template.(*asciiCharset).s, pass)
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

	const unicodeVersion = "13.0.0"
	testCasesUni := []testCase{
		{"ᴓĜPḄŚḸꝊẺꝉḂKṙṿƒćʢꝤᶧꜢＹĊʝḣꟇÔ", unicode.Latin},
		{"𝈡ΧὐΘῑ𝈂ἇώὰ𐅐ῆ᾿ΰΐ𐅛ᾇᾅυᾅ𐅦ͻ𝈱Άᾗΐ", unicode.Greek},
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

		pass, err := Repeat(FromRangeTable(tc.tab), "", size).Password(testRand)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, tc.tab, pass)
	}
}
