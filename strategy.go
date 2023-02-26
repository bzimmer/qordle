package qordle

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func strategyFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "strategy",
			Aliases: []string{"s"},
			Usage:   "use the specified strategy",
			Value:   cli.NewStringSlice("frequency"),
		},
		&cli.BoolFlag{
			Name:    "speculate",
			Aliases: []string{"S"},
			Usage:   "speculate if necessary",
			Value:   false,
		},
	}
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

func mkdictf(scores map[string]float64, less func(i, j float64) bool) Dictionary {
	type tuple struct {
		word string
		rank float64
	}
	tuples := make([]tuple, 0, len(scores))
	for word, rank := range scores {
		tuples = append(tuples, tuple{word, rank})
	}
	sort.Slice(tuples, func(i, j int) bool {
		return less(tuples[i].rank, tuples[j].rank)
	})
	words := make(Dictionary, len(tuples))
	for i := range tuples {
		words[i] = tuples[i].word
	}
	return words
}

// Position sorts the word list be letter position
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

// Frequency sorts the wordlist by letter frequency
type Frequency struct{}

func (s *Frequency) String() string {
	return "frequency"
}

func (s *Frequency) Apply(words Dictionary) Dictionary {
	// find the most common letters in the word list
	freq := make(map[rune]int)
	for i := range words {
		word := []rune(words[i])
		for j := range word {
			freq[word[j]]++
		}
	}

	// map each word to its sum of letters (skip duplicates)
	scores := make(map[int][]string)
	for i, word := range words {
		n := 0
		word := []rune(word)
		s := make(map[rune]struct{}, len(word))
		for j := range word {
			if _, ok := s[word[j]]; !ok {
				s[word[j]] = struct{}{}
				n += freq[word[j]]
			}
		}
		scores[n] = append(scores[n], words[i])
	}

	return mkdict(scores)
}

// Bigram sorts the dictionary by the bigram frequency of the word
type Bigram struct{}

func (s *Bigram) String() string {
	return "bigram"
}

func (s *Bigram) Apply(words Dictionary) Dictionary {
	var i int
	var val float64
	res := make(map[string]float64, len(words))
	for _, word := range words {
		switch n := len(word); n {
		case 0, 1:
		default:
			i, val = 0, 0.0
			for i+2 < n {
				val += bigrams[word[i:i+2]]
				i++
			}
			res[word] = val
		}
	}
	if len(res) == 0 {
		return words
	}
	return mkdictf(res, func(i, j float64) bool {
		return i > j
	})
}

// Chain chains multiple strategies to sort the wordlist
type Chain struct {
	strategies []Strategy
}

func (s *Chain) String() string {
	names := make([]string, len(s.strategies))
	for i := range s.strategies {
		names[i] = s.strategies[i].String()
	}
	return fmt.Sprintf("chain{%s}", strings.Join(names, ","))
}

func (s *Chain) Apply(words Dictionary) Dictionary {
	switch n := len(s.strategies); n {
	case 0:
		return words
	case 1:
		return s.strategies[0].Apply(words)
	}

	var wg sync.WaitGroup
	wordc := make(chan Dictionary, len(s.strategies))
	for i := range s.strategies {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			wordc <- s.strategies[i].Apply(words)
		}(i)
	}
	go func() {
		defer close(wordc)
		wg.Wait()
	}()

	n := float64(len(words))
	res := make(map[string]float64, len(words))
	for w := range wordc {
		for i, w := range w {
			res[w] += float64(i) / n
		}
	}
	return mkdictf(res, func(i, j float64) bool {
		return i < j
	})
}

func NewChain(strategies ...Strategy) Strategy {
	return &Chain{strategies: strategies}
}

// Speculate attempts to find a word which eliminates the most letters
type Speculate struct {
	words       Dictionary
	strategy    Strategy
	speculation int
}

func (s *Speculate) String() string {
	if s.strategy == nil {
		return "speculate"
	}
	return fmt.Sprintf("speculate{%s}", s.strategy.String())
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
