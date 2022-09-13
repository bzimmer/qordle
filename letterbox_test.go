package qordle_test

import (
	"testing"

	"github.com/bzimmer/qordle"
	"github.com/stretchr/testify/assert"
)

func TestTrie(t *testing.T) {
	for _, tt := range []struct {
		name, pattern string
		words         []string
		prefix, word  bool
	}{
		{
			name:    "add",
			pattern: "foo",
			prefix:  false,
			word:    true,
			words:   []string{"foo"},
		},
		{
			name:    "add",
			pattern: "foo",
			prefix:  true,
			word:    true,
			words:   []string{"foo", "food"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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
