package passit

import (
	_ "embed" // for go:embed
	"errors"
	"io"
	"strings"
	"sync"
)

//go:generate go run emoji_generate.go unicode_generate_gen.go unicode_generate_ucd.go -unicode 13.0.0

type emoji struct {
	list  []string
	count int
}

func (e *emoji) Password(r io.Reader) (string, error) {
	if e.count <= 0 {
		return "", errors.New("passit: count must be greater than zero")
	}

	emoji := make([]string, e.count)
	for i := range emoji {
		idx, err := readIntN(r, len(e.list))
		if err != nil {
			return "", err
		}

		emoji[i] = e.list[idx]
	}

	return strings.Join(emoji, ""), nil
}

//go:embed emoji_13.0.txt
var emoji13List string

var emoji13ListVal struct {
	sync.Once
	list []string
}

// Emoji13 returns a Template that generates passwords contain count number of emoji
// from the Unicode 13.0 emoji list.
func Emoji13(count int) Template {
	emoji13ListVal.Do(func() {
		emoji13ListVal.list = strings.Split(emoji13List, "\n")
	})
	return &emoji{emoji13ListVal.list, count}
}
