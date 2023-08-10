package emojilist

import _ "embed" // for go:embed

//go:generate go run emoji_generate.go emoji_generate_gen.go emoji_generate_ucd.go -unicode 13.0.0
//go:generate go run emoji_generate.go emoji_generate_gen.go emoji_generate_ucd.go -unicode 15.0.0

var (
	// Unicode13 is a list of fully-qualified emoji from Unicode 13.0.
	//
	// This is generated from the Unicode 13.0 emoji-test.txt file.
	// See https://www.unicode.org/Public/emoji/13.0/emoji-test.txt.
	//
	//go:embed emoji_13.0.txt
	Unicode13 string

	// Unicode15 is a list of fully-qualified emoji from Unicode 15.0.
	//
	// This is generated from the Unicode 15.0 emoji-test.txt file.
	// See https://www.unicode.org/Public/emoji/15.0/emoji-test.txt.
	//
	//go:embed emoji_15.0.txt
	Unicode15 string
)
