package passit

import (
	_ "embed" // for go:embed
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type words struct {
	list  []string
	count int
}

// FromWords returns a Template factory that generates passwords of count words
// length by joining random words from list. It returns an error if the list of
// words is invalid.
func FromWords(list ...string) (func(count int) Template, error) {
	if len(list) < 2 {
		return nil, errors.New("passit: list too short")
	} else if len(list) > maxReadIntN {
		return nil, errors.New("passit: list too long")
	}

	seen := make(map[string]struct{}, len(list))
	for _, word := range list {
		if len(word) < 1 {
			return nil, errors.New("passit: empty word in list")
		} else if !utf8.ValidString(word) {
			return nil, errors.New("passit: word contains invalid unicode rune")
		} else if idx := strings.IndexFunc(word, notAllowed); idx >= 0 {
			r, _ := utf8.DecodeRuneInString(word[idx:])
			return nil, fmt.Errorf("passit: word contains prohibited rune %U", r)
		} else if strings.IndexFunc(word, unicode.IsSpace) >= 0 {
			return nil, errors.New("passit: word contains space")
		}

		if _, dup := seen[word]; dup {
			return nil, errors.New("passit: list contains duplicate word")
		}
		seen[word] = struct{}{}
	}

	list = append([]string(nil), list...)
	return func(count int) Template { return &words{list, count} }, nil
}

func (w *words) Password(r io.Reader) (string, error) {
	if w.count <= 0 {
		return "", errors.New("passit: count must be greater than zero")
	}

	words := make([]string, w.count)
	for i := range words {
		idx, err := readIntN(r, len(w.list))
		if err != nil {
			return "", err
		}

		words[i] = w.list[idx]
	}

	return strings.Join(words, " "), nil
}

// This wordlist was taken from:
// https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt.
//
// eff_large_wordlist.txt is licensed by the Electronic Frontier Foundation under a
// CC BY 3.0 US license (https://creativecommons.org/licenses/by/3.0/us/).
//
//go:embed eff_large_wordlist.txt
var effLargeWordlist string

var effLargeWordlistVal struct {
	sync.Once
	list []string
}

// EFFLargeWordlist returns a Template that generates passwords of count words
// length by joining random words from the EFF Large Wordlist for Passphrases
// (eff_large_wordlist.txt).
func EFFLargeWordlist(count int) Template {
	effLargeWordlistVal.Do(func() {
		effLargeWordlistVal.list = strings.Split(effLargeWordlist, "\n")
	})
	return &words{effLargeWordlistVal.list, count}
}
