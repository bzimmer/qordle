package qordle_test

import (
	"encoding/json"
	"io"
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
	a := assert.New(t)

	decode := func(c *cli.Context) *qordle.Round {
		var res qordle.Scoreboard
		dec := json.NewDecoder(c.App.Writer.(io.Reader))
		err := dec.Decode(&res)
		if err != nil {
			a.FailNow(err.Error())
		}
		return res.Rounds[len(res.Rounds)-1]
	}

	for _, tt := range []harness{
		{
			name: "table",
			args: []string{"play", "-s", "position", "table"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal(
					[]string{"so.arE", "mAiLE", "cABLE", "fABLE", "gABLE", "hABLE", "TABLE"},
					round.Scores)
				return nil
			},
		},
		{
			name: "one word solution",
			args: []string{"play", "soare"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"SOARE"}, round.Scores)
				return nil
			},
		},
		{
			name: "two word solution",
			args: []string{"play", "-s", "frequency", "--start", "brain", "raise"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"b.r.a.in", "RAISE"}, round.Scores)
				return nil
			},
		},
		{
			name: "six letter word with explicit strategy",
			args: []string{"play", "-s", "position", "-w", "qordle", "--start", "shadow", "treaty"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"sh.adow", "canAan", "a.e.rATe", "TREATY"}, round.Scores)
				return nil
			},
		},
		{
			name: "failed to find secret",
			args: []string{"play", "12345"},
		},
		{
			name: "secret and guess lengths do not match",
			args: []string{"play", "123456"},
			err:  "secret and guess lengths do not match",
		},
		{
			name: "invalid strategy",
			args: []string{"play", "-s", "foobar", "table"},
			err:  "unknown strategy `foobar`",
		},
		{
			name: "invalid wordlist",
			args: []string{"play", "-w", "foobar", "table"},
			err:  "invalid wordlist `foobar`",
		},
		{
			name: "display a progress bar",
			args: []string{"play", "-B", "-s", "position", "-w", "qordle", "--start", "shadow", "treaty"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"sh.adow", "canAan", "a.e.rATe", "TREATY"}, round.Scores)
				return nil
			},
		},
		{
			name: "six letter word with no implicit strategy",
			args: []string{"play", "-w", "qordle", "--start", "shadow", "treaty"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"sh.adow", ".alin.e.r", "p.e.rAc.t", ".rugAT.e", "TREATY"}, round.Scores)
				return nil
			},
		},
		{
			name: "six letter word with no implicit strategy and speculation",
			args: []string{"play", "-w", "qordle", "--start", "shadow", "-S", "treaty"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"sh.adow", ".alin.e.r", "p.e.rAc.t", ".rugAT.e", "TREATY"}, round.Scores)
				return nil
			},
		},
		{
			name: "read from stdin and display a progress bar",
			args: []string{"play", "-w", "qordle", "-S"},
			before: func(c *cli.Context) error {
				c.App.Reader = strings.NewReader("train")
				return nil
			},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{"soA.re", ".r.iA.n.t", "TRAIN"}, round.Scores)
				return nil
			},
		},
		{
			name: "exceed the number of rounds",
			args: []string{"play", "--start", "brain", "-S", "-r", "2", "sills"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.False(round.Success)
				a.Equal([]string{
					"bra.in", ".i.sLet", "kILoS", "mILdS", "hILuS",
					"gyp.sy", "f.locS", "jowLS", "vILLS", "zILLS"},
					round.Scores)
				return nil
			},
		},
		{
			name: "fail to find the solution",
			args: []string{"play", "qwert"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.False(round.Success)
				a.Equal([]string{"soaR.e", ".en.tRy", "piERT", "blERT", "chERT"}, round.Scores)
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandPlay)
		})
	}
}
