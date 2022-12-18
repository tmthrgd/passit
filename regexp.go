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
)

const maxUnboundedRepeatCount = 15

type regexpGenerator func(*strings.Builder, io.Reader) error

// RegexpParser is a regular expressions parser that parses patterns into a
// Generator that generates passwords matching the parsed regexp. The zero-value is
// a usable parser.
type RegexpParser struct {
	anyTab          *unicode.RangeTable
	specialCaptures map[string]SpecialCaptureFactory
}

// ParseRegexp is a shortcut for new(RegexpParser).Parse(pattern, flags).
func ParseRegexp(pattern string, flags syntax.Flags) (Generator, error) {
	return new(RegexpParser).Parse(pattern, flags)
}

// SetAnyRangeTable sets the unicode.RangeTable used when generating any (.)
// characters or when restricting character classes ([a-z]) with a user provided
// one. By default a subset of ASCII is used. Calling SetAnyRangeTable(nil) will
// reset the RegexpParser back to the default.
//
// The regexp Generator is only deterministic if the same unicode.RangeTable is
// used. Be aware that the builtin unicode.X tables are subject to change as new
// versions of Unicode are released and are not suitable for deterministic use.
func (p *RegexpParser) SetAnyRangeTable(tab *unicode.RangeTable) {
	p.anyTab = tab
}

// SetSpecialCapture adds a special capture factory to use for matching named
// captures. A regexp pattern such as "(?P<name>)" will invoke the factory and use
// the returned Generator instead of the contents of the capture.
func (p *RegexpParser) SetSpecialCapture(name string, factory SpecialCaptureFactory) {
	if p.specialCaptures == nil {
		p.specialCaptures = make(map[string]SpecialCaptureFactory)
	}
	p.specialCaptures[name] = factory
}

// Parse parses the regexp pattern according to the flags and returns a Generator.
// It returns an error if the regexp is invalid. It uses regexp/syntax to parse the
// pattern.
//
// All regexp features supported by regexp/syntax are supported, though some may
// have no effect.
func (p *RegexpParser) Parse(pattern string, flags syntax.Flags) (Generator, error) {
	// Note: The FoldCase, OneLine, DotNL and NonGreedy flags can be set or
	//   cleared within the pattern.

	// If we're not going to generate newlines, we can set syntax.MatchNL in
	// flags. This simplifies the parsed character classes and avoids needing to
	// call anyCharNotNL. It does this without changing the output of the
	// Generator.
	if !p.hasAnyNL() {
		flags |= syntax.MatchNL
	}

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

	return p.parse(r)
}

func (rg regexpGenerator) Password(r io.Reader) (string, error) {
	var b strings.Builder
	if err := rg(&b, r); err != nil {
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
	case syntax.OpAnyCharNotNL:
		return p.anyCharNotNL(r)
	case syntax.OpAnyChar:
		return p.anyChar(r)
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

func (p *RegexpParser) literal(sr *syntax.Regexp) (regexpGenerator, error) {
	if sr.Flags&syntax.FoldCase != 0 {
		return p.foldedLiteral(sr)
	}

	return p.rawLiteral(sr.Rune), nil
}

func (*RegexpParser) rawLiteral(runes []rune) regexpGenerator {
	s := string(runes)
	return func(b *strings.Builder, r io.Reader) error {
		b.WriteString(s)
		return nil
	}
}

func (p *RegexpParser) foldedLiteral(sr *syntax.Regexp) (regexpGenerator, error) {
	gens := make([]regexpGenerator, 0, len(sr.Rune))
	litStart := -1
	for i, c := range sr.Rune {
		// SimpleFold(c) returns c if there are no equivalent runes.
		if unicode.SimpleFold(c) == c {
			if litStart < 0 {
				litStart = i
			}
			continue
		}
		if litStart >= 0 {
			gens = append(gens, p.rawLiteral(sr.Rune[litStart:i+1]))
			litStart = -1
		}

		gen, err := p.foldedRune(c)
		if err != nil {
			return nil, err
		}
		gens = append(gens, gen)
	}

	if litStart >= 0 {
		gens = append(gens, p.rawLiteral(sr.Rune[litStart:]))
	}

	if len(gens) == 1 {
		return gens[0], nil
	}

	return p.concatGens(gens), nil
}

func (p *RegexpParser) foldedRune(c rune) (regexpGenerator, error) {
	// We generate a syntax.Regexp here and pass it to charClass rather than
	// generating the unicode.RangeTable directly so that we get a nicer error
	// message.
	sr := &syntax.Regexp{Op: syntax.OpCharClass}
	sr.Rune = append(sr.Rune[:0], c, c)

	for f := unicode.SimpleFold(c); f != c; f = unicode.SimpleFold(f) {
		sr.Rune = append(sr.Rune, f, f)
	}

	// We don't need to sort sr.Rune as regexp/syntax ensures that the rune
	// present in the literal is always the minimum rune. See
	// regexp/syntax.minFoldRune.

	return p.charClass(sr)
}

func (p *RegexpParser) charClass(sr *syntax.Regexp) (regexpGenerator, error) {
	anyTab := p.anyRangeTable()
	var tab unicode.RangeTable
	for i := 0; i < len(sr.Rune); i += 2 {
		addIntersectingRunes(&tab, sr.Rune[i], sr.Rune[i+1], anyTab)
	}

	setLatinOffset(&tab)
	return p.charClassInternal(sr, &tab)
}

func (p *RegexpParser) anyCharNotNL(sr *syntax.Regexp) (regexpGenerator, error) {
	return p.charClassInternal(sr, p.anyRangeTableNoNL())
}

func (p *RegexpParser) anyChar(sr *syntax.Regexp) (regexpGenerator, error) {
	return p.charClassInternal(sr, p.anyRangeTable())
}

func (*RegexpParser) charClassInternal(sr *syntax.Regexp, tab *unicode.RangeTable) (regexpGenerator, error) {
	count := countRunesInTable(tab)
	if count == 0 {
		return nil, fmt.Errorf("passit: character class %s contains zero allowed runes", sr)
	}

	return func(b *strings.Builder, r io.Reader) error {
		idx, err := readIntN(r, count)
		b.WriteRune(getRuneInTable(tab, idx))
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

	gen, err := factory(sr)
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		pass, err := gen.Password(r)
		b.WriteString(pass)
		return err
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

	// N can never overflow as syntax.Parse will return an error if min or max
	// exceed 1000.
	N := max - min + 1

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

	return p.concatGens(gens), nil
}

func (p *RegexpParser) concatGens(gens []regexpGenerator) regexpGenerator {
	return func(b *strings.Builder, r io.Reader) error {
		for _, gen := range gens {
			if err := gen(b, r); err != nil {
				return err
			}
		}

		return nil
	}
}

func (p *RegexpParser) alternate(sr *syntax.Regexp) (regexpGenerator, error) {
	gens, err := p.parseMany(sr.Sub)
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		gen, err := readSliceN(r, gens)
		if err != nil {
			return err
		}

		return gen(b, r)
	}, nil
}

func (p *RegexpParser) hasAnyNL() bool {
	// rangeTableASCII doesn't include \n so we only need to test this if
	// SetAnyRangeTable was called.
	return p.anyTab != nil && unicode.Is(p.anyTab, '\n')
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

func (p *RegexpParser) anyRangeTableNoNL() *unicode.RangeTable {
	if !p.hasAnyNL() {
		return p.anyRangeTable()
	}

	return removeNLFromRangeTable(p.anyRangeTable())
}

// SpecialCaptureFactory represents a special capture factory to be used with
// (*RegexpParser).SetSpecialCapture.
type SpecialCaptureFactory func(*syntax.Regexp) (Generator, error)

// SpecialCaptureBasic returns a special capture factory that doesn't accept any
// input and always returns the provided Generator.
func SpecialCaptureBasic(gen Generator) SpecialCaptureFactory {
	return func(sr *syntax.Regexp) (Generator, error) {
		if sr.Sub[0].Op == syntax.OpEmptyMatch {
			return gen, nil
		}

		return nil, errors.New("passit: unsupported capture")
	}
}

// SpecialCaptureWithRepeat returns a special capture factory that parses the
// capture value for a count to be used with Repeat(gen, sep, count). If the
// capture is empty, the Generator is returned directly.
func SpecialCaptureWithRepeat(gen Generator, sep string) SpecialCaptureFactory {
	return func(sr *syntax.Regexp) (Generator, error) {
		switch sr.Sub[0].Op {
		case syntax.OpEmptyMatch:
			return gen, nil
		case syntax.OpLiteral:
			count, err := strconv.ParseUint(string(sr.Sub[0].Rune), 10, bits.UintSize-1)
			if err != nil {
				return nil, fmt.Errorf("passit: failed to parse capture: %w", err)
			}

			return Repeat(gen, sep, int(count)), nil
		default:
			return nil, errors.New("passit: unsupported capture")
		}
	}
}
