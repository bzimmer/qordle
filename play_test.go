package qordle_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/bzimmer/qordle"
)

func TestGame(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, start, secret       string
		strategy, errStrategy     string
		dictionary, errDictionary string
	}{
		{
			name:       "no startng word",
			secret:     "shine",
			strategy:   "frequency",
			dictionary: "solutions",
		},
		{
			name:       "starting word",
			start:      "soare",
			secret:     "shine",
			strategy:   "frequency",
			dictionary: "solutions",
		},
		{
			name:       "simple with qordle dictionary",
			start:      "soare",
			secret:     "shine",
			strategy:   "frequency",
			dictionary: "qordle",
		},
		{
			name:          "missing dictionary",
			start:         "soare",
			secret:        "train",
			strategy:      "frequency",
			errDictionary: "missing dictionary",
		},
		{
			name:        "missing strategy",
			start:       "soare",
			secret:      "train",
			dictionary:  "qordle",
			errStrategy: "missing strategy",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			a.NotEqual(tt.strategy, tt.errStrategy)
			a.NotEqual(tt.dictionary, tt.errDictionary)
			var err error
			var st qordle.Strategy
			if tt.strategy != "" {
				st, err = qordle.NewStrategy(tt.strategy)
				a.NoError(err)
				a.NotNil(st)
			}
			var dt qordle.Dictionary
			if tt.dictionary != "" {
				dt, err = qordle.Read(tt.dictionary)
				a.NoError(err)
				a.NotNil(dt)
			}
			game := qordle.NewGame(
				qordle.WithStrategy(st),
				qordle.WithDictionary(dt),
				qordle.WithStart(tt.start))
			scoreboard, err := game.Play(tt.secret)
			if tt.errStrategy != "" {
				a.Error(err)
				a.Equal(tt.errStrategy, err.Error())
				return
			}
			if tt.errDictionary != "" {
				a.Error(err)
				a.Equal(tt.errDictionary, err.Error())
				return
			}
			a.NoError(err)
			a.NotNil(scoreboard)
			a.GreaterOrEqual(len(scoreboard.Rounds), 1)
			upper := cases.Upper(language.English)
			winner := scoreboard.Rounds[len(scoreboard.Rounds)-1]
			a.Equal(upper.String(tt.secret), winner.Scores[len(winner.Scores)-1])
		})
	}
}

func TestPlayCommand(t *testing.T) {
	for _, tt := range []struct {
		name, err               string
		args, guesses, wordlist []string
		success                 bool
	}{
		{
			name:    "table",
			args:    []string{"-s", "position", "table"},
			guesses: []string{"so~arE", "mAiLE", "cABLE", "fABLE", "gABLE", "hABLE", "TABLE"},
			success: true,
		},
		{
			name:    "one word solution",
			args:    []string{"soare"},
			guesses: []string{"SOARE"},
			success: true,
		},
		{
			name:    "two word solution",
			args:    []string{"-s", "frequency", "--start", "brain", "raise"},
			guesses: []string{"b~r~a~in", "RAISE"},
			success: true,
		},
		{
			name: "failed to find secret",
			args: []string{"12345"},
		},
		{
			name: "secret and guess lengths do not match",
			args: []string{"123456"},
			err:  "secret and guess lengths do not match",
		},
		{
			name: "invalid strategy",
			args: []string{"-s", "foobar", "table"},
			err:  "unknown strategy `foobar`",
		},
		{
			name: "invalid wordlist",
			args: []string{"-w", "foobar", "table"},
			err:  "invalid wordlist `foobar`",
		},
		{
			name:    "six letter word with explicit strategy",
			args:    []string{"-s", "position", "-w", "qordle", "--start", "shadow", "treaty"},
			guesses: []string{"sh~adow", "canAan", "a~e~rATe", "TREATY"},
			success: true,
		},
		{
			name:    "auto play",
			args:    []string{"-A", "-s", "position", "-w", "qordle", "--start", "shadow", "treaty"},
			guesses: []string{"sh~adow", "canAan", "a~e~rATe", "TREATY"},
			success: true,
		},
		{
			name:    "six letter word with no implicit strategy",
			args:    []string{"-w", "qordle", "--start", "shadow", "treaty"},
			guesses: []string{"sh~adow", "~alin~e~r", "p~e~rAc~t", "~rugAT~e", "TREATY"},
			success: true,
		},
		{
			name:    "six letter word with no implicit strategy and speculation",
			args:    []string{"-w", "qordle", "--start", "shadow", "-S", "treaty"},
			guesses: []string{"sh~adow", "~alin~e~r", "p~e~rAc~t", "~rugAT~e", "TREATY"},
			success: true,
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
				a.Error(err)
				a.Equal(tt.err, err.Error())
				return
			}
			var res qordle.Scoreboard
			err = json.Unmarshal([]byte(builder.String()), &res)
			a.NoError(err)
			if tt.success {
				winner := res.Rounds[len(res.Rounds)-1]
				a.Equal(tt.guesses, winner.Scores)
			} else {
				a.False(tt.success)
			}
		})
	}
}
