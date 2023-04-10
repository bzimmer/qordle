package qordle_test

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestLetterboxedCommand(t *testing.T) {
	a := assert.New(t)
	decode := func(c *cli.Context) [][]string {
		var results [][]string
		dec := json.NewDecoder(c.App.Writer.(io.Reader))
		for {
			var res []string
			err := dec.Decode(&res)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return results
				}
				a.FailNow(err.Error())
			}
			results = append(results, res)
		}
	}
	for _, tt := range []harness{
		{
			name: "letterboxed with 0 arguments",
			args: []string{"letterboxed", "-w", "solutions", "--max", "4"},
			after: func(c *cli.Context) error {
				a.Equal(3, len(decode(c)))
				return nil
			},
		},
		{
			name: "letterboxed with 1 argument",
			args: []string{"letterboxed", "-w", "solutions", "--max", "4", "rul-eya-gdh-opb"},
			after: func(c *cli.Context) error {
				a.Equal(57, len(decode(c)))
				return nil
			},
		},
		{
			name: "letterboxed with 4 arguments",
			args: []string{"letterboxed", "-w", "solutions", "--max", "4", "rul", "eya", "gdh", "opb"},
			after: func(c *cli.Context) error {
				a.Equal(57, len(decode(c)))
				return nil
			},
		},
		{
			name: "letterboxed with invalid wordlist",
			args: []string{"letterboxed", "-w", "missing", "--max", "4"},
			err:  "invalid wordlist `missing`",
		},
		{
			name: "letterboxed with 5 arguments",
			args: []string{"letterboxed", "-w", "solutions", "--max", "4", "rul", "eya", "gdh", "opb", "qst"},
			err:  "found 5 sides",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandLetterBoxed)
		})
	}
}
