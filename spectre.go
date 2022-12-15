package passit

import (
	"errors"
	"io"
	"strings"
)

// SpectreTemplate implements a variant of the Spectre / Master Password encoding
// algorithm by Maarten Billemont for generating passwords.
//
// This algorithm is not compatible with any of the officially published algorithms,
// but it does produce passwords using the same templates that are indistinguishable
// from the official algorithm. Unlike that algorithm, this doesn't exhibit a modulo
// bias.
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

func (st SpectreTemplate) readTemplate(r io.Reader) (string, error) {
	// A benchmark of just (SpectreTemplate).Password shows strings.Split being
	// responsible for 88% of all allocated data.
	return readSliceN(r, strings.Split(string(st), ":"))
}

// Password implements Template.
func (st SpectreTemplate) Password(r io.Reader) (string, error) {
	template, err := st.readTemplate(r)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.Grow(len(template))

	for _, c := range []byte(template) {
		chars, ok := spectreChars[c]
		if !ok {
			return "", errors.New("passit: template contains invalid character")
		}

		n, err := readIntN(r, len(chars))
		if err != nil {
			return "", err
		}

		sb.WriteByte(chars[n])
	}

	return sb.String(), nil
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
