package password

import (
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

func FromWords(list ...string) (func(count int) Template, error) {
	if len(list) < 2 {
		return nil, errors.New("strongroom/password: list too short")
	} else if len(list) > maxUint32 {
		return nil, errors.New("strongroom/password: list too long")
	}

	seen := make(map[string]struct{}, len(list))
	for _, word := range list {
		if len(word) < 1 {
			return nil, errors.New("strongroom/password: empty word in list")
		} else if !utf8.ValidString(word) {
			return nil, errors.New("strongroom/password: word contains invalid unicode rune")
		} else if idx := strings.IndexFunc(word, notAllowed); idx >= 0 {
			r, _ := utf8.DecodeRuneInString(word[idx:])
			return nil, fmt.Errorf("strongroom/password: word contains prohibited rune %U", r)
		} else if strings.IndexFunc(word, unicode.IsSpace) >= 0 {
			return nil, errors.New("strongroom/password: word contains space")
		}

		if _, dup := seen[word]; dup {
			return nil, errors.New("strongroom/password: list contains duplicate word")
		}
		seen[word] = struct{}{}
	}

	list = append([]string(nil), list...)
	return func(count int) Template { return &words{list, count} }, nil
}

func (w *words) Password(r io.Reader) (string, error) {
	if w.count <= 0 {
		return "", errors.New("strongroom/password: count must be greater than zero")
	}

	words := make([]string, w.count)
	for i := range words {
		idx, err := readUint32n(r, uint32(len(w.list)))
		if err != nil {
			return "", err
		}

		words[i] = w.list[idx]
	}

	return strings.Join(words, " "), nil
}

var defaultWords struct {
	tmpl func(int) Template
	sync.Once
}

func DefaultWords(count int) Template {
	defaultWords.Do(func() {
		tmpl, err := FromWords(strings.Split(defaultWordlist, "\n")...)
		if err != nil {
			panic("strongroom/password: internal error: " + err.Error())
		}
		defaultWords.tmpl = tmpl
	})
	return defaultWords.tmpl(count)
}
