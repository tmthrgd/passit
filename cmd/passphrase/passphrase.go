// Command passphrase generates random passphrases using either
// the EFF Large Wordlist for Passphrases (eff_large_wordlist.txt),
// the EFF Short Wordlist for Passphrases #1 (eff_short_wordlist_1.txt), or
// the EFF Short Wordlist for Passphrases #2 (eff_short_wordlist_2_0.txt).
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"

	"go.tmthrgd.dev/passit"
)

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "passphrase is a tool that generates random passphrases using either")
		fmt.Fprintln(out, "the EFF Large Wordlist for Passphrases (eff_large_wordlist.txt),")
		fmt.Fprintln(out, "the EFF Short Wordlist for Passphrases #1 (eff_short_wordlist_1.txt), or")
		fmt.Fprintln(out, "the EFF Short Wordlist for Passphrases #2 (eff_short_wordlist_2_0.txt).")
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	list := flag.String("l", "large", "the wordlist to use; valid options are large, short1 and short2")
	count := flag.Int("n", 6, "the number of words in the generated password")
	sep := flag.String("s", " ", "the separator to use between words")
	flag.Parse()

	var tmpl passit.Template
	switch *list {
	case "large":
		tmpl = passit.EFFLargeWordlist
	case "short1":
		tmpl = passit.EFFShortWordlist1
	case "short2":
		tmpl = passit.EFFShortWordlist2
	default:
		log.SetFlags(0)
		log.Fatal("passit: invalid wordlist specified")
	}

	pass, err := passit.Repeat(tmpl, *sep, *count).Password(rand.Reader)
	if err != nil {
		log.SetFlags(0)
		log.Fatal(err)
	}

	fmt.Println(pass)
}
