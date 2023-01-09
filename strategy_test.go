package qordle_test

import (
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

func TestSpeculate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                    string
		strategy                qordle.Strategy
		words, result, guessing qordle.Dictionary
	}{
		{
			name:     "no guessing game",
			words:    qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			guessing: qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result:   qordle.Dictionary{"false", "halse", "easle", "fause", "haste"},
			strategy: new(qordle.Frequency),
		},
		{
			name:  "guessing",
			words: qordle.Dictionary{"gyppy", "ghyll", "hyphy", "glyph", "layer"},
			result: qordle.Dictionary{"glyph", "ghyll", "gyppy", "hyphy", "fears",
				"gears", "hears", "lears", "pears", "wears", "years", "sears"},
			guessing: qordle.Dictionary{"fears", "gears", "hears", "lears", "pears",
				"wears", "years", "sears"},
			strategy: new(qordle.Frequency),
		},
		{
			name:     "one word",
			words:    qordle.Dictionary{"layer"},
			result:   qordle.Dictionary{"layer"},
			guessing: qordle.Dictionary{"layer"},
			strategy: new(qordle.Frequency),
		},
		{
			name:     "empty",
			words:    qordle.Dictionary{},
			result:   qordle.Dictionary{},
			guessing: qordle.Dictionary{},
			strategy: new(qordle.Frequency),
		},
		{
			name:     "no strategy",
			words:    qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			guessing: qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result:   qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			strategy: nil,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			s := qordle.NewSpeculator(tt.words, tt.strategy)
			dictionary := s.Apply(tt.guessing)
			a.Equal(tt.result, dictionary)
			a.Equal("speculate", s.String())
		})
	}
}

func FuzzSpeculate(f *testing.F) {
	for _, x := range []string{"foo", "label", "start"} {
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
