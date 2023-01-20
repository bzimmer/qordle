package qordle_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestSuggestCommand(t *testing.T) {
	for _, tt := range []struct {
		name, strategy, dictionary, err string
		speculate                       bool
		args                            []string
	}{
		{
			name:     "alpha",
			strategy: "a",
		},
		{
			name:     "frequency",
			strategy: "f",
		},
		{
			name:     "position",
			strategy: "p",
		},
		{
			name:     "unknown",
			strategy: "u",
			err:      "unknown strategy `u`",
		},
		{
			name:      "speculate",
			speculate: true,
		},
		{
			name: "bad pattern",
			args: []string{"--pattern", "[A-Z"},
			err:  "error parsing regexp: missing closing ]: `[A-Z`",
		},
		{
			name: "bad wordlist",
			args: []string{"-w", "foobar"},
			err:  "invalid wordlist `foobar`",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			app := &cli.App{
				Name:     tt.name,
				Writer:   &strings.Builder{},
				Commands: []*cli.Command{qordle.CommandSuggest()},
			}
			args := []string{"qordle", "suggest"}
			if tt.strategy != "" {
				args = append(args, "-s", tt.strategy)
			}
			if tt.speculate {
				args = append(args, "-S")
			}
			if len(tt.args) > 0 {
				args = append(args, tt.args...)
			}
			args = append(args, "raise", "fol~l~y")
			err := app.Run(args)
			if tt.err != "" {
				a.Equal(tt.err, err.Error())
				return
			}
			a.NoError(err)
		})
	}
}
