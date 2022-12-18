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

const (
	maxUnboundedRepeatCount = 15

	// questNoChance is the likelihood that Z? will output nothing.
	// The fraction should be reduced.
	questNoChanceNumerator   = 1
	questNoChanceDenominator = 2
)

type regexpGenerator func(*strings.Builder, io.Reader) error

func (rg regexpGenerator) Password(r io.Reader) (string, error) {
	var b strings.Builder
	if err := rg(&b, r); err != nil {
		return "", err
	}

	return b.String(), nil
}

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
//
// It is an error to use named captures (?P<name>) except to refer to special
// capture factories added with SetSpecialCapture.
func (p *RegexpParser) Parse(pattern string, flags syntax.Flags) (Generator, error) {
	// Note: The FoldCase, OneLine, DotNL and NonGreedy flags can be set or
	//   cleared within the pattern.

	// The Literal flag is odd in that it can interact with FoldCase in ways
	// that may be invalid. syntax.literalRegexp doesn't call syntax.minFoldRune
	// so the resulting OpLiteral won't properly be the minimum folded runes.
	// That means foldedLiteral will create an invalid character class.
	if flags&(syntax.Literal|syntax.FoldCase) == syntax.Literal|syntax.FoldCase {
		return nil, errors.New("passit: Literal flag is unsupported when used with FoldCase")
	}

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

	return p.parse(r)
}

func (p *RegexpParser) parse(r *syntax.Regexp) (regexpGenerator, error) {
	var (
		gen regexpGenerator
		err error
	)
	switch r.Op {
	case syntax.OpEmptyMatch:
		// This is handled by onlyEmptyOutput.
	case syntax.OpLiteral:
		gen, err = p.literal(r)
	case syntax.OpCharClass:
		gen, err = p.charClass(r)
	case syntax.OpAnyCharNotNL:
		gen, err = p.anyCharNotNL(r)
	case syntax.OpAnyChar:
		gen, err = p.anyChar(r)
	case syntax.OpBeginLine, syntax.OpEndLine,
		syntax.OpBeginText, syntax.OpEndText,
		syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		// This is handled by onlyEmptyOutput.
	case syntax.OpCapture:
		gen, err = p.capture(r)
	case syntax.OpStar:
		gen, err = p.star(r)
	case syntax.OpPlus:
		gen, err = p.plus(r)
	case syntax.OpQuest:
		gen, err = p.quest(r)
	case syntax.OpRepeat:
		gen, err = p.repeat(r)
	case syntax.OpConcat:
		gen, err = p.concat(r)
	case syntax.OpAlternate:
		gen, err = p.alternate(r)
	default:
		err = fmt.Errorf("passit: invalid regexp %q, unhandled op %s", r, r.Op)
	}
	if err != nil {
		return nil, err
	}

	// Check onlyEmptyOutput after we've parsed the syntax.Regexp to ensure we
	// surface any errors.
	if onlyEmptyOutput(r) {
		return func(*strings.Builder, io.Reader) error {
			return nil
		}, nil
	}

	if gen == nil {
		panic("passit: internal error: gen is nil after onlyEmptyOutput")
	}
	return gen, nil
}

func (p *RegexpParser) literal(sr *syntax.Regexp) (regexpGenerator, error) {
	// FoldCase is the only flag relevant here.
	if sr.Flags&syntax.FoldCase != 0 {
		return p.foldedLiteral(sr)
	}

	return rawLiteral(sr.Rune), nil
}

func rawLiteral(runes []rune) regexpGenerator {
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
			gens = append(gens, rawLiteral(sr.Rune[litStart:i+1]))
			litStart = -1
		}

		gen, err := p.foldedRune(c)
		if err != nil {
			return nil, err
		}
		gens = append(gens, gen)
	}

	if litStart >= 0 {
		gens = append(gens, rawLiteral(sr.Rune[litStart:]))
	}

	if len(gens) == 1 {
		return gens[0], nil
	}

	return concatGenerators(gens), nil
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
	return charClassGenerator(sr, &tab)
}

func (p *RegexpParser) anyCharNotNL(sr *syntax.Regexp) (regexpGenerator, error) {
	return charClassGenerator(sr, p.anyRangeTableNoNL())
}

func (p *RegexpParser) anyChar(sr *syntax.Regexp) (regexpGenerator, error) {
	return charClassGenerator(sr, p.anyRangeTable())
}

func charClassGenerator(sr *syntax.Regexp, tab *unicode.RangeTable) (regexpGenerator, error) {
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
		return nil, errors.New("passit: named capture refers to unknown special capture factory")
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
	// NonGreedy, which we ignore, is the only relevant flag here.
	sr.Min, sr.Max = 0, -1
	return p.repeat(sr)
}

func (p *RegexpParser) plus(sr *syntax.Regexp) (regexpGenerator, error) {
	// NonGreedy, which we ignore, is the only relevant flag here.
	sr.Min, sr.Max = 1, -1
	return p.repeat(sr)
}

func (p *RegexpParser) quest(sr *syntax.Regexp) (regexpGenerator, error) {
	// NonGreedy, which we ignore, is the only relevant flag here.

	gen, err := p.parse(sr.Sub[0])
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		n, err := readIntN(r, questNoChanceDenominator)
		if err != nil {
			return err
		}
		if n < questNoChanceNumerator {
			return nil
		}
		return gen(b, r)
	}, nil
}

func (p *RegexpParser) repeat(sr *syntax.Regexp) (regexpGenerator, error) {
	// NonGreedy, which we ignore, is the only relevant flag here.

	min := sr.Min
	max := sr.Max
	if max == -1 {
		max = sr.Min + maxUnboundedRepeatCount
	}

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
	gens := make([]regexpGenerator, 0, len(sr.Sub))
	for _, r := range sr.Sub {
		gen, err := p.parse(r)
		if err != nil {
			return nil, err
		}

		// Skip past empty generators.
		if !onlyEmptyOutput(r) {
			gens = append(gens, gen)
		}
	}

	return concatGenerators(gens), nil
}

func concatGenerators(gens []regexpGenerator) regexpGenerator {
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
	gens := make([]regexpGenerator, len(sr.Sub))
	for i, r := range sr.Sub {
		gen, err := p.parse(r)
		if err != nil {
			return nil, err
		}

		// We don't skip empty generators here as they change the behaviour
		// of the generator.
		gens[i] = gen
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

func onlyEmptyOutput(sr *syntax.Regexp) bool {
	switch sr.Op {
	case syntax.OpEmptyMatch:
		return true
	case syntax.OpCapture:
		if sr.Name != "" {
			return false
		}
		// fallout
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest:
		// fallout
	case syntax.OpRepeat:
		if sr.Min == 0 && sr.Max == 0 {
			return true
		}
		// fallout
	case syntax.OpConcat, syntax.OpAlternate:
		// fallout
	case syntax.OpBeginLine, syntax.OpEndLine,
		syntax.OpBeginText, syntax.OpEndText,
		syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return true
	default:
		return false
	}

	for _, sub := range sr.Sub {
		if !onlyEmptyOutput(sub) {
			return false
		}
	}

	return true
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
