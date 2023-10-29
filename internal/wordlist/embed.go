package wordlist

import _ "embed" // for go:embed

var (
	// OrchardStreetMedium is a wordlist that was taken from:
	// https://github.com/sts10/orchard-street-wordlists/blob/016e8d6634c03dd08afd5b14753784aff325ebd1/lists/orchard-street-medium.txt.
	//
	// orchard-street-medium.txt is licensed by Sam Schlinkert under a
	// CC BY-SA 4.0 license (https://creativecommons.org/licenses/by-sa/4.0/).
	//
	//go:embed orchard-street-medium.txt
	OrchardStreetMedium string

	// OrchardStreetLong is a wordlist that was taken from:
	// https://github.com/sts10/orchard-street-wordlists/blob/016e8d6634c03dd08afd5b14753784aff325ebd1/lists/orchard-street-long.txt.
	//
	// orchard-street-long.txt is licensed by Sam Schlinkert under a
	// CC BY-SA 4.0 license (https://creativecommons.org/licenses/by-sa/4.0/).
	//
	//go:embed orchard-street-long.txt
	OrchardStreetLong string

	// OrchardStreetMedium is a wordlist that was taken from:
	// https://github.com/sts10/orchard-street-wordlists/blob/016e8d6634c03dd08afd5b14753784aff325ebd1/lists/orchard-street-alpha.txt.
	//
	// orchard-street-alpha.txt is licensed by Sam Schlinkert under a
	// CC BY-SA 4.0 license (https://creativecommons.org/licenses/by-sa/4.0/).
	//
	//go:embed orchard-street-alpha.txt
	OrchardStreetAlpha string

	// OrchardStreetLong is a wordlist that was taken from:
	// https://github.com/sts10/orchard-street-wordlists/blob/016e8d6634c03dd08afd5b14753784aff325ebd1/lists/orchard-street-qwerty.txt.
	//
	// orchard-street-qwerty.txt is licensed by Sam Schlinkert under a
	// CC BY-SA 4.0 license (https://creativecommons.org/licenses/by-sa/4.0/).
	//
	//go:embed orchard-street-qwerty.txt
	OrchardStreetQWERTY string

	// STS10Wordlist is a wordlist that was taken from:
	// https://github.com/sts10/generated-wordlists/tree/e0daeebbffbb/lists/1password-replacement
	// where it was called 1password-replacement.txt.
	//
	// 1password-replacement.txt is licensed by Sam Schlinkert under a CC BY 3.0
	// license (https://creativecommons.org/licenses/by/3.0/).
	//
	//go:embed sts10_wordlist.txt
	STS10Wordlist string

	// EFFLargeWordlist is a wordlist that was taken from:
	// https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt.
	//
	// eff_large_wordlist.txt is licensed by the Electronic Frontier Foundation
	// under a CC BY 3.0 US license
	// (https://creativecommons.org/licenses/by/3.0/us/).
	//
	//go:embed eff_large_wordlist.txt
	EFFLargeWordlist string

	// EFFShortWordlist1 is a wordlist that was taken from:
	// https://www.eff.org/files/2016/09/08/eff_short_wordlist_1.txt.
	//
	// eff_short_wordlist_1.txt is licensed by the Electronic Frontier Foundation
	// under a CC BY 3.0 US license
	// (https://creativecommons.org/licenses/by/3.0/us/).
	//
	//go:embed eff_short_wordlist_1.txt
	EFFShortWordlist1 string

	// EFFShortWordlist2 is a wordlist that was taken from:
	// https://www.eff.org/files/2016/09/08/eff_short_wordlist_2_0.txt.
	//
	// eff_short_wordlist_2_0.txt is licensed by the Electronic Frontier Foundation
	// under a CC BY 3.0 US license
	// (https://creativecommons.org/licenses/by/3.0/us/).
	//
	//go:embed eff_short_wordlist_2_0.txt
	EFFShortWordlist2 string
)
