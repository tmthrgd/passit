// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The code in this file was taken from golang.org/x/text/internal/export/unicode.

// +build generate

// Unicode table generator.
// Data read from the web.

package main

import (
	"flag"
	"fmt"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

func main() {
	flag.Parse()
	setupOutput()
	loadChars()
	loadProperties()
	printCategories()
	printSizes()
	flushOutput()
}

var output *gen_CodeWriter

func setupOutput() {
	output = gen_NewCodeWriter()
}

func flushOutput() {
	output.WriteGoFile("tables.go", "password")
}

func printf(format string, args ...interface{}) {
	fmt.Fprintf(output, format, args...)
}

func print(args ...interface{}) {
	fmt.Fprint(output, args...)
}

func println(args ...interface{}) {
	fmt.Fprintln(output, args...)
}

type Char struct {
	codePoint rune // if zero, this index is not a valid code point.
	category  string
}

const MaxChar = 0x10FFFF

var chars = make([]Char, MaxChar+1)
var props = make(map[string][]rune)

func loadChars() {
	ucd_Parse(gen_OpenUCDFile("UnicodeData.txt"), func(p *ucd_Parser) {
		codePoint := p.Rune(0)
		chars[codePoint] = Char{
			codePoint: codePoint,
			category:  p.String(ucd_GeneralCategory),
		}
	})
}

func loadProperties() {
	ucd_Parse(gen_OpenUCDFile("PropList.txt"), func(p *ucd_Parser) {
		name := p.String(1)
		props[name] = append(props[name], p.Rune(0))
	})
}

func printCategories() {
	println("import \"unicode\"\n\n")

	println("// unicodeVersion is the Unicode edition from which the tables are derived.")
	printf("const unicodeVersion = %q\n\n", gen_UnicodeVersion())

	// TODO(tmthrgd): Review these ranges.

	deprecated := rangetable.New(props["Deprecated"]...)
	ignorable := rangetable.New(props["Other_Default_Ignorable_Code_Point"]...)

	dumpRange("rangeTableASCII", func(code rune) bool {
		return code >= 0x20 && code <= 0x7e
	}, false)

	allowed := func(code rune) bool {
		if code <= 0x7e { // Special case ASCII.
			return code >= 0x20
		}

		if unicode.In(code, deprecated, ignorable) {
			return false
		}

		c := chars[code]
		switch c.category {
		case "":
			return false
		case "Lu", "Ll", "Lt", "Lo":
			return true
		case "Sm", "Sc", "So":
			return true
		}
		switch c.category[0] {
		case 'N':
			return true
		case 'P':
			return true
		default:
			return false
		}
	}
	dumpRange("allowedRangeTable", allowed, false)
	dumpRange("allowedRangeTableStride1", allowed, true)
}

type Op func(code rune) bool

func dumpRange(name string, inCategory Op, unstridify bool) {
	runes := []rune{}
	for i := range chars {
		r := rune(i)
		if inCategory(r) {
			runes = append(runes, r)
		}
	}
	printRangeTable(name, runes, unstridify)
}

func printRangeTable(name string, runes []rune, unstridify bool) {
	rt := rangetable.New(runes...)
	if unstridify {
		rt = unstridifyRangeTable(rt)
	}
	printf("var %s = &unicode.RangeTable{\n", name)
	println("\tR16: []unicode.Range16{")
	for _, r := range rt.R16 {
		printf("\t\t{Lo: %#04x, Hi: %#04x, Stride: %d},\n", r.Lo, r.Hi, r.Stride)
		range16Count++
	}
	println("\t},")
	if len(rt.R32) > 0 {
		println("\tR32: []unicode.Range32{")
		for _, r := range rt.R32 {
			printf("\t\t{Lo: %#x, Hi: %#x, Stride: %d},\n", r.Lo, r.Hi, r.Stride)
			range32Count++
		}
		println("\t},")
	}
	if rt.LatinOffset > 0 {
		printf("\tLatinOffset: %d,\n", rt.LatinOffset)
	}
	printf("}\n\n")
}

var range16Count = 0  // Number of entries in the 16-bit range tables.
var range32Count = 0  // Number of entries in the 32-bit range tables.
var foldPairCount = 0 // Number of fold pairs in the exception tables.

func printSizes() {
	println()
	printf("// Range entries: %d 16-bit, %d 32-bit, %d total.\n", range16Count, range32Count, range16Count+range32Count)
	range16Bytes := range16Count * 3 * 2
	range32Bytes := range32Count * 3 * 4
	printf("// Range bytes: %d 16-bit, %d 32-bit, %d total.\n", range16Bytes, range32Bytes, range16Bytes+range32Bytes)
}

func unstridifyRangeTable(tab *unicode.RangeTable) *unicode.RangeTable {
	rt := &unicode.RangeTable{
		R16: tab.R16[:len(tab.R16):len(tab.R16)],
		R32: tab.R32[:len(tab.R32):len(tab.R32)],
	}

	for i := 0; i < len(rt.R16); i++ {
		if r16 := rt.R16[i]; r16.Stride != 1 {
			size := int((r16.Hi-r16.Lo)/r16.Stride) + 1
			rt.R16 = append(rt.R16, make([]unicode.Range16, size-1)...)
			copy(rt.R16[i+size:], rt.R16[i+1:])

			for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
				if r <= unicode.MaxLatin1 {
					rt.LatinOffset++
				}

				rt.R16[i] = unicode.Range16{Lo: uint16(r), Hi: uint16(r), Stride: 1}
				i++
			}
			i--
		} else if r16.Hi <= unicode.MaxLatin1 {
			rt.LatinOffset++
		}
	}

	for i := 0; i < len(rt.R32); i++ {
		if r32 := rt.R32[i]; r32.Stride != 1 {
			size := int((r32.Hi-r32.Lo)/r32.Stride) + 1
			rt.R32 = append(rt.R32, make([]unicode.Range32, size-1)...)
			copy(rt.R32[i+size:], rt.R32[i+1:])

			for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
				rt.R32[i] = unicode.Range32{Lo: uint32(r), Hi: uint32(r), Stride: 1}
				i++
			}
			i--
		}
	}

	return rt
}