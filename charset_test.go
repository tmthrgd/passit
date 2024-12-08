package passit

import (
	"strings"
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
		{"0110100100010010000000010", "01"},
		{"2352984966922368666874797", "0123456789"},
		{"yzxeishgyluaruksywwtcxdji", "abcdefghijklmnopqrstuvwxyz"},
		{"YZXEISHGYLUARUKSYWWTCXDJI", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"yXisHgYluArukyWwtCXdjIRmD", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"},
		{"ovNa1Os7MObQ0ruaoUCwj2DdZ", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"},
		{"y+)rvfut_lhnehk%ljjtpk#ji", "abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+"},
		{"ΓΤΞΙΧυχΖτξυΔβΧψΥΠΖΟωωγοοτ", "ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω"},
		{"👒💩👗🎩🚋🚱💬🚋🌴🌴🍳👒🎩🍉🛰💅🛰🛰🛰💅💻🍈🐝🍳🐝", "🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋"},
	} {
		tr := newTestRand()

		pass, err := Repeat(FromCharset(tc.charset), "", 25).Password(tr)
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
		{"2352984966922368666874797", Digit},
		{"yzxeishgyluaruksywwtcxdji", LatinLower},
		{"YZXEISHGYLUARUKSYWWTCXDJI", LatinUpper},
		{"YxIShGyLUaRUKYwWTcxDJirMd", LatinMixed},
		{"4rd6x4ix2e8rwqhkqk08smzst", LatinLowerDigit},
		{"4RD6X4IX2E8RWQHKQK08SMZST", LatinUpperDigit},
		{"OVnA1oS7moBq0RUAOucWJ2dDz", LatinMixedDigit},
		{"'*,?:+-`)-_^+?,/]#{/_};\"[", ASCIINoLettersNumbers},
		{"38\\#>-#2+]}&]+\"%%1{}!'.-\"", ASCIINoLetters},
		{")lMM\\KmzULOyAQ$WB>xgf{$K#", ASCIIGraphic},
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

func TestASCIINoLettersNumbers(t *testing.T) {
	var expect strings.Builder
	for r := range unicode.MaxASCII + 1 {
		if unicode.In(r, unicode.P, unicode.S) {
			expect.WriteRune(r)
		}
	}
	assert.Equal(t, expect.String(), ASCIINoLettersNumbers.(*asciiGenerator).s,
		"ASCIINoLettersNumbers doesn't match ASCII characters in Unicode categories P or S")
}

func TestASCIINoLetters(t *testing.T) {
	var expect strings.Builder
	for r := range unicode.MaxASCII + 1 {
		if unicode.In(r, unicode.N, unicode.P, unicode.S) {
			expect.WriteRune(r)
		}
	}
	assert.Equal(t, expect.String(), ASCIINoLetters.(*asciiGenerator).s,
		"ASCIINoLetters doesn't match ASCII characters in Unicode categories N, P or S")
}

func TestASCIIGraphic(t *testing.T) {
	var expect strings.Builder
	for r := range unicode.MaxASCII + 1 {
		if unicode.IsGraphic(r) && !unicode.IsSpace(r) {
			expect.WriteRune(r)
		}
	}
	assert.Equal(t, expect.String(), ASCIIGraphic.(*asciiGenerator).s,
		"ASCIIGraphic doesn't match ASCII characters in Unicode categories L, M, N, P, S, Zs")
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
		{"0110100100010010000000010", newTable("01")},
		{"2352984966922368666874797", newTable("0123456789")},
		{"yzxeishgyluaruksywwtcxdji", newTable("abcdefghijklmnopqrstuvwxyz")},
		{"YZXEISHGYLUARUKSYWWTCXDJI", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ")},
		{"yXisHgYluArukyWwtCXdjIRmD", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")},
		{"elDQrEixCERGqhkQeK2mZs3TP", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")},
		{"m~yfj(ihz_*b&*^t_@@hd^r@+", newTable("abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+")},
		{"ΕξγΡτρυΛοδρΗΔτχπηΛεωωΖζζο", newTable("ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω")},
		{"🍉💩💻💅🛰🚱🍳🛰🔱🔱👒🍉💅🍧🐂🐝🐂🐂🐂🐝🚋💬🐊👒🐊", newTable("🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋")},
		{"ED9Ed60F4A148f2068a49Ab7f", unicode.ASCII_Hex_Digit},
	}

	const unicodeVersion = "15.0.0"
	testCasesUni := []testCase{
		{"ʊⅴXǞǟṓŅ𝼈Ņꝧ𐞓Ꞇ𝼕áÆẞḔꝅḩʎĀĮꟑＴｆ", unicode.Latin},
		{"ἳ𝈛ῥᵡ𐅺όΫ𐅷𐆎ἓ𐅖Ί𝈂ΗᾁῈϼᴧρὺᵞ𐅰Ϟ𐅬θ", unicode.Greek},
		{"₷₼૱₧﷼૱₼𑿟৳௹𞋿₪௹＄₻₾₩₵₡𞋿$⃀₲￡¢", unicode.Sc},
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

func BenchmarkCharsetPassword(b *testing.B) {
	for _, tc := range []struct{ name, charset string }{
		{"ASCII", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"},
		{"Unicode", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω0123456789"},
	} {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			benchmarkGeneratorPassword(b, FromCharset(tc.charset))
		})
	}
}
