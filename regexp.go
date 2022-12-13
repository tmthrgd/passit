package passit

import (
	"errors"
	"fmt"
	"io"
	"math/bits"
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxUnboundedRepeatCount = 15

type regexpGenerator func(*strings.Builder, io.Reader) error

// RegexpParser is a regular expressions parser that parses patterns into a Template
// that generates passwords matching the parsed regexp. The zero-value is a usable
// parser.
type RegexpParser struct {
	anyTab          *unicode.RangeTable
	specialCaptures map[string]SpecialCaptureFactory
}

// ParseRegexp is a shortcut for new(RegexpParser).Parse(pattern, flags).
func ParseRegexp(pattern string, flags syntax.Flags) (Template, error) {
	return new(RegexpParser).Parse(pattern, flags)
}

// SetAnyRangeTable sets the unicode.RangeTable used when generating any (.)
// characters or when restricting character classes ([a-z]) with a user provided
// one. By default a subset of ASCII is used.
func (p *RegexpParser) SetAnyRangeTable(tab *unicode.RangeTable) {
	p.anyTab = tab
}

// SetSpecialCapture adds a special capture factory to use for matching named
// captures. A regexp pattern such as "(?P<name>)" will invoke the factory and use
// the returned Template instead of the contents of the capture.
func (p *RegexpParser) SetSpecialCapture(name string, factory SpecialCaptureFactory) {
	if p.specialCaptures == nil {
		p.specialCaptures = make(map[string]SpecialCaptureFactory)
	}

	p.specialCaptures[name] = factory
}

type regexpTemplate struct{ gen regexpGenerator }

// Parse parses the regexp pattern according to the flags and returns a Template. It
// returns an error if the regexp is invalid. It uses regexp/syntax to parse the
// pattern.
//
// All regexp features supported by regexp/syntax are supported, though some may
// have no effect.
//
// Neither syntax.MatchNL nor syntax.FoldCase will have any effect whether present
// or not.
func (p *RegexpParser) Parse(pattern string, flags syntax.Flags) (Template, error) {
	// We intentionally never generate newlines, but passing syntax.MatchNL to
	// syntax.Parse simplifies the parsed character classes.
	flags |= syntax.MatchNL

	// The generator acts badly when used with syntax.FoldCase. Zero it out.
	flags &= ^syntax.FoldCase

	r, err := syntax.Parse(pattern, flags)
	if err != nil {
		return nil, err
	}

	// Simplify is great if one is trying to match the regexp, but less so when
	// generating a string from a regexp. Simplify transforms the regexp in ways
	// that might not be obvious. Because the output is dependent on the
	// specific stream read from the io.Reader, some of these transformations
	// are less than ideal. It's also possible that the simplifications
	// performed could differ between go releases.
	//
	// For example x{3,6} is transformed to xxx(?:x(?:xx?)?)?. This requires
	// more reads from the underlying io.Reader and is more work for the
	// generator.
	//
	// TODO(tmthrgd): It may be worth adding some simplifications of our own.
	//
	// r = r.Simplify()

	gen, err := p.parse(r)
	if err != nil {
		return nil, err
	}

	return regexpTemplate{gen}, nil
}

func (rt regexpTemplate) Password(r io.Reader) (string, error) {
	var b strings.Builder
	if err := rt.gen(&b, r); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (p *RegexpParser) parse(r *syntax.Regexp) (regexpGenerator, error) {
	switch r.Op {
	case syntax.OpEmptyMatch:
		return p.noop(r)
	case syntax.OpLiteral:
		return p.literal(r)
	case syntax.OpCharClass:
		return p.charClass(r)
	case syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		return p.anyCharNotNL(r)
	case syntax.OpBeginLine, syntax.OpEndLine,
		syntax.OpBeginText, syntax.OpEndText,
		syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return p.noop(r)
	case syntax.OpCapture:
		return p.capture(r)
	case syntax.OpStar:
		return p.star(r)
	case syntax.OpPlus:
		return p.plus(r)
	case syntax.OpQuest:
		return p.quest(r)
	case syntax.OpRepeat:
		return p.repeat(r)
	case syntax.OpConcat:
		return p.concat(r)
	case syntax.OpAlternate:
		return p.alternate(r)
	default:
		return nil, fmt.Errorf("passit: invalid regexp %q, unhandled op %s", r, r.Op)
	}
}

func (p *RegexpParser) parseMany(rs []*syntax.Regexp) ([]regexpGenerator, error) {
	gens := make([]regexpGenerator, len(rs))
	for i, r := range rs {
		gen, err := p.parse(r)
		if err != nil {
			return nil, err
		}

		gens[i] = gen
	}

	return gens, nil
}

func (*RegexpParser) noop(*syntax.Regexp) (regexpGenerator, error) {
	return func(*strings.Builder, io.Reader) error {
		return nil
	}, nil
}

func (*RegexpParser) literal(sr *syntax.Regexp) (regexpGenerator, error) {
	s := string(sr.Rune)
	return func(b *strings.Builder, r io.Reader) error {
		b.WriteString(s)
		return nil
	}, nil
}

func (p *RegexpParser) charClass(sr *syntax.Regexp) (regexpGenerator, error) {
	tab := new(unicode.RangeTable)
	for i := 0; i < len(sr.Rune); i += 2 {
		AppendToRangeTable(tab, sr.Rune[i], sr.Rune[i+1])
	}

	// intersectRangeTables requires that the first RangeTable have a Stride of
	// 1. This is safe as AppendToRangeTable only ever adds ranges with Stride
	// set to 1.
	tab = intersectRangeTables(tab, p.anyRangeTable())

	return p.charClassInternal(sr, tab)
}

func (p *RegexpParser) anyCharNotNL(sr *syntax.Regexp) (regexpGenerator, error) {
	return p.charClassInternal(sr, p.anyRangeTable())
}

func (*RegexpParser) charClassInternal(sr *syntax.Regexp, tab *unicode.RangeTable) (regexpGenerator, error) {
	count := countTableRunes(tab)
	if count == 0 {
		return nil, fmt.Errorf("passit: character class %s contains zero allowed runes", sr)
	}

	return func(b *strings.Builder, r io.Reader) error {
		v, err := readRune(r, tab, count)
		b.WriteRune(v)
		return err
	}, nil
}

func (p *RegexpParser) capture(sr *syntax.Regexp) (regexpGenerator, error) {
	if sr.Name != "" {
		return p.namedCapture(sr)
	}

	return p.parse(sr.Sub[0])
}

func (p *RegexpParser) namedCapture(sr *syntax.Regexp) (regexpGenerator, error) {
	factory, ok := p.specialCaptures[sr.Name]
	if !ok {
		return p.parse(sr.Sub[0])
	}

	tmpl, err := factory(sr)
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		pass, err := tmpl.Password(r)
		if err != nil {
			return err
		} else if !utf8.ValidString(pass) {
			return errors.New("passit: special capture output contains invalid unicode rune")
		}

		b.WriteString(pass)
		return nil
	}, nil
}

func (p *RegexpParser) star(sr *syntax.Regexp) (regexpGenerator, error) {
	return p.repeatInternal(sr, 0, maxUnboundedRepeatCount)
}

func (p *RegexpParser) plus(sr *syntax.Regexp) (regexpGenerator, error) {
	// We use maxUnboundedRepeatCount+1 here so that x{1,} and x+ are identical,
	// x{0,} and x* are already identical.
	return p.repeatInternal(sr, 1, maxUnboundedRepeatCount+1)
}

func (p *RegexpParser) quest(sr *syntax.Regexp) (regexpGenerator, error) {
	return p.repeatInternal(sr, 0, 1)
}

func (p *RegexpParser) repeat(sr *syntax.Regexp) (regexpGenerator, error) {
	max := sr.Max
	if max == -1 {
		max = sr.Min + maxUnboundedRepeatCount
	}

	return p.repeatInternal(sr, sr.Min, max)
}

func (p *RegexpParser) repeatInternal(sr *syntax.Regexp, min, max int) (regexpGenerator, error) {
	gen, err := p.parse(sr.Sub[0])
	if err != nil {
		return nil, err
	}

	N := max - min + 1
	if N < 1 || N > maxReadIntN {
		return nil, errors.New("passit: [min,max] range too large")
	}

	return func(b *strings.Builder, r io.Reader) error {
		n, err := readIntN(r, N)
		if err != nil {
			return err
		}

		for n += min; n > 0; n-- {
			if err := gen(b, r); err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func (p *RegexpParser) concat(sr *syntax.Regexp) (regexpGenerator, error) {
	gens, err := p.parseMany(sr.Sub)
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		for _, gen := range gens {
			if err := gen(b, r); err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func (p *RegexpParser) alternate(sr *syntax.Regexp) (regexpGenerator, error) {
	gens, err := p.parseMany(sr.Sub)
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		idx, err := readIntN(r, len(gens))
		if err != nil {
			return err
		}

		return gens[idx](b, r)
	}, nil
}

var rangeTableASCII = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x0020, Hi: 0x007e, Stride: 1},
	},
	LatinOffset: 1,
}

func (p *RegexpParser) anyRangeTable() *unicode.RangeTable {
	if p.anyTab != nil {
		return p.anyTab
	}

	return rangeTableASCII
}

// SpecialCaptureFactory represents a special capture factory to be used with
// (*RegexpParser).SetSpecialCapture.
type SpecialCaptureFactory func(*syntax.Regexp) (Template, error)

// SpecialCaptureBasic returns a special capture factory that doesn't accept any
// input and always returns the provided Template.
func SpecialCaptureBasic(tmpl Template) SpecialCaptureFactory {
	return func(sr *syntax.Regexp) (Template, error) {
		if sr.Sub[0].Op == syntax.OpEmptyMatch {
			return tmpl, nil
		}

		return nil, errors.New("passit: unsupported capture")
	}
}

// SpecialCaptureWithRepeat returns a special capture factory that parses the
// capture value for a count to be used with Repeat(tmpl, sep, count). If the
// capture is empty, the Template is returned directly.
func SpecialCaptureWithRepeat(tmpl Template, sep string) SpecialCaptureFactory {
	return func(sr *syntax.Regexp) (Template, error) {
		switch sr.Sub[0].Op {
		case syntax.OpEmptyMatch:
			return tmpl, nil
		case syntax.OpLiteral:
			count, err := strconv.ParseUint(string(sr.Sub[0].Rune), 10, bits.UintSize-1)
			if err != nil {
				return nil, fmt.Errorf("passit: failed to parse capture: %w", err)
			}

			return Repeat(tmpl, sep, int(count)), nil
		default:
			return nil, errors.New("passit: unsupported capture")
		}
	}
}
