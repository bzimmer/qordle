package qordle

import (
	"regexp"
	"strings"
	"unicode"
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

func Begins(prefix string) FilterFunc {
	return func(word string) bool {
		return strings.HasPrefix(word, prefix)
	}
}

func Ends(suffix string) FilterFunc {
	return func(word string) bool {
		return strings.HasSuffix(word, suffix)
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
