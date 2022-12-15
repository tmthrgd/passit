// Package passit provides various password generators.
package passit

import "io"

// Template is an interface for generating passwords.
type Template interface {
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

// The TemplateFunc type is an adapter to allow the use of ordinary functions as
// password generators. If f is a function with the appropriate signature,
// TemplateFunc(f) is a Template that calls f.
type TemplateFunc func(r io.Reader) (string, error)

// Password implements Template, calling f(r).
func (f TemplateFunc) Password(r io.Reader) (string, error) {
	return f(r)
}
