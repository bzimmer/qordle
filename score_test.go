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
