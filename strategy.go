package qordle

import (
	"fmt"
	"sort"

	"github.com/rs/zerolog/log"
)

func strategy(code string) (Strategy, error) {
	switch code {
	case "a", "alpha":
		return new(Alpha), nil
	case "p", "pos", "position":
		return new(Position), nil
	case "f", "freq", "frequency":
		return new(Frequency), nil
	}
	return nil, fmt.Errorf("unknown strategy `%s`", code)
}

type Strategy interface {
	Apply(words Dictionary) Dictionary
}

// Alpha orders the dictionary alphabetically
type Alpha struct{}

func (s *Alpha) Apply(words Dictionary) Dictionary {
	dict := make(Dictionary, len(words))
	copy(dict, words)
	sort.Strings(dict)
	return dict
}

func mkdict(op string, scores map[int][]string) Dictionary {
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
		log.Debug().Int("score", ranks[i]).Strs("words", q).Msg(op)
		dict = append(dict, q...)
	}
	return dict
}

// Position orders words by their letter's optimal position
type Position struct{}

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

	if log.Debug().Enabled() {
		for letter, val := range pos {
			for index, count := range val {
				log.Debug().
					Str("letter", string(letter)).
					Int("index", index).
					Int("count", count).
					Msg("position")
			}
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
