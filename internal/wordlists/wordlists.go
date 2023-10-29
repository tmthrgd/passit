// Package wordlists provides utility functions for working with wordlists.
package wordlists

import "go.tmthrgd.dev/passit"

// NameToGenerator returns a [passit.Generator] that corresponds to the particular
// name. It returns nil if the name is unknown.
func NameToGenerator(name string) passit.Generator {
	switch name {
	case "orchard:medium":
		return passit.OrchardStreetMedium
	case "orchard:long":
		return passit.OrchardStreetLong
	case "orchard:alpha":
		return passit.OrchardStreetAlpha
	case "orchard:qwerty":
		return passit.OrchardStreetQWERTY
	case "eff:large", "eff":
		return passit.EFFLargeWordlist
	case "eff:short1":
		return passit.EFFShortWordlist1
	case "eff:short2":
		return passit.EFFShortWordlist2
	default:
		return nil
	}
}
