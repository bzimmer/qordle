package qordle_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestSuggestCommand(t *testing.T) {
	a := assert.New(t)
	for _, tt := range []harness{
		{
			name: "alpha",
			args: []string{"suggest", "--strategy", "a", "raise", "fol.l.y"},
		},
		{
			name: "frequency",
			args: []string{"suggest", "--strategy", "f", "raise", "fol.l.y"},
		},
		{
			name: "position",
			args: []string{"suggest", "--strategy", "p", "raise", "fol.l.y"},
		},
		{
			name: "combination",
			args: []string{"suggest", "-s", "b", "-s", "f", "raise", "fol.l.y"},
		},
		{
			name: "unknown",
			args: []string{"suggest", "--strategy", "u", "raise", "fol.l.y"},
			err:  "unknown strategy `u`",
		},
		{
			name: "unknown strategy chain",
			args: []string{"suggest", "-s", "f", "-s", "q", "raise", "fol.l.y"},
			err:  "unknown strategy `q`",
		},
		{
			name: "speculate",
			args: []string{"suggest", "-S", "raise", "fol.l.y"},
		},
		{
			name: "bad guesses",
			args: []string{"suggest", "fol.l."},
			err:  "too few characters",
		},
		{
			name: "bad wordlist",
			args: []string{"suggest", "-w", "foobar", "raise", "fol.l.y"},
			err:  "invalid wordlist `foobar`",
		},
		{
			name: "speculate for ?ound",
			args: []string{"suggest", "-w", "solutions", "-S", "--strategy", "frequency", "trai.n", ".o.u.nce", "bOUND"},
			after: func(c *cli.Context) error {
				var res []string
				dec := json.NewDecoder(c.App.Writer.(io.Reader))
				err := dec.Decode(&res)
				a.NoError(err)
				a.Equal([]string{"smash", "found", "hound", "mound", "pound", "sound", "wound"}, res)
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandSuggest)
		})
	}
}

func TestValidateCommand(t *testing.T) {
	for _, tt := range []harness{
		{
			name: "invalid",
			args: []string{"validate", "raise", "fol.l.y"},
		},
		{
			name: "invalid",
			args: []string{"validate", "yleaz", "fol.l.y"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandValidate)
		})
	}
}
