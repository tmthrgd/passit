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
		{"1100110101010101110011111", "01"},
		{"9724130549434343534257971", "0123456789"},
		{"hxkebberczktmtylzpcqvlrzt", "abcdefghijklmnopqrstuvwxyz"},
		{"HXKEBBERCZKTMTYLZPCQVLRZT", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"hxKEBbErCZKtMtyLzpcqVlRzt", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"},
		{"HHG0Rbyp8RmXKhARJxsch7dlF", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"},
		{"u))$!!r$p+xtztll+ccq*_e+^", "abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+"},
		{"ΗσΡΗΣΗνΗωΛαΥΕπΠΝνξΩνΞΧνχζ", "ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω"},
		{"🍳💻👒💬🍧🍉🔱👗🍈🍳🍈💩💬💩🍈🍉👗💩💬👒👞💻🍳🐝🍧", "🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋"},
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
		{"9724130549434343534257971", Number},
		{"hxkebberczktmtylzpcqvlrzt", LatinLower},
		{"HXKEBBERCZKTMTYLZPCQVLRZT", LatinUpper},
		{"HXkebBeRczkTmTYlZPCQvLrZT", LatinMixed},
		{"rvgmjdip4r0benclx3uknnb93", LatinLowerNumber},
		{"RVGMJDIP4R0BENCLX3UKNNB93", LatinUpperNumber},
		{"hhg0rBYP8rMxkHarjXSCH7DLf", LatinMixedNumber},
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
		{"1100110101010101110011111", newTable("01")},
		{"9724130549434343534257971", newTable("0123456789")},
		{"hxkebberczktmtylzpcqvlrzt", newTable("abcdefghijklmnopqrstuvwxyz")},
		{"HXKEBBERCZKTMTYLZPCQVLRZT", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ")},
		{"hxKEBbErCZKtMtyLzpcqVlRzt", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")},
		{"776qHRofyHcNAX0H9niSXxTb5", newTable("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")},
		{"iyysppfsd~lhnh__~$$ewz&~u", newTable("abcdefghijklmnopqrstuvwxyz~!@#$%^&*()_+")},
		{"ΝμιΝλΝβΝωΦΒπΙθηαβδψβγτβυΜ", newTable("ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσςΤτΥυΦφΧχΨψΩω")},
		{"👒🚋🍉🍳🍈🍧🌴💻💬👒💬💩🍳💩💬🍧💻💩🍳🍉🎩🚋👒🐊🍈", newTable("🔱🍧👒🍉💬👞🛰🐝💅🍳🐊🐂🎩💩🍈👗🌴💻🚱🚋")},
		{"5f6E512B6bCf6d0fFfCa737DF", unicode.ASCII_Hex_Digit},
	}

	const unicodeVersion = "13.0.0"
	testCasesUni := []testCase{
		{"ḷɑꭏꭩɻḅﬁɩꜰꞤᴮᵑｓṗᴮꜻảṭŲꝶḕÿꝧᶜᵫ", unicode.Latin},
		{"ψἼ𐅀ὁὔὊ𝈼Νᶿὦ𝈖𝈧ϵ𝈭ὅ𐆅ἲᾒῪ͵𐅝῾𝈋𐅯Ὴ", unicode.Greek},
		{"߾߾؋￠₡₫﹩₹𞋿₡₶₧৳₱$₡৲﷼₼€₱𑿠₭₵֏", unicode.Sc},
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
