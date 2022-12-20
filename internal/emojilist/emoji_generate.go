// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The code in this file was taken from golang.org/x/text/internal/export/unicode.

//go:build ignore

// Unicode table generator.
// Data read from the web.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	flag.Parse()
	loadEmoji()
	writeEmoji()
}

var emoji [][]rune

func emojiVersion() string {
	vers := gen_UnicodeVersion()
	return vers[:strings.LastIndex(vers, ".")]
}

func loadEmoji() {
	ucd_Parse(gen_OpenUnicodeFile("emoji", emojiVersion(), "emoji-test.txt"), func(p *ucd_Parser) {
		if p.String(1) == "fully-qualified" {
			emoji = append(emoji, p.Runes(0))
		}
	})
}

func writeEmoji() {
	filename := fmt.Sprintf("emoji_%s.txt", emojiVersion())
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", filename, err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatalf("Could not close file %s: %v", filename, err)
		}
	}()

	sort.Slice(emoji, func(i, j int) bool {
		emojiI, emojiJ := string(emoji[i]), string(emoji[j])

		// Sort first by number of bytes for countEmojiInString.
		if len(emojiI) != len(emojiJ) {
			return len(emojiI) < len(emojiJ)
		}

		// Then by string representation.
		return emojiI < emojiJ
	})

	for i, runes := range emoji {
		fmt.Fprint(f, string(runes))
		if i < len(emoji)-1 {
			fmt.Fprintln(f)
		}
	}
}
