package qordle

import (
	"sort"
)

func Suggest(words Dictionary) Dictionary {
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

	// map each word to it's sum of letters skipping duplicates
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

	// sort the words by their letter summation
	ranks := make([]int, 0, len(x))
	for k := range x {
		ranks = append(ranks, k)
	}
	sort.Ints(ranks)

	// construct the new dictionary
	dict := make(Dictionary, 0)
	for i := len(ranks) - 1; i >= 0; i-- {
		dict = append(dict, x[ranks[i]]...)
	}

	return dict
}
