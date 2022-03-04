package qordle

import (
	"regexp"
	"strings"
	"unicode"
)

type SolveFunc func(string) bool

func Solve(words Dictionary, fns ...SolveFunc) Dictionary {
	dict := new(dictionary)
	for _, word := range words.Words() {
		matches := true
		for i := range fns {
			matches = matches && fns[i](word)
		}
		if matches {
			dict.words = append(dict.words, word)
		}
	}
	return dict
}

func Lower() SolveFunc {
	return func(word string) bool {
		return unicode.IsLower(rune(word[0]))
	}
}

func Length(length int) SolveFunc {
	return func(word string) bool {
		return len(word) == length
	}
}

func Begins(prefix string) SolveFunc {
	return func(word string) bool {
		return strings.HasPrefix(word, prefix)
	}
}

func Ends(suffix string) SolveFunc {
	return func(word string) bool {
		return strings.HasSuffix(word, suffix)
	}
}

func Misses(misses string) SolveFunc {
	return func(word string) bool {
		return !strings.ContainsAny(word, misses)
	}
}

func Hits(hits string) SolveFunc {
	return func(word string) bool {
		for i := range hits {
			if !strings.Contains(word, string(hits[i])) {
				return false
			}
		}
		return true
	}
}

func Pattern(pattern string) SolveFunc {
	re := regexp.MustCompile(pattern)
	return func(word string) bool {
		return re.MatchString(word)
	}
}
