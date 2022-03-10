package qordle

import (
	"bufio"

	"github.com/spf13/afero"
)

type Dictionary interface {
	Words() []string
}

type dictionary struct {
	words []string
}

func (d *dictionary) Words() []string {
	return d.words
}

func DictionaryFs(fs afero.Fs, path string) (Dictionary, error) {
	fp, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	d := new(dictionary)
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		d.words = append(d.words, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func DictionarySlice(words []string) (Dictionary, error) {
	return &dictionary{words: words}, nil
}
