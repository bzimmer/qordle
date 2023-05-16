package qordle

import (
	"errors"
	"strings"
	"unicode"

	set "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// criterion describes the state for a letter
type criterion struct {
	exact  rune
	misses set.Set[rune]
}

type criteria []criterion

func (c criteria) String() string {
	var buf strings.Builder
	for i := range c {
		switch {
		case c[i].exact != rune(0):
			buf.WriteString(string(c[i].exact))
		default:
			var bar []rune
			c[i].misses.Each(func(r rune) bool {
				bar = append(bar, r)
				return true
			})
			buf.WriteString("[^" + string(bar) + "]")
		}
	}
	return buf.String()
}

// FilterFunc performs a validation on the word
type FilterFunc func(string) bool

// parsed holds letter -> index -> mark
type parsed map[rune]map[int]Mark

var ErrInvalidFormat = errors.New("invalid pattern format")

func Filter(words Dictionary, fns ...FilterFunc) Dictionary {
	var count int
	res := make([]string, len(words))
	for _, word := range words {
		matches := true
		for i := 0; matches && i < len(fns); i++ {
			matches = fns[i](word)
		}
		if matches {
			res[count] = word
			count++
		}
	}
	return res[:count]
}

func NoOp(_ string) bool {
	return true
}

func IsLower() FilterFunc {
	return func(word string) bool {
		return unicode.IsLower(rune(word[0]))
	}
}

func Length(length int) FilterFunc {
	return func(word string) bool {
		return len(word) == length
	}
}

func filter(criteria criteria, required map[rune]int) FilterFunc {
	return func(word string) bool {
		if len(word) != len(criteria) {
			log.Debug().
				Str("word", word).
				Int("expected", len(criteria)).
				Int("found", len(word)).
				Str("reason", "length").
				Msg("filter")
			return false
		}
		ws, rs := []rune(word), make(map[rune]int)
		for i := range ws {
			rs[ws[i]]++
			if criteria[i].exact != 0 && criteria[i].exact != ws[i] {
				log.Debug().
					Str("word", word).
					Int("i", i).
					Str("expected", string(criteria[i].exact)).
					Str("found", string(ws[i])).
					Str("reason", "exact").
					Msg("filter")
				return false
			}
			if criteria[i].misses.Contains(ws[i]) {
				log.Debug().
					Str("word", word).
					Int("i", i).
					Str("found", string(ws[i])).
					Str("reason", "invalid").
					Msg("filter")
				return false
			}
		}
		for key, val := range required {
			num, ok := rs[key]
			if !ok || num < val {
				log.Debug().
					Str("word", word).
					Str("letter", string(key)).
					Int("expected", val).
					Int("found", num).
					Str("reason", "required").
					Msg("filter")
				return false
			}
		}
		return true
	}
}

func compile(marks parsed) (criteria, map[rune]int) {
	var crit criteria
	for _, states := range marks {
		for range states {
			crit = append(crit, criterion{misses: set.NewThreadUnsafeSet[rune]()})
		}
	}
	required := make(map[rune]int)
	for letter, states := range marks {
		var constrained bool
		for i, mark := range states {
			switch mark {
			case MarkExact:
				constrained = true
				required[letter]++
				// letter must appear at this index
				crit[i].exact = letter
			case MarkMisplaced:
				constrained = true
				required[letter]++
				// letter cannot appear at this index but can appear elsewhere
				crit[i].misses.Add(letter)
			case MarkMiss:
				// letter cannot appear at this index but can appear elsewhere but
				// only if unconstrained
				crit[i].misses.Add(letter)
			}
		}
		// if unconstrained, the current letter is not found in the word at any index
		for i := 0; !constrained && i < len(crit); i++ {
			crit[i].misses.Add(letter)
		}
	}
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		req := make(map[string]int, len(required))
		for key, val := range required {
			req[string(key)] = val
		}
		log.Debug().
			Str("criteria", crit.String()).
			Any("required", req).
			Msg("compile")
	}
	return crit, required
}

func parse(feedback string) (parsed, error) {
	ix, rs, marks := 0, []rune(feedback), make(parsed)
	for i := 0; i < len(rs); i++ {
		var mark Mark
		switch {
		case unicode.IsSpace(rs[i]):
			fallthrough
		case unicode.IsNumber(rs[i]):
			fallthrough
		case unicode.IsLower(rs[i]):
			mark = MarkMiss
		case unicode.IsUpper(rs[i]):
			mark = MarkExact
		default:
			i++
			switch {
			case i >= len(rs):
				log.Debug().
					Str("feedback", feedback).
					Int("i", i).
					Int("len", len(rs)).
					Str("reason", "length").
					Msg("parse")
				return nil, ErrInvalidFormat
			case unicode.IsLower(rs[i]):
				mark = MarkMisplaced
			default:
				log.Debug().
					Str("feedback", feedback).
					Int("i", i).
					Str("rune", string(rs[i])).
					Str("reason", "invalid").
					Msg("parse")
				return nil, ErrInvalidFormat
			}
		}
		lower := unicode.ToLower(rs[i])
		m, ok := marks[lower]
		if !ok {
			m = make(map[int]Mark)
			marks[lower] = m
		}
		m[ix] = mark
		ix++
	}
	return marks, nil
}

func Guess(guesses ...string) (FilterFunc, error) {
	var fns []FilterFunc
	for _, guess := range guesses {
		marks, err := parse(guess)
		if err != nil {
			return nil, err
		}
		fns = append(fns, filter(compile(marks)))
	}
	return func(word string) bool {
		for _, fn := range fns {
			if !fn(word) {
				return false
			}
		}
		return true
	}, nil
}
