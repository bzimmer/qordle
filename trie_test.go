package qordle_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

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
			trie := &qordle.Trie[any]{}
			for _, w := range tt.words {
				trie.Add(w, nil)
			}
			node := trie.Node(tt.pattern)
			a.Equal(tt.prefix, node.Prefix())
			a.Equal(tt.word, node.Word())
		})
	}
}

func TestStrings(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	trie := &qordle.Trie[any]{}
	for _, w := range []string{"foo", "bar", "baz", "barter"} {
		trie.Add(w, nil)
	}
	res := trie.Strings()
	a.NotNil(res)
	sort.Strings(res)
	a.Equal([]string{"bar", "barter", "baz", "foo"}, res)
}
