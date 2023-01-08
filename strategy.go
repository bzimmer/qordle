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

func mkdict(_ string, scores map[int][]string) Dictionary {
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

	return mkdict("position", scores)
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

	return mkdict("frequency", scores)
}

// Speculate attempts to find a word which eliminates the most letters
type Speculate struct {
	words      Dictionary
	strategy   Strategy
	speculated bool
}

func (s *Speculate) String() string {
	return "speculate"
}

func (s *Speculate) hamming(s1 string, s2 string) []rune {
	var runes []rune
	r1 := []rune(s1)
	r2 := []rune(s2)
	for i, v := range r1 {
		if r2[i] != v {
			runes = append(runes, v)
		}
	}
	return runes
}

func (s *Speculate) with(words Dictionary) Dictionary {
	runes := make(map[rune]struct{})
	for i := 1; i < len(words); i++ {
		r := s.hamming(words[i-1], words[i])
		if len(r) != 1 {
			// only speculate guessing games
			return nil
		}
		for x := range r {
			runes[r[x]] = struct{}{}
		}
	}

	n := 0
	next := make(map[int][]string)
	for _, word := range s.words {
		var q int
		for _, r := range word {
			if _, ok := runes[r]; ok {
				q++
			}
		}
		if q >= n {
			// this check is a small optimization
			//  only more complete words will be added
			n = q
			next[n] = append(next[n], word)
		}
	}

	return Dictionary(next[n])
}

func (s *Speculate) Apply(words Dictionary) Dictionary {
	switch {
	case s.speculated:
		return s.strategy.Apply(words)
	case len(words) == 1:
		return words
	}
	with := s.with(words)
	if with == nil {
		return s.strategy.Apply(words)
	}
	s.speculated = true
	with = s.strategy.Apply(with)
	log.Info().Strs("words", words).Strs("with", with).Msg("speculate")
	return append(with, s.strategy.Apply(words)...)
}
