package qordle_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDictionaryFs(t *testing.T) {
	path := "/tmp/share/dict/qordle.txt"
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
				a.Empty(tt.result)
			case false:
				a.NoError(err)
				a.Equal(tt.result, dictionary)
			}
		})
	}
}

func TestDictionaryEmbedded(t *testing.T) {
	a := assert.New(t)
	dictionary, err := qordle.DictionaryEmbed()
	a.NoError(err)
	a.GreaterOrEqual(len(dictionary), 10)
}
