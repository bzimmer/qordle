package qordle

import (
	"fmt"
	"sort"

	"github.com/rs/zerolog/log"
)

func NewStrategy(code string) (Strategy, error) {
	switch code {
	case "a", "alpha":
		return new(Alpha), nil
	case "p", "pos", "position":
		return new(Position), nil
	case "", "f", "freq", "frequency":
		return new(Frequency), nil
	}
	return nil, fmt.Errorf("unknown strategy `%s`", code)
}

type Strategy interface {
	String() string
	Apply(words Dictionary) Dictionary
}

// Alpha orders the dictionary alphabetically
type Alpha struct{}

func (s *Alpha) String() string {
	return "alpha"
}

func (s *Alpha) Apply(words Dictionary) Dictionary {
	dict := make(Dictionary, len(words))
	copy(dict, words)
	sort.Strings(dict)
	return dict
}

func mkdict(scores map[int][]string) Dictionary {
	// sort the words by their positional scores
	ranks := make([]int, 0, len(scores))
	for k := range scores {
		ranks = append(ranks, k)
	}
	sort.Ints(ranks)

	// construct the new dictionary
	dict := make(Dictionary, 0)
	for i := len(ranks) - 1; i >= 0; i-- {
		// alpha sort to ensure stability in the output
		q := scores[ranks[i]]
		sort.Strings(q)
		dict = append(dict, q...)
	}
	return dict
}

// Position orders words by their letter's optimal position
type Position struct{}

func (s *Position) String() string {
	return "position"
}

func (s *Position) Apply(words Dictionary) Dictionary {
	// count the number of times a letter appears at the position
	pos := make(map[rune]map[int]int)
	for _, word := range words {
		for index, letter := range []rune(word) {
			if _, ok := pos[letter]; !ok {
				pos[letter] = make(map[int]int)
			}
			pos[letter][index]++
		}
	}

	// score the word by summing the position count for each letter
	scores := make(map[int][]string)
	for _, word := range words {
		s := 0
		for index, letter := range []rune(word) {
			s += pos[letter][index]
		}
		scores[s] = append(scores[s], word)
	}

	return mkdict(scores)
}

// Frequency orders the dictionary by words containing the most frequent letters
type Frequency struct{}

func (s *Frequency) String() string {
	return "frequency"
}

func (s *Frequency) Apply(words Dictionary) Dictionary {
	// find the most common letters in the word list
	freq := make(map[rune]int)
	for i := range words {
		w := []rune(words[i])
		s := make(map[rune]bool, 0)
		for j := range w {
			if _, ok := s[w[j]]; !ok {
				s[w[j]] = true
				freq[w[j]]++
			}
		}
	}

	// map each word to its sum of letters (skip duplicates)
	scores := make(map[int][]string)
	for i, word := range words {
		n := 0
		word := []rune(word)
		s := make(map[rune]bool, 0)
		for j := range word {
			if _, ok := s[word[j]]; !ok {
				s[word[j]] = true
				n += freq[word[j]]
			}
		}
		scores[n] = append(scores[n], words[i])
	}

	return mkdict(scores)
}

// Speculate attempts to find a word which eliminates the most letters
type Speculate struct {
	words       Dictionary
	strategy    Strategy
	speculation int
}

func (s *Speculate) String() string {
	if s.strategy == nil {
		return "speculate;"
	}
	return "speculate;" + s.strategy.String()
}

func (s *Speculate) hamming(s1 string, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)
	// check the rune array as the lengths might differ after conversion
	if len(r1) != len(r2) {
		return -1
	}
	index := -1
	for i, v := range r1 {
		if r2[i] != v {
			switch {
			case index == -1:
				// no index has been seen yet
				index = i
			case index != i:
				// an index has been seen and this one is different
				return -1
			}
		}
	}
	return index
}

func (s *Speculate) with(words Dictionary) Dictionary {
	index := -1
	for i := 1; i < len(words); i++ {
		x := s.hamming(words[i-1], words[i])
		switch {
		case x == -1:
			// the words do not have a common index
			return nil
		case index == -1:
			// no index has been seen yet
			index = x
		case index != x:
			// an index has been seen and this one is different
			return nil
		}
	}

	runes := make(map[rune]struct{}, len(words))
	for i := range words {
		runes[rune(words[i][index])] = struct{}{}
	}

	n, next, length := 0, make(map[int][]string), len(words[0])
	for _, word := range s.words {
		// only use words of the same length
		if len(word) == length {
			var q int
			for _, r := range word {
				if _, ok := runes[r]; ok {
					q++
				}
			}
			if q >= n {
				// only add words with more information
				n = q
				next[n] = append(next[n], word)
			}
		}
	}

	return Dictionary(next[n])
}

func (s *Speculate) Apply(words Dictionary) Dictionary {
	if len(words) <= s.speculation || s.strategy == nil {
		return words
	}
	with := s.with(words)
	if len(with) == 0 {
		return s.strategy.Apply(words)
	}
	with = s.strategy.Apply(with)
	log.Debug().Strs("words", words).Strs("with", with).Msg(s.String())
	return append(with[:1], s.strategy.Apply(words)...)
}

func NewSpeculator(words Dictionary, strategy Strategy) Strategy {
	// four words was chosen emperically as the cut off for being useful
	const speculation = 4
	return &Speculate{words: words, strategy: strategy, speculation: speculation}
}
