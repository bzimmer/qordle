package qordle_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestSuggestCommand(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, strategy, dictionary, err string
		speculate                       bool
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
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
