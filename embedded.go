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

// STS10Wordlist is a Generator that returns a random word from Sam Schlinkert's
// '1Password Replacement List'.
//
// This wordlist is licensed by Sam Schlinkert under a CC BY 3.0 license.
var STS10Wordlist Generator = &embeddedGenerator{raw: &wordlist.STS10Wordlist}

// EFFLargeWordlist is a Generator that returns a random word from the
// EFF Large Wordlist for Passphrases (eff_large_wordlist.txt).
//
// This wordlist is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license.
var EFFLargeWordlist Generator = &embeddedGenerator{raw: &wordlist.EFFLargeWordlist}

// EFFShortWordlist1 is a Generator that returns a random word from the
// EFF Short Wordlist for Passphrases #1 (eff_short_wordlist_1.txt).
//
// This wordlist is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license.
var EFFShortWordlist1 Generator = &embeddedGenerator{raw: &wordlist.EFFShortWordlist1}

// EFFShortWordlist2 is a Generator that returns a random word from the
// EFF Short Wordlist for Passphrases #2 (eff_short_wordlist_2_0.txt).
//
// This wordlist is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license.
var EFFShortWordlist2 Generator = &embeddedGenerator{raw: &wordlist.EFFShortWordlist2}

// Emoji13 is a Generator that returns a random emoji from the Unicode 13.0 emoji
// list.
var Emoji13 Generator = &embeddedGenerator{raw: &emojilist.Unicode13}
