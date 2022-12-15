package passit

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/rangetable"
)

func TestCharset(t *testing.T) {
	for _, tc := range []struct{ expect, charset string }{
		{"", ""},
		{"~~~~~~~~~~~~~~~~~~~~~~~~~", "~"},
		{"0110000100000101100110101", "01"},
		{"0778244948606109948934141", "0123456789"},
		{"chzqoyupqagyotuvfssvfazip", "abcdefghijklmnopqrstuvwxyz"},
		{"CHZQOYUPQAGYOTUVFSSVFAZIP", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"CHzQoyuPQAGYOtUVfSsvfaZiP", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"},
		{"sjpUAg8nkoSSQrctlGCdFAxkN", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"},
		{"cu+qoyuc#~t_b^uv%%fv%nmip", "abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+"},
		{"κΔφδΦΓΥΖΞκκΤΠΔΖΠοΣυγψΝηΔΝ", "ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω"},
		{"🐊🐝🐝💅🎩🍈🍈🚋💬💅🛰🔱🛰🍧🔱🍳🚋🍈💅🚋🍉🍈🍧🍈🐂", "🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋"},
	} {
		tr := newTestRand()

		gen, err := FromCharset(tc.charset)
		if !assert.NoError(t, err) {
			continue
		}

		pass, err := Repeat(gen, "", 25).Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, tc.charset, pass)
	}
}

func TestFixedCharset(t *testing.T) {
	for _, tc := range []struct {
		expect string
		gen    Generator
	}{
		{"0778244948606109948934141", Digit},
		{"chzqoyupqagyotuvfssvfazip", LatinLower},
		{"CHZQOYUPQAGYOTUVFSSVFAZIP", LatinUpper},
		{"chZqOYUpqagyoTuvFsSVFAzIp", LatinMixed},
		{"0x92i4olu68ewri7h861h87op", LatinLowerDigit},
		{"0X92I4OLU68EWRI7H861H87OP", LatinUpperDigit},
		{"SJPuaG8NKOssqRCTLgcDfaXKn", LatinMixedDigit},
	} {
		const size = 25

		tr := newTestRand()

		pass, err := Repeat(tc.gen, "", size).Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Equal(t, size, utf8.RuneCountInString(pass),
			"utf8.RuneCountInString(%q)", pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, tc.gen.(*asciiGenerator).s, pass)
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
		{"", new(unicode.RangeTable)},
		{"~~~~~~~~~~~~~~~~~~~~~~~~~", newTable("~")},
		{"0110000100000101100110101", newTable("01")},
		{"0778244948606109948934141", newTable("0123456789")},
		{"chzqoyupqagyotuvfssvfazip", newTable("abcdefghijklmnopqrstuvwxyz")},
		{"CHZQOYUPQAGYOTUVFSSVFAZIP", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ")},
		{"CHzQoyuPQAGYOtUVfSsvfaZiP", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")},
		{"iZfK0WydaeIIGhSjb62T50naD", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")},
		{"$i~ecmi$rohz#uijtt(jtba+d", newTable("abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+")},
		{"ΥΗσΘςΕπΛγΥΥξηΗΛηζλρΖχαΞΗα", newTable("ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω")},
		{"👗🐊🐊🐝💅💬💬🛰🍳🐝🐂🌴🐂🍈🌴👒🛰💬🐝🛰🍧💬🍈💬👞", newTable("🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋")},
		{"e7FCC065cCCe6FA9F047BC9e1", unicode.ASCII_Hex_Digit},
	}

	const unicodeVersion = "13.0.0"
	testCasesUni := []testCase{
		{"ᵻḙꞪiǰↇꝢṛŸḨẨĠỤǉŦꝋɡḆＹɅẁṦǟḊꭒ", unicode.Latin},
		{"ἳ𝈛ῥᵡ𐅺όΫ𐅷𐆎ἓ𐅖Ί𝈂ΗᾁῈϼᴧρὺᵞ𐅰Ϟ𐅬θ", unicode.Greek},
		{"₼₳₹₤$₰𞋿₷₴₸₢₢₠₻€₽₵؋£₭֏$﷼₴௹", unicode.Sc},
	}
	if unicode.Version == unicodeVersion {
		testCases = append(testCases, testCasesUni...)
	} else {
		t.Logf("skipping %d test cases due to mismatched unicode versions; have %s, want %s", len(testCasesUni), unicode.Version, unicodeVersion)
	}

	for _, tc := range testCases {
		tr := newTestRand()

		pass, err := Repeat(FromRangeTable(tc.tab), "", 25).Password(tr)
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tc.expect, pass)
		assert.Truef(t, utf8.ValidString(pass),
			"utf8.ValidString(%q)", pass)
		allRunesAllowed(t, tc.tab, pass)
	}
}
