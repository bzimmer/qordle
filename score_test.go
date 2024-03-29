package qordle_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestScore(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, secret    string
		guesses, scores []string
		err             error
	}{
		{
			name:    "simple",
			secret:  "buyer",
			guesses: []string{"brain", "beret"},
			scores:  []string{"B.rain", "Be.rEt"},
		},
		{
			name:    "different lengths",
			secret:  "humph",
			guesses: []string{"table", "tables"},
			err:     qordle.ErrInvalidLength,
		},
		{
			name:    "empty",
			secret:  "empty",
			guesses: []string{},
			scores:  []string{},
		},
		{
			name:    "alphanumeric",
			secret:  "humph",
			guesses: []string{"12345"},
			scores:  []string{"12345"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			scores, err := qordle.Score(tt.secret, tt.guesses...)
			if tt.err != nil {
				a.ErrorIs(err, tt.err)
				return
			}
			a.NoError(err)
			a.Equal(tt.scores, scores)
		})
	}
}

func TestScoreCommand(t *testing.T) {
	a := assert.New(t)
	for _, tt := range []harness{
		{
			name: "score",
			args: []string{"score", "table", "cable"},
			after: func(c *cli.Context) error {
				var res []string
				dec := json.NewDecoder(c.App.Writer.(io.Reader))
				a.NoError(dec.Decode(&res))
				a.Equal([]string{"cABLE"}, res)
				return nil
			},
		},
		{
			name: "score with error",
			err:  "secret and guess lengths do not match",
			args: []string{"score", "tableau", "cable"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandScore)
		})
	}
}
