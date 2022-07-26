package passit

import (
	"encoding/ascii85"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

type encoding struct {
	encodeToString func([]byte) string
	count          int
}

// Hex returns a Template that encodes count-bytes with encoding/hex.
func Hex(count int) Template {
	return &encoding{hex.EncodeToString, count}
}

// Base32Std returns a Template that encodes count-bytes with
// encoding/base32.StdEncoding without padding.
func Base32Std(count int) Template {
	return &encoding{base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString, count}
}

// Base32Std returns a Template that encodes count-bytes with
// encoding/base32.HexEncoding without padding.
func Base32Hex(count int) Template {
	return &encoding{base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString, count}
}

// Base64Std returns a Template that encodes count-bytes with
// encoding/base64.RawStdEncoding.
func Base64Std(count int) Template {
	return &encoding{base64.RawStdEncoding.EncodeToString, count}
}

// Base64URL returns a Template that encodes count-bytes with
// encoding/base64.RawURLEncoding.
func Base64URL(count int) Template {
	return &encoding{base64.RawURLEncoding.EncodeToString, count}
}

// Ascii85 returns a Template that encodes count-bytes with encoding/ascii85.
func Ascii85(count int) Template {
	return &encoding{func(src []byte) string {
		dst := make([]byte, ascii85.MaxEncodedLen(len(src)))
		n := ascii85.Encode(dst, src)
		return string(dst[:n])
	}, count}
}

func (e *encoding) Password(r io.Reader) (string, error) {
	if e.count <= 0 {
		return "", errors.New("passit: count must be greater than zero")
	}

	buf := make([]byte, e.count)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}

	return e.encodeToString(buf), nil
}
