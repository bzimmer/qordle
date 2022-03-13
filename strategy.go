package qordle

import (
	"sort"
)

type Strategy func(words Dictionary) Dictionary

// Alpha orders the dictionary alphabetically
func Alpha(words Dictionary) Dictionary {
	dict := make(Dictionary, len(words))
	copy(dict, words)
	sort.Strings(dict)
	return dict
}

// Frequency orders the dictionary by words containing the most frequent letters
func Frequency(words Dictionary) Dictionary {
	// find the most common letters in the word list
	m := make(map[rune]int)
	for i := range words {
		w := []rune(words[i])
		s := make(map[rune]bool, 0)
		for j := range w {
			if _, ok := s[w[j]]; !ok {
				s[w[j]] = true
				m[w[j]]++
			}
		}
	}

	// map each word to its sum of letters (skip duplicates)
	x := make(map[int][]string)
	for i := range words {
		n := 0
		w := []rune(words[i])
		s := make(map[rune]bool, 0)
		for j := range w {
			if _, ok := s[w[j]]; !ok {
				s[w[j]] = true
				n += m[w[j]]
			}
		}
		x[n] = append(x[n], words[i])
	}

	// sort the words by their letter sums
	ranks := make([]int, 0, len(x))
	for k := range x {
		ranks = append(ranks, k)
	}
	sort.Ints(ranks)

	// construct the new dictionary
	dict := make(Dictionary, 0)
	for i := len(ranks) - 1; i >= 0; i-- {
		// alpha sort to ensure stability in the output
		q := x[ranks[i]]
		sort.Strings(q)
		dict = append(dict, q...)
	}

	return dict
}
