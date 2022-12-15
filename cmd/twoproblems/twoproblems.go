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
	"os"
	"regexp/syntax"

	"go.tmthrgd.dev/passit"
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
	rp.SetSpecialCapture("largeword", passit.SpecialCaptureWithRepeat(passit.EFFLargeWordlist, " "))
	rp.SetSpecialCapture("short1word", passit.SpecialCaptureWithRepeat(passit.EFFShortWordlist1, " "))
	rp.SetSpecialCapture("short2word", passit.SpecialCaptureWithRepeat(passit.EFFShortWordlist2, " "))
	rp.SetSpecialCapture("emoji", passit.SpecialCaptureWithRepeat(passit.Emoji13, ""))

	tmpl, err := rp.Parse(flag.Arg(0), syntax.Perl)
	if err != nil {
		return fmt.Errorf("twoproblems: failed to parse %q pattern: %w", flag.Arg(0), err)
	}

	pass, err := tmpl.Password(bufio.NewReader(rand.Reader))
	if err != nil {
		return fmt.Errorf("twoproblems: failed to generate password: %w", err)
	}

	fmt.Println(pass)
	return nil
}
