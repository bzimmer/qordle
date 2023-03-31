package qordle_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/qordle"
)

type FilterFuncFunc func(string) qordle.FilterFunc

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
		name, word string
		guesses    []string
		result     bool
		err        error
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
			guesses: []string{"s.hOut"},
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
			guesses: []string{"brain", "clov.e"},
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
			guesses: []string{"br.ain", ".l.egAl"},
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
		{
			name:    "error with poor format",
			word:    "pleat",
			guesses: []string{"br.ain", ".l.Eg...A....."},
			result:  true,
			err:     qordle.ErrInvalidFormat,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			ff, err := qordle.Guess(tt.guesses...)
			if tt.err != nil {
				a.ErrorIs(err, tt.err)
				return
			}
			a.Equal(tt.result, ff(tt.word))
		})
	}
}

func FuzzGuesses(f *testing.F) {
	for _, x := range []string{"br#ain", "#l#EgAl", "foo", "start", "12345", "rüsch"} {
		f.Add(x)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, err := qordle.Guess(s)
		if err != nil && !errors.Is(err, qordle.ErrInvalidFormat) {
			panic(err)
		}
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
			fns: func() []qordle.FilterFunc {
				ff, err := qordle.Guess("brain", "south", "@l@edg@e")
				if err != nil {
					panic(err)
				}
				return []qordle.FilterFunc{ff}
			}()},
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
