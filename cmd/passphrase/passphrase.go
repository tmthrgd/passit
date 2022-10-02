// Command passphrase generates random passphrases using the EFF Large Wordlist for
// Passphrases (eff_large_wordlist.txt).
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
		fmt.Fprintln(out, "passphrase is a tool that generates random passphrases using the")
		fmt.Fprintln(out, "EFF Large Wordlist for Passphrases (eff_large_wordlist.txt).")
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	count := flag.Int("n", 6, "the number of words in the generated password")
	sep := flag.String("s", " ", "the separator to use between words")
	flag.Parse()

	pass, err := passit.Repeat(passit.EFFLargeWordlist, *sep, *count).Password(rand.Reader)
	if err != nil {
		log.SetFlags(0)
		log.Fatal(err)
	}

	fmt.Println(pass)
}
