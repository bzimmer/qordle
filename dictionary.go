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

func (dict Dictionary) union(other Dictionary) Dictionary {
	if dict == nil {
		res := make(Dictionary, len(other))
		copy(res, other)
		return res
	}
	words := map[string]struct{}{}
	for _, d := range []Dictionary{dict, other} {
		for i := range d {
			words[d[i]] = struct{}{}
		}
	}
	i, dictionary := 0, make(Dictionary, len(words))
	for k := range words {
		dictionary[i] = k
		i++
	}
	return dictionary
}

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

func wordlistFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "wordlist",
			Aliases: []string{"w"},
			Usage:   "use the specified embedded word list",
		},
		// &cli.StringSliceFlag{
		// 	Name:    "Wordlist",
		// 	Aliases: []string{"W"},
		// 	Usage:   "use the specified external word list",
		// },
	}
}

func wordlists(c *cli.Context, wordlists ...string) (Dictionary, error) {
	var readers []func() (Dictionary, error)
	switch {
	// case c.IsSet("Wordlist"):
	// 	for _, wordlist := range c.StringSlice("Wordlist") {
	// 		w := wordlist
	// 		readers = append(readers, func() (Dictionary, error) {
	// 			fp, err := os.Open(w)
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			defer fp.Close()
	// 			return read(fp)
	// 		})
	// 	}
	case c.IsSet("wordlist"):
		wordlists = c.StringSlice("wordlist")
		fallthrough
	default:
		for _, wordlist := range wordlists {
			w := wordlist
			readers = append(readers, func() (Dictionary, error) {
				return Read(w)
			})
		}
	}
	var words Dictionary
	for _, reader := range readers {
		res, err := reader()
		if err != nil {
			return nil, err
		}
		words = words.union(res)
	}
	return words, nil
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
