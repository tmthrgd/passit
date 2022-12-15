package emojilist

import _ "embed" // for go:embed

//go:generate go run emoji_generate.go emoji_generate_gen.go emoji_generate_ucd.go -unicode 13.0.0

var (
	// Unicode13 is a list of emoji from Unicode 13.0.
	//
	//go:embed emoji_13.0.txt
	Unicode13 string
)
