// Command passphrase generates random passphrases using one of the embedded
// wordlists supported by passit.
package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"go.tmthrgd.dev/passit"
	"go.tmthrgd.dev/passit/internal/wordlists"
	"golang.org/x/text/language"
)

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "passphrase is a tool that generates random passphrases using")
		fmt.Fprintln(out, "one of the embedded wordlists supported by passit.")
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
	list := flag.String("l", "orchard:long",
		"the wordlist to use; valid options are:\n"+
			"orchard:medium, orchard:long, orchard:alpha, "+
			"orchard:qwerty, eff:large, eff:short1 and eff:short2")
	words := flag.Int("n", 6, "the number of words in the generated passphrase")
	sep := flag.String("s", " ", "the separator to use between words")
	titleCase := flag.Bool("t", false, "generate a title case passphrase")
	upperCase := flag.Bool("u", false, "generate an upper case passphrase")
	count := flag.Int("c", 1, "the number of passwords to generate, one per line")
	flag.Parse()

	gen := wordlists.NameToGenerator(*list)
	if gen == nil {
		return errors.New("passphrase: invalid wordlist specified")
	}

	if *upperCase {
		gen = passit.UpperCase(gen)
	} else if *titleCase {
		gen = passit.TitleCase(gen, language.English)
	}

	gen = passit.Repeat(gen, *sep, *words)

	r := bufio.NewReader(rand.Reader)
	for range *count {
		pass, err := gen.Password(r)
		if err != nil {
			return err
		}
		fmt.Println(pass)
	}

	return nil
}
