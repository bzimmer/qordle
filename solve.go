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
	dict := new(dictionary)
	for _, word := range words.Words() {
		matches := true
		for i := 0; matches && i < len(fns); i++ {
			matches = fns[i](word)
		}
		if matches {
			dict.words = append(dict.words, word)
		}
	}
	return dict
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

func Guesses(guesses ...string) (FilterFunc, error) {
	var fns []FilterFunc
	for _, guess := range guesses {
		x := []rune(guess)
		var misses string
		var pattern []string
		for i := 0; i < len(x); i++ {
			switch {
			case x[i] == '~':
				i++
				pattern = append(pattern, fmt.Sprintf("[^%c", x[i]))
			case unicode.IsUpper(x[i]):
				pattern = append(pattern, string(unicode.ToLower(x[i])))
			case unicode.IsLower(x[i]):
				misses += string(x[i])
				pattern = append(pattern, "")
			}
		}
		var re string
		for _, s := range pattern {
			switch {
			case s == "":
				re += fmt.Sprintf("[^%s]", misses)
			case s[0] == '[':
				re += fmt.Sprintf("%s%s]", s, misses)
			default:
				re += s
			}
		}
		log.Debug().Str("guess", guess).Str("pattern", re).Msg("guesses")
		p, err := Pattern(re)
		if err != nil {
			return nil, err
		}
		fns = append(fns, p)
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
