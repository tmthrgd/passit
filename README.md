# passit

[`passit`](https://pkg.go.dev/go.tmthrgd.dev/passit) is a password generation
toolkit for Go. It features a variety of different password generators from charsets
to regular expressions, wordlists and emoji.

All generators implement the following interface:

```go
// Generator is an interface for generating passwords.
type Generator interface {
	// Password returns a randomly generated password using r as the source of
	// randomness.
	//
	// The returned password may or may not be deterministic with respect to r.
	// All generators in this package are deterministic unless otherwise noted.
	//
	// The output of r should be indistinguishable from a random string of the
	// same length. This is a property of a good CSPRNG. Fundamentally the
	// strength of the generated password is only as good as the provided source
	// of randomness.
	//
	// r should implement the io.ByteReader interface for improved performance.
	Password(r io.Reader) (string, error)
}
```

## Generators

The package provides a number of generators that produce output from a fixed set:

| Generator             | Description                                               | Examples                |
| --------------------- | --------------------------------------------------------- | ----------------------- |
| `Digit`               | [0-9]                                                     | "0" "7"                 |
| `LatinLower`          | [a-z]                                                     | "a" "j"                 |
| `LatinLowerDigit`     | [a-z0-9]                                                  | "a" "j" "0" "7"         |
| `LatinUpper`          | [A-Z]                                                     | "A" "J"                 |
| `LatinUpperDigit`     | [A-Z0-9]                                                  | "A" "J" "0" "7"         |
| `LatinMixed`          | [a-zA-Z]                                                  | "a" "j" "A" "J"         |
| `LatinMixedDigit`     | [a-zA-Z0-9]                                               | "a" "j" "A" "J" "0" "7" |
| `OrchardStreetMedium` | A word from Sam Schlinkert's Orchard Street Medium List   | "abandon" "forest"      |
| `OrchardStreetLong`   | A word from Sam Schlinkert's Orchard Street Long List     | "abandon" "mobility"    |
| `OrchardStreetAlpha`  | A word from Sam Schlinkert's Orchard Street Alpha List    | "abbot" "points"        |
| `OrchardStreetQWERTY` | A word from Sam Schlinkert's Orchard Street QWERTY List   | "access" "peg"          |
| `STS10Wordlist`       | A word from Sam Schlinkert's '1Password Replacement List' | "aback" "loophole"      |
| `EFFLargeWordlist`    | A word from the EFF Large Wordlist for Passphrases        | "abacus" "partition"    |
| `EFFShortWordlist1`   | A word from the EFF Short Wordlist for Passphrases #1     | "acid" "match"          |
| `EFFShortWordlist2`   | A word from the EFF Short Wordlist for Passphrases #2     | "aardvark" "jaywalker"  |
| `Emoji13`             | A Unicode 13.0 fully-qualified emoji                      | "⌚" "🕸️" "🧎🏾‍♀️"          |
| `Emoji15`             | A Unicode 15.0 fully-qualified emoji                      | "⌚" "🏎️" "🧏🏿‍♂️"          |
| `HexLower`            | Lowercase hexadecimal encoding                            | "66e94bd4ef8a2c3b"      |
| `HexUpper`            | Uppercase hexadecimal encoding                            | "66E94BD4EF8A2C3B"      |
| `Base32`              | Base32 standard encoding                                  | "M3UUXVHPRIWDW"         |
| `Base32Hex`           | Base32 hexadecimal encoding                               | "CRKKNL7FH8M3M"         |
| `Base64`              | Base64 standard encoding                                  | "ZulL1O+KLDs"           |
| `Base64URL`           | Base64 URL encoding                                       | "ZulL1O-KLDs"           |
| `Ascii85`             | Ascii85 encoding                                          | "B'Dt<mtrYX"            |
| `SpectreMaximum`      | The Spectre maximum template                              | "i7,o%yC4&fmQ1r*qfcWq"  |
| `SpectreLong`         | The Spectre long template                                 | "ZikzXuwuHeve1("        |
| `SpectreMedium`       | The Spectre medium template                               | "Zik2~Puh"              |
| `SpectreBasic`        | The Spectre basic template                                | "izJ24tHJ"              |
| `SpectreShort`        | The Spectre short template                                | "His8"                  |
| `SpectrePIN`          | The Spectre PIN template                                  | "0778"                  |
| `SpectreName`         | The Spectre name template                                 | "hiskixuwu"             |
| `SpectrePhrase`       | The Spectre phrase template                               | "zi kixpu hoy vezamcu"  |
| `Empty`               | Empty string                                              | ""                      |
| `Hyphen`              | ASCII hyphen-minus                                        | "-"                     |
| `Space`               | ASCII space                                               | " "                     |

The package also provides a number of generators that produce output based on user input:

| Generator        | Description                                           |
| ---------------- | ----------------------------------------------------- |
| `String`         | A fixed string                                        |
| `RegexpParser`   | Password that matches a regular expression pattern    |
| `FromCharset`    | A rune from a charset string                          |
| `FromRangeTable` | A rune from a `unicode.RangeTable`                    |
| `FromSlice`      | A string from a slice of strings                      |

There are also a number of 'helper' generators that interact with the output of other generators:

| Generator         | Description                                                                     |
| ----------------- | ------------------------------------------------------------------------------- |
| `Alternate`       | Select a generator at random                                                    |
| `Join`            | Concatenate the output of multiple generators                                   |
| `Repeat`          | Invoke a generator multiple times and concatenate the output                    |
| `RandomRepeat`    | Invoke a generator a random number of times and concatenate the output          |
| `RejectionSample` | Continually invoke a generator until the output passes a test                   |
| `Transform`       | Invoke a generator and convert the output according to a user supplied function |
| `LowerCase`       | Invoke a generator and convert the output to lower case                         |
| `UpperCase`       | Invoke a generator and convert the output to upper case                         |
| `TitleCase`       | Invoke a generator and convert the output to language-specific title case       |

Most generators only generate a single of something, be it a rune, ASCII character
or word. For generating longer passwords use `Repeat` or `RandomRepeat`, possibly
with `Join` or `Alternate`. In this way the various generators can be composed to
generator arbitrarily long and complex passwords, or short and simple passwords as
is needed.

The generators are designed to map from a random string / stream to a text password.
This is not designed to be a reversible process and decoding the password to the
original random string is not possible.

## Commands

Two commands for easy CLI password generation are provided.

### passphrase

`passphrase` is a tool that generates random passphrases using one of the embedded
wordlists supported by passit.

```shell
$ go install go.tmthrgd.dev/passit/cmd/passphrase@latest
$ passphrase -n 5 -s -
assumes-forth-humanities-exemption-paid
```

### twoproblems

`twoproblems` is a tool that generates random passwords based on a regular
expression template.

```shell
$ go install go.tmthrgd.dev/passit/cmd/twoproblems@latest
$ twoproblems '[[:alpha:]]{15}-[[:digit:]]{3}[[:punct:]]{2}'
KsMtvHnSOmqjIll-277&$
$ twoproblems '[[:alnum:]]{5}-(?P<word>count=5&sep=-)-[[:punct:]]{5}'
7Gfyv-degree-currency-revolt-coined-requested-*@`}[
```

Two special captures (`(?P<name>)`) are supported:

1. `(?P<word>)`: A word from any of the supported wordlists. This takes parameters
 similar to a URL query string (e.g. `x=1&y=2&y=3`), including supporting percent
 escaping. The supported parameters are:
   - `case` transform the word to a given case ('lower' - default, 'upper' or
     'title');
   - `count` generate N multiple words instead of just one;
   - `list` the name of a supported wordlist to use ('orchard:medium',
     'orchard:long' - default, 'orchard:alpha, 'orchard:qwerty, 'sts10',
     'eff:large' / 'eff', 'eff:short1' or 'eff:short2');
   - `sep` a separator to insert between words (defaults to a space).
1. `(?P<emoji>)`: A Unicode 15.0 emoji returned from `Emoji15`. This can take a
 number to generate multiple emoji.

## Sources of randomness

**Note:** Remember that wrapping the `io.Reader` with `bufio.NewReader` (if it
doesn't already implement `io.ByteReader`) will greatly improve the performance of
the generators.

For generating random passwords, `Password` should be called with
[`crypto/rand.Reader`](https://pkg.go.dev/crypto/rand#pkg-variables). Avoid using
poor quality sources of randomness like math/rand.

```go
import "crypto/rand"

func ExampleEFFLargeWordlist_WithCryptoRand() {
	pass, _ := passit.Repeat(passit.EFFLargeWordlist, "-", 4).Password(rand.Reader)
	fmt.Println(pass)
}
```

For generating deterministic passwords, `Password` should be called with a
deterministic stream that should be indistinguishable from a random string of the
same length. Good examples of sources for this would be
[HKDF](https://pkg.go.dev/golang.org/x/crypto/hkdf),
[ChaCha20](https://pkg.go.dev/golang.org/x/crypto/chacha20) or AES-CTR with proper
key generation. Care must be taken when using deterministic password generation as
the generated password is only ever as good as the provided source of randomness.

```go
func ExampleEFFLargeWordlist_WithHKDF() {
	secret, salt, info := []byte("secret"), []byte("salt"), []byte("info")

	r := hkdf.New(sha512.New, secret, salt, info)

	pass, _ := passit.Repeat(passit.EFFLargeWordlist, "-", 4).Password(r)
	fmt.Println(pass) // Output: king-unflawed-vagrancy-laxative
}

func ExampleEFFLargeWordlist_WithChaCha20() {
	key := []byte("secret key for password generate")

	c, _ := chacha20.NewUnauthenticatedCipher(key, []byte("IV for PWDGN"))
	r := cipher.StreamReader{S: c, R: zeroReader{}}

	pass, _ := passit.Repeat(passit.EFFLargeWordlist, "-", 4).Password(r)
	fmt.Println(pass) // Output: penny-enclose-preoccupy-sappy
}

func ExampleEFFLargeWordlist_WithAESCTR() {
	key := []byte("secret key for password generate")

	block, _ := aes.NewCipher(key)
	ctr := cipher.NewCTR(block, []byte("IV for PWDGN CTR"))
	r := cipher.StreamReader{S: ctr, R: zeroReader{}}

	pass, _ := passit.Repeat(passit.EFFLargeWordlist, "-", 4).Password(r)
	fmt.Println(pass) // Output: juncture-net-unseen-pegboard
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}
```

## License

This library is Copyright (c) 2022, Tom Thorogood and is licensed under a
BSD 3-Clause License.

Embedded wordlists are subject to their own copyright license referenced in the
documentation.
