package qordle_test

import (
	"fmt"
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type dictionaryTest struct {
	words []string
}

func (d *dictionaryTest) Words() []string {
	return d.words
}

func TestDictionaryFs(t *testing.T) {
	for _, tt := range []struct {
		name          string
		words, result []string
	}{
		{
			name:   "readfs",
			words:  []string{"hoody", "foobar"},
			result: []string{"hoody", "foobar"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			afs := afero.NewMemMapFs()
			path := "/tmp/share/dict/words"
			a.NoError(afs.MkdirAll("/tmp/share/dict", 0755))
			fp, err := afs.Create(path)
			a.NoError(err)
			for _, word := range tt.words {
				fmt.Fprintln(fp, word)
			}
			a.NoError(fp.Close())

			dictionary, err := qordle.DictionaryFs(afs, path)
			a.NoError(err)
			a.Equal(tt.result, dictionary.Words())
		})
	}
}
