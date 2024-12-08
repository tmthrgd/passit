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
	"net/url"
	"os"
	"regexp/syntax"
	"strconv"
	"strings"

	"go.tmthrgd.dev/passit"
	"go.tmthrgd.dev/passit/internal/wordlists"
	"golang.org/x/text/language"
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
	count := flag.Int("c", 1, "the number of passwords to generate, one per line")
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

	r := bufio.NewReader(rand.Reader)
	for range *count {
		pass, err := gen.Password(r)
		if err != nil {
			return fmt.Errorf("twoproblems: failed to generate password: %w", err)
		}
		fmt.Println(pass)
	}

	return nil
}

func wordlist(sr *syntax.Regexp) (passit.Generator, error) {
	switch sr.Sub[0].Op {
	case syntax.OpEmptyMatch:
		return passit.OrchardStreetLong, nil
	case syntax.OpLiteral:
		p, err := parseWordlistParams(string(sr.Sub[0].Rune))
		if err != nil {
			return nil, fmt.Errorf("twoproblems: failed to parse word parameters: %w", err)
		}

		gen := passit.OrchardStreetLong
		if v, ok := p["list"]; ok {
			name := strings.ToLower(v)
			gen = wordlists.NameToGenerator(name)
			if gen == nil {
				return nil, fmt.Errorf("twoproblems: unsupported wordlist %q", name)
			}
		}

		if v, ok := p["case"]; ok {
			switch case_ := strings.ToLower(v); case_ {
			case "lower":
				// Already lower case.
			case "upper":
				gen = passit.UpperCase(gen)
			case "title":
				gen = passit.TitleCase(gen, language.English)
			default:
				return nil, fmt.Errorf("twoproblems: unsupported case %q", case_)
			}
		}

		sep := " "
		if v, ok := p["sep"]; ok {
			sep = v
		}

		if v, ok := p["count"]; ok {
			count, err := strconv.ParseUint(v, 10, bits.UintSize-1)
			if err != nil {
				return nil, fmt.Errorf("twoproblems: failed to parse wordlist count: %w", err)
			}

			gen = passit.Repeat(gen, sep, int(count))
		}

		return gen, nil
	default:
		return nil, errors.New("twoproblems: unsupported capture")
	}
}

func parseWordlistParams(query string) (map[string]string, error) {
	// This is like url.ParseQuery, but uses PathUnescape instead of
	// QueryUnescape and lowercases all keys.

	v := make(map[string]string)
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		key, value, _ := strings.Cut(key, "=")
		key = strings.ToLower(key)
		key, err := url.PathUnescape(key)
		if err != nil {
			return nil, err
		}
		value, err = url.PathUnescape(value)
		if err != nil {
			return nil, err
		}
		v[key] = value
	}
	return v, nil
}
