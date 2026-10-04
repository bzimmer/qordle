package qordle

import (
	"math"
	"runtime"
	"slices"
	"sync"
)

// feedback encodes the marks for guess against secret as a base-3 number,
// matching Check without allocating. Both words must be ASCII and of equal
// length.
func feedback(secret, guess string) int {
	var exact [16]bool
	var unmatched [256]uint8
	for i := range len(guess) {
		if secret[i] == guess[i] {
			exact[i] = true
		} else {
			unmatched[secret[i]]++
		}
	}
	code, place := 0, 1
	for i := range len(guess) {
		switch {
		case exact[i]:
			code += int(MarkExact) * place
		case unmatched[guess[i]] > 0:
			unmatched[guess[i]]--
			code += int(MarkMisplaced) * place
		}
		place *= 3
	}
	return code
}

// entropy returns the expected information, in bits, revealed by guess when
// each of words is equally likely to be the secret, reusing counts as
// scratch space sized for every feedback code.
func entropy(words Dictionary, guess string, counts []int) float64 {
	clear(counts)
	for _, secret := range words {
		counts[feedback(secret, guess)]++
	}
	n := float64(len(words))
	var sum float64
	for _, c := range counts {
		if c > 1 {
			sum += float64(c) * math.Log2(float64(c))
		}
	}
	return math.Log2(n) - sum/n
}

// Probe puts a guess ahead of the ranked words when it is expected to reveal
// more than guessing a remaining word, even though it may not be a candidate.
// A guess scores the entropy of its feedback across the remaining words: one
// splitting them into many small groups of identical feedback beats one
// leaving a few large groups.
// Each remaining word earns a bonus for the chance of being the secret, so
// with only a few words left a candidate wins over a pure probe.
type Probe struct {
	guesses   Dictionary
	preferred map[string]struct{}
	strategy  Strategy
}

func (s *Probe) String() string {
	return "probe{" + s.strategy.String() + "}"
}

// candidates returns the words to plan against, the preferred words when any
// remain since those are the likely secrets.
func (s *Probe) candidates(words Dictionary) Dictionary {
	var res Dictionary
	for _, word := range words {
		if _, ok := s.preferred[word]; ok {
			res = append(res, word)
		}
	}
	if len(res) == 0 {
		return words
	}
	return res
}

// best returns the guess with the highest score against the candidates.
func (s *Probe) best(candidates Dictionary) string {
	length := len(candidates[0])
	buckets := int(math.Pow(3, float64(length)))
	isCandidate := make(map[string]struct{}, len(candidates))
	for _, word := range candidates {
		isCandidate[word] = struct{}{}
	}
	pool := make(Dictionary, 0, len(s.guesses)+len(candidates))
	pool = append(pool, candidates...)
	for _, word := range s.guesses {
		if _, ok := isCandidate[word]; !ok && len(word) == length {
			pool = append(pool, word)
		}
	}
	n := float64(len(candidates))
	bonus := math.Log2(n) / n
	scores := make([]float64, len(pool))
	indices := make(chan int)
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), len(pool)) {
		wg.Go(func() {
			counts := make([]int, buckets)
			for i := range indices {
				scores[i] = entropy(candidates, pool[i], counts)
				if i < len(candidates) {
					scores[i] += bonus
				}
			}
		})
	}
	for i := range pool {
		indices <- i
	}
	close(indices)
	wg.Wait()
	best := 0
	for i := range scores {
		// strictly greater keeps the earliest, so candidates win ties
		if scores[i] > scores[best] {
			best = i
		}
	}
	return pool[best]
}

func (s *Probe) Apply(words Dictionary) Dictionary {
	ranked := s.strategy.Apply(words)
	candidates := s.candidates(words)
	if len(candidates) <= 2 || len(candidates[0]) > 10 {
		return ranked
	}
	for _, word := range candidates {
		if len(word) != len(candidates[0]) {
			return ranked
		}
	}
	best := s.best(candidates)
	if slices.Contains(candidates, best) {
		return ranked
	}
	return append(Dictionary{best}, ranked...)
}

// NewProbe returns a strategy which may lead with any of guesses, planning
// against the preferred words while they remain.
func NewProbe(guesses, preferred Dictionary, strategy Strategy) Strategy {
	set := make(map[string]struct{}, len(preferred))
	for _, word := range preferred {
		set[word] = struct{}{}
	}
	return &Probe{guesses: guesses, preferred: set, strategy: strategy}
}
