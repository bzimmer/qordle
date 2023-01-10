package qordle_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestDictionaryEm(t *testing.T) {
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

func TestCommandWordlists(t *testing.T) {
	for _, tt := range []struct {
		name         string
		dictionaries []string
	}{
		{
			name:         "wordlists",
			dictionaries: []string{"possible", "qordle", "solutions"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			builder := &strings.Builder{}
			app := &cli.App{
				Name:     tt.name,
				Writer:   builder,
				Commands: []*cli.Command{qordle.CommandWordlists()},
			}
			err := app.Run([]string{"qordle", "wordlists"})
			a.NoError(err)
			res := []string{}
			err = json.Unmarshal([]byte(builder.String()), &res)
			a.NoError(err)
			a.Equal(tt.dictionaries, res)
		})
	}
}
