package qordle_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestRead(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, path string
		err        bool
	}{
		{
			name: "valid file",
			path: "solutions",
			err:  false,
		},
		{
			name: "invalid file",
			path: "missing-file",
			err:  true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			dictionary, err := qordle.Read(tt.path)
			switch tt.err {
			case true:
				a.Error(err)
				a.Empty(dictionary)
			case false:
				a.NoError(err)
				a.NotEmpty(dictionary)
			}
		})
	}
}

func TestWordlistsCommand(t *testing.T) {
	a := assert.New(t)
	for _, tt := range []harness{
		{
			name: "wordlists",
			args: []string{"wordlists"},
			after: func(c *cli.Context) error {
				var res []string
				dec := json.NewDecoder(c.App.Writer.(io.Reader))
				a.NoError(dec.Decode(&res))
				a.Equal([]string{"possible", "qordle", "solutions"}, res)
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandWordlists)
		})
	}
}
