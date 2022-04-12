package qordle

import (
	"bufio"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

type Dictionary []string

const Data = "data"

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

func DictionaryFs(afs afero.Fs, path string) (Dictionary, error) {
	fp, err := afs.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	log.Info().Str("path", path).Msg("fs")
	return dict(fp)
}

func DictionaryEm(path string) (Dictionary, error) {
	fp, err := data.Open(filepath.Join(Data, path))
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return dict(fp)
}

func ListEm() ([]string, error) {
	dicts := make([]string, 0)
	werr := fs.WalkDir(data, Data, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dicts = append(dicts, strings.Replace(d.Name(), ".txt", "", 1))
		return nil
	})
	if werr != nil {
		return nil, werr
	}
	sort.Strings(dicts)
	return dicts, nil
}

func CommandWordlists() *cli.Command {
	return &cli.Command{
		Name:  "wordlists",
		Usage: "list all available wordlists",
		Action: func(c *cli.Context) error {
			enc := json.NewEncoder(c.App.Writer)
			list, err := ListEm()
			if err != nil {
				return err
			}
			return enc.Encode(list)
		},
	}
}
