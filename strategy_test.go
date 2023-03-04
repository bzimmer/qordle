package qordle_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/qordle"
)

func TestAlpha(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		words, result qordle.Dictionary
	}{
		{
			name:   "simple",
			words:  qordle.Dictionary{"easle", "fause", "false", "haste", "halse"},
			result: qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
		},
		{
			name:   "empty",
			words:  qordle.Dictionary{},
			result: qordle.Dictionary{},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := new(qordle.Alpha)
			dictionary := s.Apply(tt.words)
			a.Equal(tt.result, dictionary)
			a.Equal("alpha", s.String())
		})
	}
}

func TestBigram(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		words, result qordle.Dictionary
	}{
		{
			name:   "simple",
			words:  qordle.Dictionary{"easle", "fause", "false", "haste", "halse"},
			result: qordle.Dictionary{"haste", "halse", "easle", "false", "fause"},
		},
		{
			name:   "one letter word",
			words:  qordle.Dictionary{"a", "b"},
			result: qordle.Dictionary{"a", "b"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := new(qordle.Bigram)
			dictionary := s.Apply(tt.words)
			a.Equal(tt.result, dictionary)
			a.Equal("bigram", s.String())
		})
	}
}

func TestChain(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		strategies    []qordle.Strategy
		words, result qordle.Dictionary
	}{
		{
			name:   "no strategies",
			words:  qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result: qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
		},
		{
			name:       "one strategy",
			strategies: []qordle.Strategy{new(qordle.Alpha)},
			words:      qordle.Dictionary{"maths", "sport", "brain", "raise"},
			result:     qordle.Dictionary{"brain", "maths", "raise", "sport"},
		},
		{
			name:   "empty words",
			words:  qordle.Dictionary{},
			result: qordle.Dictionary{},
		},
		{
			name:       "position and frequency",
			strategies: []qordle.Strategy{new(qordle.Position), new(qordle.Frequency)},
			words:      qordle.Dictionary{"maths", "sport", "brain", "raise"},
			result:     qordle.Dictionary{"raise", "maths", "brain", "sport"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := qordle.NewChain(tt.strategies...)
			dictionary := s.Apply(tt.words)
			a.Equal(tt.result, dictionary)
			a.Contains(s.String(), "chain")
		})
	}
}

func TestFrequency(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		words, result qordle.Dictionary
	}{
		{
			name:   "frequency one",
			words:  qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result: qordle.Dictionary{"false", "halse", "easle", "fause", "haste"},
		},
		{
			name:   "frequency two",
			words:  qordle.Dictionary{"maths", "sport", "brain", "raise"},
			result: qordle.Dictionary{"raise", "brain", "maths", "sport"},
		},
		{
			name:   "empty",
			words:  qordle.Dictionary{},
			result: qordle.Dictionary{},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := new(qordle.Frequency)
			dictionary := s.Apply(tt.words)
			a.Equal(tt.result, dictionary)
			a.Equal("frequency", s.String())
		})
	}
}

func TestPosition(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		words, result qordle.Dictionary
	}{
		{
			name: "position",
			words: qordle.Dictionary{
				"irate",
				"raise",
				"cater",
				"slate",
				"stale",
				"steal",
			},
			result: []string{"slate", "stale", "irate", "raise", "steal", "cater"},
		},
		{
			name:   "empty",
			words:  qordle.Dictionary{},
			result: qordle.Dictionary{},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := new(qordle.Position)
			dictionary := s.Apply(tt.words)
			a.Equal(tt.result, dictionary)
			a.Equal("position", s.String())
		})
	}
}

type identity struct{}

func (s *identity) String() string {
	return "identity"
}

func (s *identity) Apply(words qordle.Dictionary) qordle.Dictionary {
	return words
}

func TestSpeculate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                 string
		strategy             qordle.Strategy
		words, result, round qordle.Dictionary
	}{
		{
			name:     "no guessing game",
			words:    qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			round:    qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result:   qordle.Dictionary{"false", "halse", "easle", "fause", "haste"},
			strategy: new(qordle.Frequency),
		},
		{
			name: "guessing",
			words: qordle.Dictionary{
				"gyppy", "ghyll", "hyphy", "glyph", "layer"},
			result: qordle.Dictionary{
				"glyph", "fears", "gears", "hears",
				"lears", "pears", "wears", "years", "sears"},
			round: qordle.Dictionary{
				"fears", "gears", "hears", "lears",
				"pears", "wears", "years", "sears"},
			strategy: new(qordle.Frequency),
		},
		{
			name:     "single letter changes",
			words:    qordle.Dictionary{"bcdef", "defab"},
			result:   qordle.Dictionary{"abcde", "abcee", "bbcee", "bccee", "bcdee"},
			round:    qordle.Dictionary{"abcde", "abcee", "bbcee", "bccee", "bcdee"},
			strategy: new(identity),
		},
		{
			name:     "one word",
			words:    qordle.Dictionary{"layer"},
			result:   qordle.Dictionary{"layer"},
			round:    qordle.Dictionary{"layer"},
			strategy: new(qordle.Frequency),
		},
		{
			name:     "empty",
			words:    qordle.Dictionary{},
			result:   qordle.Dictionary{},
			round:    qordle.Dictionary{},
			strategy: new(qordle.Frequency),
		},
		{
			name:     "no strategy",
			words:    qordle.Dictionary{"branch", "brain", "soare"},
			round:    qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result:   qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			strategy: nil,
		},
		{
			name: "main dictionary has words of different lengths",
			// the second word would be the ideal guess but it's the wrong length
			words: qordle.Dictionary{"glyph", "glyphf"},
			result: qordle.Dictionary{
				"glyph", "fears", "gears", "hears",
				"lears", "pears", "wears", "years", "sears"},
			round: qordle.Dictionary{
				"fears", "gears", "hears", "lears",
				"pears", "wears", "years", "sears"},
			strategy: new(qordle.Frequency),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := qordle.NewSpeculator(tt.words, tt.strategy)
			dictionary := s.Apply(tt.round)
			a.Equal(tt.result, dictionary)
			if tt.strategy == nil {
				a.Equal("speculate", s.String())
			} else {
				name := fmt.Sprintf("speculate{%s}", tt.strategy.String())
				a.Equal(name, s.String())
			}
		})
	}
}

func FuzzSpeculate(f *testing.F) {
	for _, x := range []string{"foo", "label", "start", "12345"} {
		f.Add(x)
	}
	f.Fuzz(func(t *testing.T, s string) {
		a := assert.New(t)
		st := qordle.NewSpeculator(
			qordle.Dictionary{
				"gyppy", "ghyll", "hyphy", "glyph", "layer", s},
			new(qordle.Frequency))
		dictionary := st.Apply(
			qordle.Dictionary{
				"fears", "gears", "hears", "lears", "pears", "wears", s,
				"years", "sears"})
		a.Greater(len(dictionary), 0)
	})
}

func BenchmarkSpeculate(b *testing.B) {
	wordlists := map[string]qordle.Dictionary{
		"none": {
			"fears", "gears", "hears", "lears",
			"pears", "wears", "years", "sears",
		},
		"head": {
			"glyph", "fears", "gears", "hears", "zears",
			"lears", "pears", "wears", "years", "sears",
		},
		"tail": {
			"fears", "gears", "hears", "lears", "zears",
			"pears", "wears", "years", "sears", "glyph",
		},
		"middle": {
			"fears", "gears", "hears", "lears", "zears",
			"pears", "glyph", "years", "sears", "wears",
		},
	}
	a := assert.New(b)
	solutions, err := qordle.Read("solutions")
	a.NoError(err)
	frequency, position := new(qordle.Frequency), new(qordle.Position)
	chain := qordle.NewChain(frequency, position)
	for _, strategy := range []qordle.Strategy{frequency, position, chain} {
		strategy = qordle.NewSpeculator(solutions, strategy)
		for key, val := range wordlists {
			b.Run("wordlist::"+strategy.String()+"::"+key, func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					a.Greater(len(strategy.Apply(val)), 0)
				}
			})
		}
	}
}

func BenchmarkChain(b *testing.B) {
	a := assert.New(b)
	solutions, err := qordle.Read("solutions")
	a.NoError(err)
	frequency, position := new(qordle.Frequency), new(qordle.Position)
	strategy := qordle.NewChain(frequency, position)
	b.Run(strategy.String(), func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			a.Equal(len(solutions), len(strategy.Apply(solutions)))
		}
	})
}

func TestStrategies(t *testing.T) {
	tests := []harness{
		{
			name: "strategies",
			args: []string{"strategies"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandStrategies)
		})
	}
}
