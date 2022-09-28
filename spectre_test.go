package passit

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpectreTemplate(t *testing.T) {
	for _, tc := range []struct {
		template SpectreTemplate
		seed     string
		expected string
	}{
		{SpectreBasic, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "FLI88cDK"},
		{SpectreLong, "25265011c0d0dc225f7a7885df1783428fac12a8f7515e59d398ff72b4a2204a", "WawiYarp2@Kodh"},
		{SpectreLong, "466351a41a15916a2e8ccd4e125424fab92223dd79678bf9214351b5bea274c5", "TeweBacg0$Tobe"},
		{SpectreLong, "5a8f33260a099de514defe173c993359f4e683c4f73181586b26610757a7de3d", "Wewa9$YaqdDaje"},
		{SpectreLong, "9114ea61e93da8806ebab1063f19a17d15634c028d71a5f4a682c5bedabaa894", "ZurdYoda6:Jogs"},
		{SpectreLong, "bc5b4ebba378da55853a06a69026fc4879e7e6461dffdbc4edb91626258c8e02", "KoyvTocoVeyx8*"},
		{SpectreLong, "e3ef2a8374e859f883b6ddf46838bd79db30e60a6cb2730202452ea312a1bc43", "LiheCuwhSerz6)"},
		{SpectreLong, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "ReqoCenuXonu1?"},
		{SpectreMaximum, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "FB22U#U*LPFWlWxaxK2."},
		{SpectreMedium, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "ReqMon0)"},
		{SpectreName, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "reqmonajo"},
		{SpectreName, "fde3fb8cca862176ca4db377b8f604942ed68da7525255346843a64d5e30d342", "wesruqori"},
		{SpectrePhrase, "96bccba4899a3622e5acb5da66d1b8bb68ca3652dbe6af992db8f3c326ca2beb", "zowp quy roxzuyu qim"},
		{SpectrePhrase, "c771ab70ead018152e482e0ec49970a0183ead1e2708cb2b4ad9a481a918db07", "lek yubgiguko ruzo"},
		{SpectrePhrase, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "re monnu mit jededda"},
		{SpectrePIN, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "3648"},
		{SpectreShort, "e90de236c69488dc45125d9d636f1b799579e9025fe16ce9987a344685c4b9e5", "Req8"},
	} {
		seed, err := hex.DecodeString(tc.seed)
		require.NoErrorf(t, err, "failed to decode seed: %+v", tc)

		pass, err := tc.template.Password(bytes.NewReader(seed))
		if assert.NoErrorf(t, err, "failed to generate password: %+v", tc) {
			assert.Equalf(t, tc.expected, pass, "incorrect password generated: %+v", tc)
		}
	}
}

var sinkString string

func BenchmarkSpectreTemplate(b *testing.B) {
	seed, err := hex.DecodeString("25265011c0d0dc225f7a7885df1783428fac12a8f7515e59d398ff72b4a2204a")
	require.NoError(b, err, "failed to decode seed")

	var (
		r    bytes.Reader
		pass string
	)
	for n := 0; n < b.N; n++ {
		r.Reset(seed)

		var err error
		pass, err = SpectreLong.Password(&r)
		if err != nil {
			require.NoErrorf(b, err, "failed to generate password")
		}
	}
	sinkString = pass
}
