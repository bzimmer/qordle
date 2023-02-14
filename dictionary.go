package qordle

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

type Dictionary []string

const data = "data"

//go:embed data
var dataFs embed.FS

func read(r io.Reader) (Dictionary, error) {
	if r == nil {
		return nil, errors.New("invalid reader")
	}
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

func Read(name string) (Dictionary, error) {
	fp, err := dataFs.Open(fmt.Sprintf("%s/%s.txt", data, name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("invalid wordlist `%s`", name)
		}
		return nil, err
	}
	defer fp.Close()
	return read(fp)
}

func wordlistFlag() cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "wordlist",
		Aliases: []string{"w"},
		Usage:   "use the specified embedded word list",
	}
}

func wordlists(c *cli.Context, wordlists ...string) (Dictionary, error) {
	if c.IsSet("wordlist") {
		wordlists = c.StringSlice("wordlist")
	}
	w := map[string]struct{}{}
	for _, wordlist := range wordlists {
		t, err := Read(wordlist)
		if err != nil {
			return nil, err
		}
		for i := range t {
			w[t[i]] = struct{}{}
		}
	}
	i, dictionary := 0, make(Dictionary, len(w))
	for k := range w {
		dictionary[i] = k
		i++
	}
	return dictionary, nil
}

func CommandWordlists() *cli.Command {
	return &cli.Command{
		Name:  "wordlists",
		Usage: "list all available wordlists",
		Action: func(c *cli.Context) error {
			var lists []string
			if err := fs.WalkDir(dataFs, data, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				lists = append(lists, strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())))
				return nil
			}); err != nil {
				return err
			}
			sort.Strings(lists)
			return Runtime(c).Encoder.Encode(lists)
		},
	}
}
