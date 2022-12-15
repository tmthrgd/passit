package passit

import (
	"encoding/ascii85"
	"encoding/base32"
	"encoding/base64"
	"io"
	"strings"
)

type encodingGenerator struct {
	encodeToString func([]byte) string
	count          int
}

func newEncoding(count int, encodeToString func([]byte) string) Generator {
	if count < 0 {
		panic("passit: count must be positive")
	}

	return &encodingGenerator{encodeToString, count}
}

func encodeToHex(hextable string, src []byte) string {
	var sb strings.Builder
	sb.Grow(len(src) * 2)

	for _, v := range src {
		sb.WriteByte(hextable[v>>4])
		sb.WriteByte(hextable[v&0x0f])
	}

	return sb.String()
}

// HexLower returns a Generator that encodes count-bytes in lowercase hexadecimal.
func HexLower(count int) Generator {
	return newEncoding(count, func(src []byte) string {
		return encodeToHex("0123456789abcdef", src)
	})
}

// HexUpper returns a Generator that encodes count-bytes in uppercase hexadecimal.
func HexUpper(count int) Generator {
	return newEncoding(count, func(src []byte) string {
		return encodeToHex("0123456789ABCDEF", src)
	})
}

// Base32Std returns a Generator that encodes count-bytes with
// encoding/base32.StdEncoding without padding.
func Base32Std(count int) Generator {
	rawStd := base32.StdEncoding.WithPadding(base32.NoPadding)
	return newEncoding(count, rawStd.EncodeToString)
}

// Base32Hex returns a Generator that encodes count-bytes with
// encoding/base32.HexEncoding without padding.
func Base32Hex(count int) Generator {
	rawHex := base32.HexEncoding.WithPadding(base32.NoPadding)
	return newEncoding(count, rawHex.EncodeToString)
}

// Base64Std returns a Generator that encodes count-bytes with
// encoding/base64.RawStdEncoding.
func Base64Std(count int) Generator {
	return newEncoding(count, base64.RawStdEncoding.EncodeToString)
}

// Base64URL returns a Generator that encodes count-bytes with
// encoding/base64.RawURLEncoding.
func Base64URL(count int) Generator {
	return newEncoding(count, base64.RawURLEncoding.EncodeToString)
}

// Ascii85 returns a Generator that encodes count-bytes with encoding/ascii85.
func Ascii85(count int) Generator {
	return newEncoding(count, func(src []byte) string {
		dst := make([]byte, ascii85.MaxEncodedLen(len(src)))
		n := ascii85.Encode(dst, src)
		return string(dst[:n])
	})
}

func (eg *encodingGenerator) Password(r io.Reader) (string, error) {
	buf := make([]byte, eg.count)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", wrapReadError(err)
	}

	return eg.encodeToString(buf), nil
}
