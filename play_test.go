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
		name, start, secret, errStrategy string
		dictionary, errDictionary        string
		strategy                         qordle.Strategy
	}{
		{
			name:       "no startng word",
			secret:     "shine",
			strategy:   new(qordle.Frequency),
			dictionary: "solutions",
		},
		{
			name:       "starting word",
			start:      "soare",
			secret:     "shine",
			strategy:   new(qordle.Frequency),
			dictionary: "solutions",
		},
		{
			name:       "simple with qordle dictionary",
			start:      "soare",
			secret:     "shine",
			strategy:   new(qordle.Frequency),
			dictionary: "qordle",
		},
		{
			name:          "empty dictionary",
			start:         "soare",
			secret:        "train",
			strategy:      new(qordle.Frequency),
			errDictionary: "empty dictionary",
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
			var err error
			var dt qordle.Dictionary
			if tt.dictionary != "" {
				dt, err = qordle.Read(tt.dictionary)
				a.NoError(err)
				a.NotNil(dt)
			}
			game := qordle.NewGame(
				qordle.WithStrategy(tt.strategy),
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
		a.NoError(err)
		return res.Rounds[len(res.Rounds)-1]
	}

	for _, tt := range []harness{
		{
			name: "table",
			args: []string{"play", "-s", "position", "--start", "soare", "table"},
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
			args: []string{"play", "--start", "soare", "soare"},
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
			name: "unable to acquire starting word",
			args: []string{"play", "123456"},
			err:  "empty dictionary",
		},
		{
			name: "secret and guess lengths do not match",
			args: []string{"play", "--start", "soare", "123456"},
			err:  "empty dictionary",
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
			name: "read from nil stdin",
			args: []string{"play", "-w", "qordle"},
			err:  "invalid reader",
			before: func(c *cli.Context) error {
				c.App.Reader = nil
				return nil
			},
		},
		{
			name: "read from stdin and display a progress bar",
			args: []string{"play", "-w", "qordle", "-S", "--start", "soare"},
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
			args: []string{"play", "--start", "soare", "qwert"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.False(round.Success)
				a.Equal([]string{"soaR.e", ".en.tRy", "piERT", "blERT", "chERT"}, round.Scores)
				return nil
			},
		},
		{
			name: "bigram strategy",
			args: []string{"play", "-s", "bigram", "-S", "aahed"},
			after: func(c *cli.Context) error {
				round := decode(c)
				a.True(round.Success)
				a.Equal([]string{".ering", "lAtED", "cAsED", ".hAdED", "AAHED"}, round.Scores)
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

func BenchmarkPlay(b *testing.B) {
	a := assert.New(b)
	solutions, err := qordle.Read("solutions")
	a.NoError(err)
	dictionary, err := qordle.Read("qordle")
	a.NoError(err)
	strategy := qordle.NewSpeculator(dictionary, new(qordle.Frequency))

	game := qordle.NewGame(
		qordle.WithDictionary(solutions),
		qordle.WithStrategy(strategy),
	)
	for _, secret := range []string{"board", "brain", "mound", "lills", "qwert"} {
		b.Run("secret::"+secret, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var board *qordle.Scoreboard
				board, err = game.Play(secret)
				a.NoError(err)
				a.Greater(len(board.Rounds), 0)
			}
		})
	}
}
