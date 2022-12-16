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
		{SpectreBasic, "Rfr9cSj2", "qt86yQw7"},
		{SpectreLong, "Dadl8(WeraHinc", "GewyBoru7=Fubu"},
		{SpectreMaximum, "R2.%r7#UK60qtJ!2wT23", "gN*LO#!SkMImynnfwa0?"},
		{SpectreMedium, "Dad9~Dun", "Yur2;Gov"},
		{SpectreName, "xoqduquwe", "rahricege"},
		{SpectrePhrase, "dadl quw neyhino gov", "wabj ruc liwbujo now"},
		{SpectrePIN, "2352", "9849"},
		{SpectreShort, "Xoq2", "Lod9"},
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

func BenchmarkSpectrePassword(b *testing.B) {
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
