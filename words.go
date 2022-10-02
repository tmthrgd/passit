package passit

import (
	"errors"
	"io"
	"strings"
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
