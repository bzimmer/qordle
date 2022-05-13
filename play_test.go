package qordle_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"

	"github.com/bzimmer/qordle"
)

func TestGame(t *testing.T) {
	for _, tt := range []struct {
		name, start, secret        string
		strategy, err_strategy     string
		dictionary, err_dictionary string
	}{
		{
			name:       "simple",
			start:      "soare",
			secret:     "shine",
			strategy:   "frequency",
			dictionary: "solutions",
		},
		{
			name:       "simple",
			start:      "soare",
			secret:     "shine",
			strategy:   "frequency",
			dictionary: "qordle",
		},
		{
			name:           "missing dictionary",
			start:          "soare",
			secret:         "train",
			strategy:       "frequency",
			err_dictionary: "missing dictionary",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			a.NotEqual(tt.strategy, tt.err_strategy)
			a.NotEqual(tt.dictionary, tt.err_dictionary)
			var err error
			var st qordle.Strategy
			if tt.strategy != "" {
				st, err = qordle.NewStrategy(tt.strategy)
				a.NoError(err)
				a.NotNil(st)
			}
			var dt qordle.Dictionary
			if tt.dictionary != "" {
				dt, err = qordle.DictionaryEm(tt.dictionary)
				a.NoError(err)
				a.NotNil(dt)
			}
			game := qordle.NewGame(
				qordle.WithStrategy(st),
				qordle.WithDictionary(dt),
				qordle.WithStart(tt.start))
			words, err := game.Play(tt.secret)
			if tt.err_strategy != "" {
				a.Error(err)
				a.Equal(tt.err_strategy, err.Error())
				return
			}
			if tt.err_dictionary != "" {
				a.Error(err)
				a.Equal(tt.err_dictionary, err.Error())
				return
			}
			a.NoError(err)
			a.NotNil(words)
			a.GreaterOrEqual(len(words), 1)
			upper := cases.Upper(qordle.Language)
			a.Equal(upper.String(tt.secret), words[len(words)-1])
		})
	}
}

func TestPlayCommand(t *testing.T) {
	for _, tt := range []struct {
		name, err               string
		args, guesses, wordlist []string
	}{
		{
			name:    "table",
			args:    []string{"-s", "position", "table"},
			guesses: []string{"so~arE", "mAiLE", "cABLE", "fABLE", "gABLE", "hABLE", "TABLE"},
		},
		{
			name:    "first guess is the secret",
			args:    []string{"soare"},
			guesses: []string{"SOARE"},
		},
		{
			name: "failed to find secret",
			args: []string{"12345"},
			err:  "failed to find secret",
		},
		{
			name: "secret and guess lengths do not match",
			args: []string{"123456"},
			err:  "secret and guess lengths do not match",
		},
		{
			name: "no word",
			err:  "expected at least one word to play",
		},
		{
			name:    "six letter word with explicit strategy",
			args:    []string{"-s", "position", "-w", "qordle", "--start", "shadow", "treaty"},
			guesses: []string{"sh~adow", "c~anA~an", "~a~e~rAT~e", "TREATY"},
		},
		{
			name:    "six letter word with no implicit strategy",
			args:    []string{"-w", "qordle", "--start", "shadow", "treaty"},
			guesses: []string{"sh~adow", "~alin~e~r", "p~e~rAc~t", "~rugAT~e", "TREATY"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			var builder strings.Builder
			app := &cli.App{
				Name:     tt.name,
				Writer:   &builder,
				Commands: []*cli.Command{qordle.CommandPlay()},
			}
			err := app.Run(append([]string{"qordle", "play"}, tt.args...))
			if tt.err != "" {
				a.Equal(tt.err, err.Error())
				return
			}
			var res []string
			err = json.Unmarshal([]byte(builder.String()), &res)
			a.NoError(err)
			a.Equal(tt.guesses, res)
		})
	}
}
