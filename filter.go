package qordle

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type FilterFunc func(string) bool

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

func Misses(misses string) FilterFunc {
	if misses == "" {
		return NoOp
	}
	return func(word string) bool {
		return !strings.ContainsAny(word, misses)
	}
}

func Hits(hits string) FilterFunc {
	if hits == "" {
		return NoOp
	}
	m := map[rune]int{}
	for i := range hits {
		m[rune(hits[i])]++
	}
	return func(word string) bool {
		n := map[rune]int{}
		for i := range word {
			n[rune(word[i])]++
		}
		for key, val := range m {
			if n[key] < val {
				return false
			}
		}
		return true
	}
}

func Pattern(pattern string) (FilterFunc, error) {
	if pattern == "" {
		return NoOp, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	fn := func(word string) bool {
		return re.MatchString(word)
	}
	return fn, nil
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

func parse(guess string) ([]FilterFunc, error) {
	x := []rune(guess)
	var hits, pattern []string
	misses := make(map[int]string, 0)
	for i := 0; i < len(x); i++ {
		switch {
		case unicode.IsSpace(x[i]):
			fallthrough
		case unicode.IsNumber(x[i]):
			fallthrough
		case unicode.IsLower(x[i]):
			misses[len(pattern)] = string(x[i])
			pattern = append(pattern, "")
		case unicode.IsUpper(x[i]):
			hit := string(unicode.ToLower(x[i]))
			hits = append(hits, hit)
			pattern = append(pattern, hit)
		default:
			i++
			if i >= len(x) {
				return nil, errors.New("too few characters")
			}
			w := string(unicode.ToLower(x[i]))
			hits = append(hits, w)
			misses[len(pattern)] = w
			pattern = append(pattern, "[^")
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
	p, err := Pattern(re)
	if err != nil {
		return nil, err
	}
	return []FilterFunc{Hits(strings.Join(hits, "")), p}, nil
}

func Guesses(guesses ...string) (FilterFunc, error) {
	var fns []FilterFunc
	for _, guess := range guesses {
		if guess == "" {
			continue
		}
		f, err := parse(guess)
		if err != nil {
			return nil, err
		}
		fns = append(fns, f...)
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
