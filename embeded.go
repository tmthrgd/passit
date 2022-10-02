package passit

import (
	_ "embed" // for go:embed
	"errors"
	"io"
	"strings"
	"sync"
)

//go:generate go run emoji_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 11.0.0
//go:generate go run emoji_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 13.0.0

type embededListVal struct {
	sync.Once
	list []string
}

type embededList struct {
	rawList string
	listVal *embededListVal
	sep     string
	count   int
}

func (l *embededList) Password(r io.Reader) (string, error) {
	if l.count <= 0 {
		return "", errors.New("passit: count must be greater than zero")
	}

	l.listVal.Do(func() {
		l.listVal.list = strings.Split(l.rawList, "\n")
	})
	list := l.listVal.list

	parts := make([]string, l.count)
	for i := range parts {
		idx, err := readIntN(r, len(list))
		if err != nil {
			return "", err
		}

		parts[i] = list[idx]
	}

	return strings.Join(parts, l.sep), nil
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
	effLargeWordlist    string
	effLargeWordlistVal embededListVal

	//go:embed emoji_11.0.txt
	emoji11List    string
	emoji11ListVal embededListVal

	//go:embed emoji_13.0.txt
	emoji13List    string
	emoji13ListVal embededListVal
)

// EFFLargeWordlist returns a Template that generates passwords of count words
// length by joining random words from the EFF Large Wordlist for Passphrases
// (eff_large_wordlist.txt).
func EFFLargeWordlist(count int) Template {
	return &embededList{effLargeWordlist, &effLargeWordlistVal, " ", count}
}

// Emoji11 returns a Template that generates passwords contain count number of emoji
// from the Unicode 11.0 emoji list.
func Emoji11(count int) Template {
	return &embededList{emoji11List, &emoji11ListVal, "", count}
}

// Emoji13 returns a Template that generates passwords contain count number of emoji
// from the Unicode 13.0 emoji list.
func Emoji13(count int) Template {
	return &embededList{emoji13List, &emoji13ListVal, "", count}
}
