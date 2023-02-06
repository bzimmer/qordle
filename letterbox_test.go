package qordle_test

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestTrie(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, pattern string
		words         []string
		prefix, word  bool
	}{
		{
			name:    "whole word",
			pattern: "foo",
			prefix:  false,
			word:    true,
			words:   []string{"foo"},
		},
		{
			name:    "prefix",
			pattern: "foo",
			prefix:  true,
			word:    true,
			words:   []string{"foo", "food"},
		},
		{
			name:    "nothing",
			pattern: "bar",
			prefix:  false,
			word:    false,
			words:   []string{"foo"},
		},
		{
			name:    "no words",
			pattern: "bar",
			prefix:  false,
			word:    false,
			words:   []string{},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			trie := qordle.NewTrie()
			for _, w := range tt.words {
				trie.Add(w)
			}
			node := trie.Node(tt.pattern)
			a.Equal(tt.prefix, node.Prefix())
			a.Equal(tt.word, node.Word())
		})
	}
}

func TestLetterboxCommand(t *testing.T) {
	a := assert.New(t)
	decode := func(c *cli.Context) [][]string {
		var results [][]string
		dec := json.NewDecoder(c.App.Writer.(io.Reader))
		for {
			var res []string
			err := dec.Decode(&res)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return results
				}
				a.FailNow(err.Error())
			}
			results = append(results, res)
		}
	}
	for _, tt := range []harness{
		{
			name: "letterbox with 0 arguments",
			args: []string{"letterbox", "-w", "solutions", "--max", "4"},
			after: func(c *cli.Context) error {
				a.Equal(3, len(decode(c)))
				return nil
			},
		},
		{
			name: "letterbox with 1 argument",
			args: []string{"letterbox", "-w", "solutions", "--max", "4", "rul-eya-gdh-opb"},
			after: func(c *cli.Context) error {
				a.Equal(57, len(decode(c)))
				return nil
			},
		},
		{
			name: "letterbox with 4 arguments",
			args: []string{"letterbox", "-w", "solutions", "--max", "4", "rul", "eya", "gdh", "opb"},
			after: func(c *cli.Context) error {
				a.Equal(57, len(decode(c)))
				return nil
			},
		},
		{
			name: "letterbox with invalid wordlist",
			args: []string{"letterbox", "-w", "missing", "--max", "4"},
			err:  "invalid wordlist `missing`",
		},
		{
			name: "letterbox with 5 arguments",
			args: []string{"letterbox", "-w", "solutions", "--max", "4", "rul", "eya", "gdh", "opb", "qst"},
			err:  "found 5 sides",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandLetterBox)
		})
	}
}
