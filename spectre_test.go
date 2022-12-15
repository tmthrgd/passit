package passit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpectreTemplate(t *testing.T) {
	for _, tc := range []struct {
		gen       SpectreTemplate
		expected1 string
		expected2 string
	}{
		{SpectreBasic, "izJ24tHJ", "eSG10PbL"},
		{SpectreLong, "ZikzXuwuHeve1(", "Cuhu3-JufeVuzd"},
		{SpectreMaximum, "i7,o%yC4&fmQ1r*qfcWq", "T4!Vxx)nNumn(Dmem7nB"},
		{SpectreMedium, "Zik2~Puh", "Yav1(Mur"},
		{SpectreName, "hiskixuwu", "hevvewucu"},
		{SpectrePhrase, "zi kixpu hoy vezamcu", "qo nezfe vuz dixudre"},
		{SpectrePIN, "0778", "2449"},
		{SpectreShort, "His8", "Zup9"},
	} {
		tr := newTestRand()

		pass, err := tc.gen.Password(tr)
		if assert.NoErrorf(t, err, "failed to generate password: %+v", tc) {
			assert.Equalf(t, tc.expected1, pass, "incorrect password generated: %+v", tc)
		}

		pass, err = tc.gen.Password(tr)
		if assert.NoErrorf(t, err, "failed to generate password: %+v", tc) {
			assert.Equalf(t, tc.expected2, pass, "incorrect password generated: %+v", tc)
		}
	}
}

var sinkString string

func BenchmarkSpectreTemplate(b *testing.B) {
	tr := newTestRand()

	var pass string
	for n := 0; n < b.N; n++ {
		var err error
		pass, err = SpectreLong.Password(tr)
		if err != nil {
			require.NoErrorf(b, err, "failed to generate password")
		}
	}
	sinkString = pass
}
