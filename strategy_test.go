package qordle_test

import (
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/stretchr/testify/assert"
)

func TestAlpha(t *testing.T) {
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
			a := assert.New(t)
			dictionary := qordle.Alpha(tt.words)
			a.Equal(tt.result, dictionary)
		})
	}
}

func TestFrequency(t *testing.T) {
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
			a := assert.New(t)
			dictionary := qordle.Frequency(tt.words)
			a.Equal(tt.result, dictionary)
		})
	}
}

func TestPosition(t *testing.T) {
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
			a := assert.New(t)
			dictionary := qordle.Position(tt.words)
			a.Equal(tt.result, dictionary)
		})
	}
}
