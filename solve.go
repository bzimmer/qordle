package qordle

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

type FilterFunc func(string) bool

func Solve(words Dictionary, fns ...FilterFunc) Dictionary {
	var res []string
	for _, word := range words {
		matches := true
		for i := 0; matches && i < len(fns); i++ {
			matches = fns[i](word)
		}
		if matches {
			res = append(res, word)
		}
	}
	return res
}

func Lower() FilterFunc {
	return func(word string) bool {
		return unicode.IsLower(rune(word[0]))
	}
}

func Length(length int) FilterFunc {
	return func(word string) bool {
		return len(word) == length
	}
}

func Misses(misses string) FilterFunc {
	return func(word string) bool {
		return !strings.ContainsAny(word, misses)
	}
}

func Hits(hits string) FilterFunc {
	return func(word string) bool {
		for i := range hits {
			if !strings.Contains(word, string(hits[i])) {
				return false
			}
		}
		return true
	}
}

func Pattern(pattern string) (FilterFunc, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(word string) bool {
		return re.MatchString(word)
	}, nil
}

func contains(m []string, s string) bool {
	for _, x := range m {
		if x == s {
			return true
		}
	}
	return false
}

func join(hits []string, misses map[int]string, idx int) string {
	var s string
	for key, val := range misses {
		switch key {
		case idx:
			s += val
		default:
			if !contains(hits, val) {
				s += val
			}
		}
	}
	return s
}

func Guesses(guesses ...string) (FilterFunc, error) {
	var fns []FilterFunc
	for _, guess := range guesses {
		x := []rune(guess)
		var hits, pattern []string
		misses := make(map[int]string, 0)
		for i := 0; i < len(x); i++ {
			switch {
			case x[i] == '~':
				i++
				hits = append(hits, string(x[i]))
				misses[len(pattern)] = string(x[i])
				pattern = append(pattern, "[^")
			case unicode.IsUpper(x[i]):
				hit := string(unicode.ToLower(x[i]))
				hits = append(hits, hit)
				pattern = append(pattern, hit)
			case unicode.IsLower(x[i]):
				misses[len(pattern)] = string(x[i])
				pattern = append(pattern, "")
			}
		}

		var re string
		for i, s := range pattern {
			switch {
			case s == "":
				re += fmt.Sprintf("[^%s]", join(hits, misses, i))
			case s[0] == '[':
				re += fmt.Sprintf("%s%s]", s, join(hits, misses, i))
			default:
				re += s
			}
		}
		log.Debug().
			Strs("hits", hits).
			Interface("misses", misses).
			Str("guess", guess).
			Str("pattern", re).
			Msg("guesses")
		p, err := Pattern(re)
		if err != nil {
			return nil, err
		}
		fns = append(fns, Hits(strings.Join(hits, "")), p)
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
