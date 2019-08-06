package password

import (
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

type charset struct {
	runes []rune
	count int
	grow  int
}

func NewCharset(template string) (func(count int) Template, error) {
	runes := []rune(template)
	if len(runes) < 2 {
		return nil, errors.New("strongroom/password: template too short")
	} else if len(runes) > maxUint32 {
		return nil, errors.New("strongroom/password: template too long")
	} else if !utf8.ValidString(template) {
		return nil, errors.New("strongroom/password: template contains invalid unicode rune")
	}

	var maxLen int
	seen := make(map[rune]struct{}, len(runes))
	for _, r := range runes {
		if len := utf8.RuneLen(r); len > maxLen {
			maxLen = len
		}

		if _, dup := seen[r]; dup {
			return nil, errors.New("strongroom/password: template contains duplicate rune")
		}
		seen[r] = struct{}{}
	}

	return func(count int) Template {
		if count <= 0 {
			panic("strongroom/password: count must be greater than zero")
		}

		grow := count * maxLen
		if maxLen > 1 && maxLen < utf8.UTFMax {
			grow += utf8.UTFMax
		}

		return &charset{runes, count, grow}
	}, nil
}

func (c *charset) Password(r io.Reader) (string, error) {
	var pass strings.Builder
	pass.Grow(c.grow)

	for i := 0; i < c.count; i++ {
		idx, err := readUint32n(r, uint32(len(c.runes)))
		if err != nil {
			return "", err
		}

		pass.WriteRune(c.runes[idx])
	}

	return pass.String(), nil
}
