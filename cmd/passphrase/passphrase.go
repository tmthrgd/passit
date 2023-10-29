// Command passphrase generates random passphrases using either
// Sam Schlinkert's '1Password Replacement List' (1password-replacement.txt),
// the EFF Large Wordlist for Passphrases (eff_large_wordlist.txt),
// the EFF Short Wordlist for Passphrases #1 (eff_short_wordlist_1.txt), or
// the EFF Short Wordlist for Passphrases #2 (eff_short_wordlist_2_0.txt).
package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"go.tmthrgd.dev/passit"
	"go.tmthrgd.dev/passit/internal/wordlists"
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
	list := flag.String("l", "orchard:medium",
		"the wordlist to use; valid options are orchard:medium, orchard:long, orchard:alpha, "+
			"orchard:qwerty, sts10, eff:large / eff, eff:short1 and eff:short2")
	count := flag.Int("n", 6, "the number of words in the generated password")
	sep := flag.String("s", " ", "the separator to use between words")
	flag.Parse()

	gen := wordlists.NameToGenerator(*list)
	if gen == nil {
		return errors.New("passphrase: invalid wordlist specified")
	}

	pass, err := passit.Repeat(gen, *sep, *count).Password(rand.Reader)
	if err != nil {
		return err
	}

	fmt.Println(pass)
	return nil
}
