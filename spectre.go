// This implements part of the Spectre / Master Password algorithm by
// Maarten Billemont.

package passit

import (
	"errors"
	"io"
	"strings"
)

// SpectreTemplate implements v1+ of the Spectre / Master Password encoding
// algorithm for generating passwords.
//
// Note: It only implements the encoding of the seed bytes into a password string
// and not the entire algorithm.
type SpectreTemplate string

// These are the standard templates defined by Spectre / Master Password.
const (
	SpectreMaximum SpectreTemplate = "anoxxxxxxxxxxxxxxxxx:axxxxxxxxxxxxxxxxxno"
	SpectreLong    SpectreTemplate = "CvcvnoCvcvCvcv:CvcvCvcvnoCvcv:CvcvCvcvCvcvno:CvccnoCvcvCvcv:CvccCvcvnoCvcv:CvccCvcvCvcvno:CvcvnoCvccCvcv:CvcvCvccnoCvcv:CvcvCvccCvcvno:CvcvnoCvcvCvcc:CvcvCvcvnoCvcc:CvcvCvcvCvccno:CvccnoCvccCvcv:CvccCvccnoCvcv:CvccCvccCvcvno:CvcvnoCvccCvcc:CvcvCvccnoCvcc:CvcvCvccCvccno:CvccnoCvcvCvcc:CvccCvcvnoCvcc:CvccCvcvCvccno"
	SpectreMedium  SpectreTemplate = "CvcnoCvc:CvcCvcno"
	SpectreBasic   SpectreTemplate = "aaanaaan:aannaaan:aaannaaa"
	SpectreShort   SpectreTemplate = "Cvcn"
	SpectrePIN     SpectreTemplate = "nnnn"
	SpectreName    SpectreTemplate = "cvccvcvcv"
	SpectrePhrase  SpectreTemplate = "cvcc cvc cvccvcv cvc:cvc cvccvcvcv cvcv:cv cvccv cvc cvcvccv"
)

// Password implements Template.
func (st SpectreTemplate) Password(r io.Reader) (string, error) {
	idx, err := readUint8(r)
	if err != nil {
		return "", err
	}

	// This call to strings.Split doesn't allocate, presumably as Go understands
	// the slice doesn't escape.
	templates := strings.Split(string(st), ":")
	template := templates[int(idx)%len(templates)]

	buf := make([]byte, len(template))
	if _, err := readBytes(r, buf); err != nil {
		return "", err
	}

	for i, c := range []byte(template) {
		chars, ok := spectreChars[c]
		if !ok {
			return "", errors.New("passit: template contains invalid character")
		}

		buf[i] = chars[int(buf[i])%len(chars)]
	}

	return string(buf), nil
}

var spectreChars = map[byte]string{
	'V': "AEIOU",
	'C': "BCDFGHJKLMNPQRSTVWXYZ",
	'v': "aeiou",
	'c': "bcdfghjklmnpqrstvwxyz",
	'A': "AEIOUBCDFGHJKLMNPQRSTVWXYZ",
	'a': "AEIOUaeiouBCDFGHJKLMNPQRSTVWXYZbcdfghjklmnpqrstvwxyz",
	'n': "0123456789",
	'o': "@&%?,=[]_:-+*$#!'^~;()/.",
	'x': "AEIOUaeiouBCDFGHJKLMNPQRSTVWXYZbcdfghjklmnpqrstvwxyz0123456789!@#$%^&*()",
	' ': " ",
}
