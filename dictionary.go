package qordle

import (
	"bufio"
	"embed"
	"io"

	"github.com/spf13/afero"
)

type Dictionary []string

//go:embed data
var data embed.FS

func dict(r io.Reader) (Dictionary, error) {
	var res []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}
	err := scanner.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DictionaryFs(fs afero.Fs, path string) (Dictionary, error) {
	fp, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return dict(fp)
}

func DictionaryEmbed() (Dictionary, error) {
	fp, err := data.Open("data/qordle.txt")
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return dict(fp)
}
