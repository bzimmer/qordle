package qordle_test

import (
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/stretchr/testify/assert"
)

func TestSuggest(t *testing.T) {
	for _, tt := range []struct {
		name          string
		words, result qordle.Dictionary
	}{
		{
			name:   "suggest one",
			words:  qordle.Dictionary{"easle", "false", "fause", "halse", "haste"},
			result: qordle.Dictionary{"false", "halse", "easle", "fause", "haste"},
		},
		{
			name:   "suggest two",
			words:  qordle.Dictionary{"maths", "sport", "brain", "raise"},
			result: qordle.Dictionary{"raise", "maths", "sport", "brain"},
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
			dictionary := qordle.Suggest(tt.words)
			a.Equal(tt.result, dictionary)
		})
	}
}
