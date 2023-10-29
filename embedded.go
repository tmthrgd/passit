package passit

import (
	"io"
	"strings"
	"sync"

	"go.tmthrgd.dev/passit/internal/emojilist"
	"go.tmthrgd.dev/passit/internal/wordlist"
)

type embeddedGenerator struct {
	once sync.Once

	// Using a pointer to the embedded string allows for dead-code elimination
	// to completely eliminate the embedded string if the Generator variable is
	// never referenced.
	raw *string

	list []string
}

func (eg *embeddedGenerator) Password(r io.Reader) (string, error) {
	eg.once.Do(func() {
		eg.list = strings.Split(*eg.raw, "\n")
	})
	return readSliceN(r, eg.list)
}

// OrchardStreetMedium is a Generator that returns a random word from
// Sam Schlinkert's Orchard Street Medium List.
//
// It contains 8,192 words and has 13.000 bits of entropy per word. This list is
// uniquely decodable and can be used with or without separators.
//
// This wordlist is licensed by Sam Schlinkert under a CC BY-SA 4.0 license.
var OrchardStreetMedium Generator = &embeddedGenerator{raw: &wordlist.OrchardStreetMedium}

// OrchardStreetLong is a Generator that returns a random word from
// Sam Schlinkert's Orchard Street Long List.
//
// It contains 17,576 words and has 14.101 bits of entropy per word. This list is
// uniquely decodable and can be used with or without separators.
//
// This wordlist is licensed by Sam Schlinkert under a CC BY-SA 4.0 license.
var OrchardStreetLong Generator = &embeddedGenerator{raw: &wordlist.OrchardStreetLong}

// OrchardStreetAlpha is a Generator that returns a random word from
// Sam Schlinkert's Orchard Street Alpha List.
//
// It contains 1,296 words and has 10.340 bits of entropy per word. This list is
// uniquely decodable and can be used with or without separators.
//
// This wordlist is licensed by Sam Schlinkert under a CC BY-SA 4.0 license.
var OrchardStreetAlpha Generator = &embeddedGenerator{raw: &wordlist.OrchardStreetAlpha}

// OrchardStreetQWERTY is a Generator that returns a random word from
// Sam Schlinkert's Orchard Street QWERTY List.
//
// It contains 1,296 words and has 10.340 bits of entropy per word. This list is
// uniquely decodable and can be used with or without separators.
//
// This wordlist is licensed by Sam Schlinkert under a CC BY-SA 4.0 license.
var OrchardStreetQWERTY Generator = &embeddedGenerator{raw: &wordlist.OrchardStreetQWERTY}

// STS10Wordlist is a Generator that returns a random word from Sam Schlinkert's
// '1Password Replacement List'.
//
// It contains 18,208 words and has 14.152 bits of entropy per word. This list is
// not uniquely decodable and should only be used with separators.
//
// This wordlist is licensed by Sam Schlinkert under a CC BY 3.0 license.
var STS10Wordlist Generator = &embeddedGenerator{raw: &wordlist.STS10Wordlist}

// EFFLargeWordlist is a Generator that returns a random word from the
// EFF Large Wordlist for Passphrases (eff_large_wordlist.txt).
//
// It contains 7,776 words and has 12.925 bits of entropy per word.
//
// This wordlist is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license.
var EFFLargeWordlist Generator = &embeddedGenerator{raw: &wordlist.EFFLargeWordlist}

// EFFShortWordlist1 is a Generator that returns a random word from the
// EFF Short Wordlist for Passphrases #1 (eff_short_wordlist_1.txt).
//
// It contains 1,296 words and has 10.340 bits of entropy per word.
//
// This wordlist is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license.
var EFFShortWordlist1 Generator = &embeddedGenerator{raw: &wordlist.EFFShortWordlist1}

// EFFShortWordlist2 is a Generator that returns a random word from the
// EFF Short Wordlist for Passphrases #2 (eff_short_wordlist_2_0.txt).
//
// It contains 1,296 words and has 10.340 bits of entropy per word.
//
// This wordlist is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license.
var EFFShortWordlist2 Generator = &embeddedGenerator{raw: &wordlist.EFFShortWordlist2}

// Emoji13 is a Generator that returns a random fully-qualified emoji from the
// Unicode 13.0 emoji list.
var Emoji13 Generator = &embeddedGenerator{raw: &emojilist.Unicode13}

// Emoji15 is a Generator that returns a random fully-qualified emoji from the
// Unicode 15.0 emoji list.
var Emoji15 Generator = &embeddedGenerator{raw: &emojilist.Unicode15}

// EmojiLatest is an alias for the latest supported emoji list.
var EmojiLatest = Emoji15
