package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/bzimmer/qordle"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const length = 7
const cutoff = 0.0001

type TableFunc func(qordle.Dictionary) (string, string)

//nolint:gochecknoglobals
var source = `// Code generated by go generate; DO NOT EDIT.
package qordle

// letter frequencies
var frequencies = map[rune]float64{
	{{.frequencies}}
}

// letter frequencies by position for all words lte to {{.length}} letters
var positions = map[rune]map[int]float64{
	{{.positions}}
}

// bigram frequencies for all bigrams with frequencies gte to {{.cutoff}}
var bigrams = map[string]float64{
	{{.bigrams}}
}
`

func chunk[T any](slice []T, batch int) [][]T {
	batches := make([][]T, 0, (len(slice)+batch-1)/batch)
	for batch < len(slice) {
		slice, batches = slice[batch:], append(batches, slice[0:batch:batch])
	}
	return append(batches, slice)
}

func frequencies(words qordle.Dictionary) (string, string) {
	var total float64
	freqs := make(map[rune]float64)
	for i := range words {
		rs := []rune(words[i])
		for j := range rs {
			total++
			freqs[rs[j]]++
		}
	}
	var builder []string
	for key := range freqs {
		builder = append(builder, fmt.Sprintf("'%c': %0.4f,", key, freqs[key]/total))
	}
	sort.Slice(builder, func(i, j int) bool { return builder[i] < builder[j] })
	var output []string
	for _, val := range chunk(builder, 5) {
		output = append(output, strings.Join(val, " "))
	}
	return "frequencies", strings.Join(output, "\n")
}

func positions(words qordle.Dictionary) (string, string) {
	idx := make(map[int]float64)
	pos := make(map[rune]map[int]float64)

	for i := range words {
		rs := []rune(words[i])
		if len(rs) > length {
			continue
		}
		for j := range rs {
			idx[j]++
			val, ok := pos[rs[j]]
			if !ok {
				val = make(map[int]float64)
				pos[rs[j]] = val
			}
			val[j]++
		}
	}

	var output []string
	for letter, idxs := range pos {
		var row []string
		for index, total := range idxs {
			s := fmt.Sprintf("%d: %0.4f", index, total/idx[index])
			row = append(row, s)
		}
		sort.Slice(row, func(i, j int) bool {
			return row[i] < row[j]
		})
		s := fmt.Sprintf("'%c': { %s }", letter, strings.Join(row, ", "))
		output = append(output, s)
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i] < output[j]
	})
	output[len(output)-1] += ","
	return "positions", strings.Join(output, ",\n")
}

func bigrams(words qordle.Dictionary) (string, string) {
	var i int
	var total float64
	grams := make(map[string]float64)
	for _, word := range words {
		switch n := len(word); n {
		case 0, 1:
		default:
			i = 0
			for i+2 < n {
				total++
				grams[word[i:i+2]]++
				i++
			}
		}
	}

	var all []string
	for key, val := range grams {
		v := val / total
		switch {
		case v > cutoff:
			all = append(all, fmt.Sprintf(`"%s": %0.4f`, key, v))
		default:
			// too small to matter
		}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i] < all[j]
	})
	var output []string
	for _, val := range chunk(all, 6) {
		output = append(output, strings.Join(val, ", "))
	}
	output[len(output)-1] += ","
	return "bigrams", strings.Join(output, ",\n")
}

func action(c *cli.Context) error {
	t, err := template.New("test").Parse(source)
	if err != nil {
		return err
	}

	words, err := qordle.Read("qordle")
	if err != nil {
		return err
	}

	data := map[string]any{"length": length, "cutoff": cutoff}
	for _, f := range []TableFunc{bigrams, frequencies, positions} {
		name, table := f(words)
		data[name] = table
	}

	var source bytes.Buffer
	err = t.Execute(&source, data)
	if err != nil {
		return err
	}
	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return err
	}
	for i := 0; i < c.NArg(); i++ {
		if err = func(name string) error {
			var fp *os.File
			fp, err = os.Create(name)
			if err != nil {
				return err
			}
			defer fp.Close()
			_, err = fp.WriteString(string(formatted))
			return err
		}(c.Args().Get(i)); err != nil {
			return err
		}
	}
	return err
}

func main() {
	app := &cli.App{
		Name:        "letters",
		HelpName:    "letters",
		Usage:       "generates the letter frequency tables",
		Description: "generates the letter frequency tables",
		Action:      action,
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				log.Error().Err(v).Msg(app.Name)
			case string:
				log.Error().Err(errors.New(v)).Msg(app.Name)
			default:
				log.Error().Err(fmt.Errorf("%v", v)).Msg(app.Name)
			}
			os.Exit(1)
		}
		if err != nil {
			log.Error().Err(err).Msg(app.Name)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	err = app.RunContext(context.Background(), os.Args)
}
