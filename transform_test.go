package passit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestLowerCase(t *testing.T) {
	tr := newTestRand()

	pass, err := LowerCase(Repeat(LatinMixed, "", 25)).Password(tr)
	require.NoError(t, err)
	assert.Equal(t, "yxishgyluarukywwtcxdjirmd", pass)
}

func TestUpperCase(t *testing.T) {
	tr := newTestRand()

	pass, err := UpperCase(Repeat(LatinMixed, "", 25)).Password(tr)
	require.NoError(t, err)
	assert.Equal(t, "YXISHGYLUARUKYWWTCXDJIRMD", pass)
}

func TestTitleCase(t *testing.T) {
	tr := newTestRand()

	pass, err := TitleCase(Repeat(OrchardStreetLong, " ", 10), language.English).Password(tr)
	require.NoError(t, err)
	assert.Equal(t, "Agreed Stopping Brilliant Elongated Richness Populous Sprung Grassland Stamens Dined", pass)

	pass, err = TitleCase(Repeat(OrchardStreetLong, "-", 10), language.English).Password(tr)
	require.NoError(t, err)
	assert.Equal(t, "Emphasizes-Weaving-Pickup-Cascades-Newborn-Provider-Noisy-Retailer-Compromise-Inventors", pass)

	pass, err = TitleCase(Repeat(OrchardStreetLong, "_", 10), language.English).Password(tr)
	require.NoError(t, err)
	assert.Equal(t, "Biplane_kingship_ambient_altered_injustices_precedes_yearning_kitten_chop_carefully", pass)
}
