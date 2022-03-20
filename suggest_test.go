package qordle_test

import (
	"strings"
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestSuggestCommand(t *testing.T) {
	for _, tt := range []struct {
		name, strategy, dictionary string
		err                        bool
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
			err:      true,
		},
		{
			name:       "invalid dictionary",
			dictionary: "blahblahblah",
			err:        true,
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
			if tt.dictionary != "" {
				args = append(args, "--dictionary", tt.dictionary)
			}
			if tt.strategy != "" {
				args = append(args, "-s", tt.strategy)
			}
			args = append(args, "raise", "fol~l~y")
			err := app.Run(args)
			if tt.err {
				a.Error(err)
				return
			}
			a.NoError(err)
		})
	}
}
