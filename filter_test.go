package qordle_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/qordle"
)

type FilterFuncFunc func(string) qordle.FilterFunc

func mustGuesses(guesses ...string) qordle.FilterFunc {
	f, err := qordle.Guesses(guesses...)
	if err != nil {
		panic(err)
	}
	return f
}

func TestPattern(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, word, pattern string
		result, err         bool
	}{
		{
			name:    "no pattern",
			pattern: "",
			word:    "foo",
			result:  true,
		},
		{
			name:    "match one letter not in word",
			pattern: "..a..",
			word:    "foo",
			result:  false,
		},
		{
			name:    "match one letter",
			pattern: "..a..",
			word:    "hoard",
			result:  true,
		},
		{
			name:    "invalid regex",
			pattern: "[a-z",
			word:    "hoard",
			result:  false,
			err:     true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			p, err := qordle.Pattern(tt.pattern)
			if tt.err {
				a.Error(err)
				a.Nil(p)
			} else {
				a.NoError(err)
				a.Equal(tt.result, p(tt.word))
			}
		})
	}
}

func TestHitsAndMisses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, input, word string
		result            bool
		ff                FilterFuncFunc
	}{
		{
			name:   "no hits",
			input:  "ab",
			word:   "foo",
			result: false,
			ff:     qordle.Hits,
		},
		{
			name:   "one hit",
			input:  "ab",
			word:   "banana",
			result: true,
			ff:     qordle.Hits,
		},
		{
			name:   "misses",
			input:  "ab",
			word:   "banana",
			result: false,
			ff:     qordle.Misses,
		},
		{
			name:   "multiple hits of same letter",
			input:  "cee",
			word:   "fleck",
			result: false,
			ff:     qordle.Hits,
		},
		{
			name:   "misses empty",
			input:  "",
			word:   "banana",
			result: true,
			ff:     qordle.Misses,
		},
		{
			name:   "hits empty",
			input:  "",
			word:   "banana",
			result: true,
			ff:     qordle.Hits,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			p := tt.ff(tt.input)
			a.Equal(tt.result, p(tt.word))
		})
	}
}

func TestLength(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, word string
		result     bool
		length     int
	}{
		{
			name:   "five",
			length: 5,
			word:   "hoody",
			result: true,
		},
		{
			name:   "three",
			length: 3,
			word:   "hoody",
			result: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			p := qordle.Length(tt.length)
			a.Equal(tt.result, p(tt.word))
		})
	}
}

func TestNoOp(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, word string
		result     bool
	}{
		{
			name: "word",
			word: "hoody",
		},
		{
			name: "empty",
			word: "",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			a.True(qordle.NoOp(tt.word))
		})
	}
}

func TestLower(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, word string
		result     bool
	}{
		{
			name:   "lower",
			word:   "hoody",
			result: true,
		},
		{
			name:   "upper",
			word:   "Hoody",
			result: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			p := qordle.IsLower()
			a.Equal(tt.result, p(tt.word))
		})
	}
}

func TestGuesses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, word  string
		guesses     []string
		err, result bool
	}{
		{
			name:    "no guesses",
			word:    "hoody",
			guesses: []string{""},
			result:  true,
		},
		{
			name:    "guesses with match",
			word:    "hoody",
			guesses: []string{"s~hOut"},
			result:  true,
		},
		{
			name:    "guesses with no match",
			word:    "hoody",
			guesses: []string{"cloud"},
			result:  false,
		},
		{
			name:    "guesses with one match out of order",
			word:    "dusty",
			guesses: []string{"brain", "clov~e"},
			result:  false,
		},
		{
			name:    "guess with numbers",
			word:    "dusty",
			guesses: []string{"12345"},
			result:  true,
		},
		{
			name:    "guesses with one match out of order but also legal",
			word:    "pleat",
			guesses: []string{"br~ain", "~l~egAl"},
			result:  true,
		},
		{
			name:    "guesses with partial as hash",
			word:    "pleat",
			guesses: []string{"br#ain", "#l#egAl"},
			result:  true,
		},
		{
			name:    "guesses with partial as hash and capital partial",
			word:    "pleat",
			guesses: []string{"br#ain", "#l#EgAl"},
			result:  true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			p, err := qordle.Guesses(tt.guesses...)
			switch tt.err {
			case true:
				a.Error(err)
			case false:
				a.NoError(err)
			}
			a.Equal(tt.result, p(tt.word))
		})
	}
}

func FuzzGuesses(f *testing.F) {
	for _, x := range []string{"br#ain", "#l#EgAl", "foo", "start", "12345", "rüsch"} {
		f.Add(x)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = qordle.Guesses(s)
	})
}

func TestFilter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		words, result qordle.Dictionary
		fns           []qordle.FilterFunc
	}{
		{
			name:   "filter",
			words:  qordle.Dictionary{"hoody", "foobar"},
			result: qordle.Dictionary{"hoody"},
			fns:    []qordle.FilterFunc{qordle.Length(5)},
		},
		{
			name:   "double letter",
			words:  qordle.Dictionary{"excel", "fleck", "expel", "sport"},
			result: qordle.Dictionary{"excel", "expel"},
			fns:    []qordle.FilterFunc{mustGuesses("brain", "south", "@l@edg@e")},
		},
		{
			name:   "filter all",
			words:  qordle.Dictionary{"hoody", "foobar"},
			result: qordle.Dictionary{"hoody", "foobar"},
			fns:    []qordle.FilterFunc{},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			dictionary := qordle.Filter(tt.words, tt.fns...)
			a.Equal(tt.result, dictionary)
		})
	}
}
