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

const (
	dottxt    = ".txt"
	dataFsDir = "data"
)

//go:embed data
var dataFs embed.FS

func dict(r io.Reader) (Dictionary, error) {
	var res []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
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
	fp, err := dataFs.Open(filepath.Join(dataFsDir, path+dottxt))
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return dict(fp)
}

func ListEm() ([]string, error) {
	dicts := make([]string, 0)
	if err := fs.WalkDir(dataFs, dataFsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dicts = append(dicts, strings.Replace(d.Name(), dottxt, "", 1))
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(dicts)
	return dicts, nil
}

func CommandWordlists() *cli.Command {
	return &cli.Command{
		Name:  "wordlists",
		Usage: "list all available wordlists",
		Action: func(c *cli.Context) error {
			list, err := ListEm()
			if err != nil {
				return err
			}
			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(list)
		},
	}
}
