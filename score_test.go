package qordle_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestMarks(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	for _, tt := range []struct {
		marks qordle.Marks
		key   int
	}{
		{
			marks: qordle.Marks{qordle.MarkExact, qordle.MarkMiss, qordle.MarkExact, qordle.MarkMisplaced, qordle.MarkMiss},
			key:   20210,
		},
		{
			marks: qordle.Marks{qordle.MarkMiss, qordle.MarkMiss, qordle.MarkExact, qordle.MarkMisplaced, qordle.MarkMiss},
			key:   210,
		},
	} {
		tt := tt
		name := fmt.Sprintf("%v", tt.marks)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tt.key, tt.marks.Key())
		})
	}
}

func TestScore(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, secret    string
		guesses, scores []string
		panic           bool
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
			panic:   true,
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
			if tt.panic {
				a.Panics(func() {
					qordle.Score(tt.secret, tt.guesses...)
				})
				return
			}
			a.Equal(tt.scores, qordle.Score(tt.secret, tt.guesses...))
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
