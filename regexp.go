package password

import (
	"fmt"
	"io"
	"regexp/syntax"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

const maxUnboundedRepeatCount = 15

const RegexpUnicodeAny syntax.Flags = regexpUnicodeAny
const regexpUnicodeAny = syntax.Simple << 1

type regexpGenerator func(*strings.Builder, io.Reader) error

// regexpFactories is initialised in func init to prevent an initialization loop.
var regexpFactories map[syntax.Op]func(*syntax.Regexp) (regexpGenerator, error)

func init() {
	regexpFactories = map[syntax.Op]func(*syntax.Regexp) (regexpGenerator, error){
		//syntax.OpNoMatch:      regexpNotImplemented,
		syntax.OpEmptyMatch:     regexpNoop,
		syntax.OpLiteral:        regexpLiteral,
		syntax.OpCharClass:      regexpCharClass,
		syntax.OpAnyCharNotNL:   regexpAnyCharNotNL,
		syntax.OpAnyChar:        regexpAnyCharNotNL,
		syntax.OpBeginLine:      regexpNoop,
		syntax.OpEndLine:        regexpNoop,
		syntax.OpBeginText:      regexpNoop,
		syntax.OpEndText:        regexpNoop,
		syntax.OpWordBoundary:   regexpNoop,
		syntax.OpNoWordBoundary: regexpNoop,
		syntax.OpCapture:        regexpCapture,
		syntax.OpStar:           regexpStar,
		syntax.OpPlus:           regexpPlus,
		syntax.OpQuest:          regexpQuest,
		syntax.OpRepeat:         regexpRepeat,
		syntax.OpConcat:         regexpConcat,
		syntax.OpAlternate:      regexpAlternate,
	}
}

type regexpTemplate struct{ gen regexpGenerator }

func ParseRegexp(pattern string, flags syntax.Flags) (Template, error) {
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

	gen, err := newRegexpGenerator(r)
	if err != nil {
		return nil, err
	}

	return &regexpTemplate{gen}, nil
}

func (rt *regexpTemplate) Password(r io.Reader) (string, error) {
	var b strings.Builder
	if err := rt.gen(&b, r); err != nil {
		return "", err
	}

	return b.String(), nil
}

func newRegexpGenerator(r *syntax.Regexp) (regexpGenerator, error) {
	factory, ok := regexpFactories[r.Op]
	if !ok {
		return nil, fmt.Errorf("strongroom/password: invalid regexp %q, unhandled op %s", r, r.Op)
	}

	return factory(r)
}

func newRegexpGenerators(rs []*syntax.Regexp) ([]regexpGenerator, error) {
	gens := make([]regexpGenerator, len(rs))
	for i, r := range rs {
		gen, err := newRegexpGenerator(r)
		if err != nil {
			return nil, err
		}

		gens[i] = gen
	}

	return gens, nil
}

func regexpNoop(*syntax.Regexp) (regexpGenerator, error) {
	return func(*strings.Builder, io.Reader) error {
		return nil
	}, nil
}

func regexpLiteral(sr *syntax.Regexp) (regexpGenerator, error) {
	s := string(sr.Rune)
	return func(b *strings.Builder, r io.Reader) error {
		b.WriteString(s)
		return nil
	}, nil
}

func regexpCharClass(sr *syntax.Regexp) (regexpGenerator, error) {
	const maxR16 = 1<<16 - 1

	tab := new(unicode.RangeTable)
	for i := 0; i < len(sr.Rune); i += 2 {
		start, end := sr.Rune[i], sr.Rune[i+1]

		if start > maxR16 {
			tab.R32 = append(tab.R32, unicode.Range32{
				Lo:     uint32(start),
				Hi:     uint32(end),
				Stride: 1,
			})
			continue
		}

		if end > maxR16 {
			tab.R32 = append(tab.R32, unicode.Range32{
				Lo:     maxR16 + 1,
				Hi:     uint32(end),
				Stride: 1,
			})
			end = maxR16
		}

		tab.R16 = append(tab.R16, unicode.Range16{
			Lo:     uint16(start),
			Hi:     uint16(end),
			Stride: 1,
		})

		if end <= unicode.MaxLatin1 {
			tab.LatinOffset++
		}
	}

	// If a character class contains both 0 and MaxRune, it's probably a negated
	// class. There is no way to directly test for this.
	negated := len(sr.Rune) > 0 && sr.Rune[0] == 0 &&
		sr.Rune[len(sr.Rune)-1] == unicode.MaxRune
	if negated {
		tab = intersectRangeTables(tab, regexpAnyRangeTable(sr.Flags))
	}

	return regexpCharClassInternal(sr, tab)
}

func regexpAnyCharNotNL(sr *syntax.Regexp) (regexpGenerator, error) {
	return regexpCharClassInternal(sr, regexpAnyRangeTable(sr.Flags))
}

func regexpCharClassInternal(sr *syntax.Regexp, tab *unicode.RangeTable) (regexpGenerator, error) {
	count := countTableRunes(tab)
	if count == 0 {
		return nil, fmt.Errorf("strongroom/password: character class %s contains zero runes", sr)
	}

	return func(b *strings.Builder, r io.Reader) error {
		v, err := readRune(r, tab, count)
		b.WriteRune(v)
		return err
	}, nil
}

func regexpCapture(sr *syntax.Regexp) (regexpGenerator, error) {
	return newRegexpGenerator(sr.Sub[0])
}

func regexpStar(sr *syntax.Regexp) (regexpGenerator, error) {
	return regexpRepeatInternal(sr, 0, maxUnboundedRepeatCount)
}

func regexpPlus(sr *syntax.Regexp) (regexpGenerator, error) {
	// We use maxUnboundedRepeatCount+1 here so that x{1,} and x+ are identical,
	// x{0,} and x* are already identical.
	//
	// TODO(tmthrgd): Is this the behaviour we want?
	return regexpRepeatInternal(sr, 1, maxUnboundedRepeatCount+1)
}

func regexpQuest(sr *syntax.Regexp) (regexpGenerator, error) {
	return regexpRepeatInternal(sr, 0, 1)
}

func regexpRepeat(sr *syntax.Regexp) (regexpGenerator, error) {
	max := sr.Max
	if max == -1 {
		max = sr.Min + maxUnboundedRepeatCount
	}

	return regexpRepeatInternal(sr, sr.Min, max)
}

func regexpRepeatInternal(sr *syntax.Regexp, min, max int) (regexpGenerator, error) {
	gen, err := newRegexpGenerator(sr.Sub[0])
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		n, err := readUint32n(r, uint32(max-min+1))
		if err != nil {
			return err
		}

		for n += uint32(min); n > 0; n-- {
			if err := gen(b, r); err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func regexpConcat(sr *syntax.Regexp) (regexpGenerator, error) {
	gens, err := newRegexpGenerators(sr.Sub)
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

func regexpAlternate(sr *syntax.Regexp) (regexpGenerator, error) {
	gens, err := newRegexpGenerators(sr.Sub)
	if err != nil {
		return nil, err
	}

	return func(b *strings.Builder, r io.Reader) error {
		idx, err := readUint32n(r, uint32(len(gens)))
		if err != nil {
			return err
		}

		return gens[idx](b, r)
	}, nil
}

var regexpAnyRangeTableASCII = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x20, Hi: 0x7e, Stride: 1},
	},
	LatinOffset: 1,
}

var regexpAnyRangeTableUni struct {
	tab *unicode.RangeTable
	sync.Once
}

func regexpAnyRangeTable(flags syntax.Flags) *unicode.RangeTable {
	if flags&RegexpUnicodeAny == 0 {
		return regexpAnyRangeTableASCII
	}

	regexpAnyRangeTableUni.Do(func() {
		regexpAnyRangeTableUni.tab = unstridifyRangeTable(rangetable.Merge(
			// TODO(tmthrgd): Review these ranges.
			unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lo,
			unicode.Mn,
			unicode.N,
			unicode.P,
			unicode.S,
		))
	})
	return regexpAnyRangeTableUni.tab
}

func unstridifyRangeTable(tab *unicode.RangeTable) *unicode.RangeTable {
	for i := 0; i < len(tab.R16); i++ {
		if r16 := tab.R16[i]; r16.Stride != 1 {
			size := int((r16.Hi-r16.Lo)/r16.Stride) + 1
			tab.R16 = append(tab.R16, make([]unicode.Range16, size-1)...)
			copy(tab.R16[i+size:], tab.R16[i+1:])

			for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
				tab.R16[i] = unicode.Range16{Lo: uint16(r), Hi: uint16(r), Stride: 1}
				i++
			}
			i--
		}
	}

	for i := 0; i < len(tab.R32); i++ {
		if r32 := tab.R32[i]; r32.Stride != 1 {
			size := int((r32.Hi-r32.Lo)/r32.Stride) + 1
			tab.R32 = append(tab.R32, make([]unicode.Range32, size-1)...)
			copy(tab.R32[i+size:], tab.R32[i+1:])

			for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
				tab.R32[i] = unicode.Range32{Lo: uint32(r), Hi: uint32(r), Stride: 1}
				i++
			}
			i--
		}
	}

	return tab
}

func intersectRangeTables(a, b *unicode.RangeTable) *unicode.RangeTable {
	var rt unicode.RangeTable

	/*if r0.Stride|r1.Stride != 1 {
		panic("strongroom/password: unicode.RangeTable has entry with Stride > 1")
	}*/

	for _, r0 := range a.R16 {
		for _, r1 := range b.R16 {
			if !anyOverlap(rune(r0.Lo), rune(r0.Hi), rune(r1.Lo), rune(r1.Hi)) {
				continue
			}

			lo, hi := r0.Lo, r0.Hi
			if lo < r1.Lo {
				lo = r1.Lo
			}
			if hi > r1.Hi {
				hi = r1.Hi
			}

			if hi <= unicode.MaxLatin1 {
				rt.LatinOffset++
			}

			rt.R16 = append(rt.R16, unicode.Range16{Lo: lo, Hi: hi, Stride: 1})
			break
		}
	}

	for _, r0 := range a.R32 {
		for _, r1 := range b.R32 {
			if !anyOverlap(rune(r0.Lo), rune(r0.Hi), rune(r1.Lo), rune(r1.Hi)) {
				continue
			}

			lo, hi := r0.Lo, r0.Hi
			if lo < r1.Lo {
				lo = r1.Lo
			}
			if hi > r1.Hi {
				hi = r1.Hi
			}

			rt.R32 = append(rt.R32, unicode.Range32{Lo: lo, Hi: hi, Stride: 1})
			break
		}
	}

	return &rt
}

func anyOverlap(aLo, aHi, bLo, bHi rune) bool {
	return aLo <= bHi && bLo <= aHi
}
