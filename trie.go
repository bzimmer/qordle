package qordle

import "strings"

type Trie[T any] struct {
	word     bool
	children map[rune]*Trie[T]
	value    T
}

func (trie *Trie[T]) Add(word string, value T) {
	node := trie
	for _, r := range strings.ToLower(word) {
		child := node.children[r]
		if child == nil {
			if node.children == nil {
				node.children = make(map[rune]*Trie[T])
			}
			child = new(Trie[T])
			node.children[r] = child
		}
		node = child
	}
	node.word = true
	node.value = value
}

func (trie *Trie[T]) Node(word string) *Trie[T] {
	node := trie
	for _, r := range word {
		child := node.children[r]
		if child == nil {
			return nil
		}
		node = child
	}
	return node
}

func (trie *Trie[T]) Value(word string) T {
	node := trie.Node(word)
	switch {
	case node == nil:
		return *new(T)
	case node.word:
		return node.value
	default:
		for {
			switch n := len(node.children); n {
			case 1:
				// pop the only child
				for _, child := range node.children {
					node = child
				}
			default:
				return node.value
			}
		}
	}
}

func (trie *Trie[T]) Prefix() bool {
	return trie != nil && len(trie.children) > 0
}

func (trie *Trie[T]) Word() bool {
	return trie != nil && trie.word
}