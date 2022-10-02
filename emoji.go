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

type emojiListVal struct {
	sync.Once
	list []string
}

type emoji struct {
	rawList string
	listVal *emojiListVal
	count   int
}

func (e *emoji) Password(r io.Reader) (string, error) {
	if e.count <= 0 {
		return "", errors.New("passit: count must be greater than zero")
	}

	e.listVal.Do(func() {
		e.listVal.list = strings.Split(e.rawList, "\n")
	})
	list := e.listVal.list

	emoji := make([]string, e.count)
	for i := range emoji {
		idx, err := readIntN(r, len(list))
		if err != nil {
			return "", err
		}

		emoji[i] = list[idx]
	}

	return strings.Join(emoji, ""), nil
}

var (
	//go:embed emoji_11.0.txt
	emoji11List    string
	emoji11ListVal emojiListVal

	//go:embed emoji_13.0.txt
	emoji13List    string
	emoji13ListVal emojiListVal
)

// Emoji11 returns a Template that generates passwords contain count number of emoji
// from the Unicode 13.0 emoji list.
func Emoji11(count int) Template { return &emoji{emoji11List, &emoji11ListVal, count} }

// Emoji13 returns a Template that generates passwords contain count number of emoji
// from the Unicode 13.0 emoji list.
func Emoji13(count int) Template { return &emoji{emoji13List, &emoji13ListVal, count} }
