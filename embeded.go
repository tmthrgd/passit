package passit

import (
	_ "embed" // for go:embed
	"io"
	"strings"
	"sync"
)

//go:generate go run emoji_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 11.0.0
//go:generate go run emoji_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 13.0.0

type embededList struct {
	once sync.Once
	raw  string
	list []string
}

func (e *embededList) Password(r io.Reader) (string, error) {
	e.once.Do(func() {
		e.list = strings.Split(e.raw, "\n")
		e.raw = ""
	})

	idx, err := readIntN(r, len(e.list))
	if err != nil {
		return "", err
	}

	return e.list[idx], nil
}

var (
	// This wordlist was taken from:
	// https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt.
	//
	// eff_large_wordlist.txt is licensed by the Electronic Frontier Foundation
	// under a CC BY 3.0 US license
	// (https://creativecommons.org/licenses/by/3.0/us/).
	//
	//go:embed eff_large_wordlist.txt
	effLargeWordlist string

	//go:embed emoji_11.0.txt
	emoji11List string

	//go:embed emoji_13.0.txt
	emoji13List string
)

// EFFLargeWordlist is a Template that returns a random word from the
// EFF Large Wordlist for Passphrases (eff_large_wordlist.txt).
var EFFLargeWordlist Template = &embededList{raw: effLargeWordlist}

// Emoji11 is a Template that returns a random emoji from the Unicode 11.0 emoji list.
var Emoji11 Template = &embededList{raw: emoji11List}

// Emoji13 is a Template that returns a random emoji from the Unicode 13.0 emoji list.
var Emoji13 Template = &embededList{raw: emoji13List}
