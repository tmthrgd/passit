package passit

import (
	"errors"
	"io"
	"strings"
)

type emoji struct {
	list  []string
	count int
}

// Emoji13 returns a Template that generates passwords contain count number of emoji
// from the Unicode 13.0 emoji list.
func Emoji13(count int) Template { return &emoji{unicodeEmoji, count} }

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
