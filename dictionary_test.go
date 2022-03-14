package qordle_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestDictionaryFs(t *testing.T) {
	path := "data/solutions.txt"
	for _, tt := range []struct {
		name, path    string
		words, result qordle.Dictionary
		err           bool
	}{
		{
			name:   "valid file",
			path:   path,
			words:  []string{"hoody", "foobar"},
			result: []string{"hoody", "foobar"},
			err:    false,
		},
		{
			name:   "invalid file",
			path:   "data/missing-file.txt",
			words:  []string{"hoody", "foobar"},
			result: []string{},
			err:    true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			afs := afero.NewMemMapFs()
			parent, _ := filepath.Split(path)
			a.NoError(afs.MkdirAll(parent, 0755))
			fp, err := afs.Create(path)
			a.NoError(err)
			for _, word := range tt.words {
				fmt.Fprintln(fp, word)
			}
			a.NoError(fp.Close())

			dictionary, err := qordle.DictionaryFs(afs, tt.path)
			switch tt.err {
			case true:
				a.Error(err)
				a.Empty(dictionary)
			case false:
				a.NoError(err)
				a.Equal(tt.result, dictionary)
			}
		})
	}
}

func TestDictionaryEm(t *testing.T) {
	for _, tt := range []struct {
		name, path string
		err        bool
	}{
		{
			name: "valid file",
			path: "solutions.txt",
			err:  false,
		},
		{
			name: "invalid file",
			path: "missing-file.txt",
			err:  true,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			dictionary, err := qordle.DictionaryEm(tt.path)
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

func TestListEm(t *testing.T) {
	a := assert.New(t)
	list, err := qordle.ListEm()
	a.NoError(err)
	a.NotEmpty(list)
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
