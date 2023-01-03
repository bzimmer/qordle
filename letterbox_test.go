package qordle_test

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
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
	for _, tt := range []struct {
		name    string
		options []string
		err     bool
		length  int
	}{
		{
			name:    "letterbox with 0 arguments",
			length:  3,
			options: []string{"-w", "solutions", "--max", "4"},
		},
		{
			name:    "letterbox with 1 argument",
			length:  57,
			options: []string{"-w", "solutions", "--max", "4", "rul-eya-gdh-opb"},
		},
		{
			name:    "letterbox with 4 arguments",
			length:  57,
			options: []string{"-w", "solutions", "--max", "4", "rul", "eya", "gdh", "opb"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			var builder strings.Builder
			app := &cli.App{
				Name:     tt.name,
				Writer:   &builder,
				Commands: []*cli.Command{qordle.CommandLetterBox()},
			}
			err := app.Run(append([]string{"qordle", "letterbox"}, tt.options...))
			if tt.err {
				a.Error(err)
				return
			}
			a.NoError(err)
			var n int
			dec := json.NewDecoder(strings.NewReader(builder.String()))
			for {
				var res []string
				if err = dec.Decode(&res); errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					a.Error(err)
				}
				n++
			}
			a.Equal(tt.length, n)
		})
	}
}
