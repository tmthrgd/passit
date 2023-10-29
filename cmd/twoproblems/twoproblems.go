// Command twoproblems generates random passwords based on a regular expression
// template.
package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/bits"
	"os"
	"regexp/syntax"
	"strconv"
	"strings"

	"go.tmthrgd.dev/passit"
	"go.tmthrgd.dev/passit/internal/wordlists"
)

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "twoproblems is a tool that generates random passwords")
		fmt.Fprintln(out, "based on a regular expression template.")
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	if err := main1(); err != nil {
		log.SetFlags(0)
		log.Fatal(err)
	}
}

func main1() error {
	flag.Parse()

	if flag.NArg() != 1 {
		return errors.New("twoproblems: missing regexp template argument")
	}

	var rp passit.RegexpParser
	rp.SetSpecialCapture("word", wordlist)
	rp.SetSpecialCapture("emoji", passit.SpecialCaptureWithRepeat(passit.EmojiLatest, ""))

	gen, err := rp.Parse(flag.Arg(0), syntax.Perl)
	if err != nil {
		return fmt.Errorf("twoproblems: failed to parse %q pattern: %w", flag.Arg(0), err)
	}

	pass, err := gen.Password(bufio.NewReader(rand.Reader))
	if err != nil {
		return fmt.Errorf("twoproblems: failed to generate password: %w", err)
	}

	fmt.Println(pass)
	return nil
}

func wordlist(sr *syntax.Regexp) (passit.Generator, error) {
	switch sr.Sub[0].Op {
	case syntax.OpEmptyMatch:
		return passit.OrchardStreetLong, nil
	case syntax.OpLiteral:
		name, rest, more := strings.Cut(string(sr.Sub[0].Rune), "/")
		name = strings.ToLower(name)

		gen := wordlists.NameToGenerator(name)
		if gen == nil && name == "" {
			gen = passit.OrchardStreetLong
		}
		if gen == nil {
			return nil, fmt.Errorf("twoproblems: unsupported wordlist %q", name)
		}

		if !more {
			return gen, nil
		}

		countStr, sep, ok := strings.Cut(rest, "/")
		if !ok {
			sep = " "
		}

		count, err := strconv.ParseUint(countStr, 10, bits.UintSize-1)
		if err != nil {
			return nil, fmt.Errorf("twoproblems: failed to parse wordlist count: %w", err)
		}

		return passit.Repeat(gen, sep, int(count)), nil
	default:
		return nil, errors.New("twoproblems: unsupported capture")
	}
}
