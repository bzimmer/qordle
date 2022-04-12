package qordle_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestScore(t *testing.T) {
	for _, tt := range []struct {
		name, secret    string
		guesses, scores []string
		err             bool
	}{
		{
			name:    "simple",
			secret:  "buyer",
			guesses: []string{"brain", "beret"},
			scores:  []string{"B~rain", "B~e~rEt"},
		},
		{
			name:    "different lengths",
			secret:  "humph",
			guesses: []string{"table", "tables"},
			err:     true,
		},
		{
			name:    "empty",
			secret:  "empty",
			guesses: []string{},
			scores:  []string{},
		},
	} {
		tt := tt
		t.Run(tt.secret, func(t *testing.T) {
			a := assert.New(t)
			scores, err := qordle.Score(tt.secret, tt.guesses...)
			if tt.err {
				a.Error(err)
				a.Nil(scores)
				return
			}
			a.NoError(err)
			a.Equal(tt.scores, scores)
		})
	}
}

func TestScoreCommand(t *testing.T) {
	for _, tt := range []struct {
		name         string
		words, score []string
		err          bool
	}{
		{
			name:  "score",
			words: []string{"table", "cable"},
			score: []string{"cABLE"},
		},
		{
			name:  "score with error",
			words: []string{"tableau", "cable"},
			err:   true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			var builder strings.Builder
			app := &cli.App{
				Name:     tt.name,
				Writer:   &builder,
				Commands: []*cli.Command{qordle.CommandScore()},
			}
			err := app.Run(append([]string{"qordle", "score"}, tt.words...))
			if tt.err {
				a.Error(err)
				return
			}
			a.NoError(err)
			var res []string
			err = json.Unmarshal([]byte(builder.String()), &res)
			a.NoError(err)
			a.Equal(tt.score, res)
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
