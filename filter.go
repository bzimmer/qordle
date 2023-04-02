package qordle

import (
	"errors"
	"strings"
	"unicode"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type criterion struct {
	exact  rune
	misses map[rune]struct{}
}

type FilterFunc func(string) bool

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

func filter(criteria []*criterion, required map[rune]int) FilterFunc {
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
			if _, ok := criteria[i].misses[ws[i]]; ok {
				log.Debug().
					Str("word", word).
					Int("i", i).
					Str("found", string(ws[i])).
					Str("reason", "miss").
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

func compile(marks map[rune]map[int]Mark) FilterFunc { //nolint:gocognit
	var times int
	var criteria []*criterion
	for _, states := range marks {
		for range states {
			criteria = append(criteria, &criterion{misses: make(map[rune]struct{})})
		}
	}
	required := make(map[rune]int)
	for letter, states := range marks {
		/*
			if the same letter exists as a misplaced or exact mark for any
			index, then only add the miss to the current index otherwise add
			it to all indices
		*/
		var variable bool
		for index, mark := range states {
			switch mark {
			case MarkExact:
				variable = true
				required[letter]++
				criteria[index].exact = letter
			case MarkMisplaced:
				variable = true
				required[letter]++
				criteria[index].misses[letter] = struct{}{}
			case MarkMiss:
				criteria[index].misses[letter] = struct{}{}
			}
		}
		for i := 0; !variable && i < len(criteria); i++ {
			times++
			criteria[i].misses[letter] = struct{}{}
		}
	}
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		var buf strings.Builder
		for i := range criteria {
			switch {
			case criteria[i].exact != rune(0):
				buf.WriteString(string(criteria[i].exact))
			default:
				var bar []string
				for letter := range criteria[i].misses {
					bar = append(bar, string(letter))
				}
				buf.WriteString("[^" + strings.Join(bar, "") + "]")
			}
		}
		log.Debug().
			Str("pattern", buf.String()).
			Any("required", required).
			Int("times", times).
			Msg("compile")
	}
	return filter(criteria, required)
}

func parse(feedback string) (FilterFunc, error) {
	ix, rs := 0, []rune(feedback)
	marks := make(map[rune]map[int]Mark)
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
					Msg("filter")
				return nil, ErrInvalidFormat
			case unicode.IsLower(rs[i]):
				mark = MarkMisplaced
			default:
				log.Debug().
					Str("feedback", feedback).
					Int("i", i).
					Str("rune", string(rs[i])).
					Str("reason", "invalid").
					Msg("filter")
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
	return compile(marks), nil
}

func Guess(guesses ...string) (FilterFunc, error) {
	var fns []FilterFunc
	for _, guess := range guesses {
		if guess == "" {
			continue
		}
		ff, err := parse(guess)
		if err != nil {
			return nil, err
		}
		fns = append(fns, ff)
	}
	if len(fns) == 0 {
		return NoOp, nil
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
