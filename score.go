package qordle

import (
	"errors"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Mark is the state for a letter
type Mark int

// Marks is a slice of marks
type Marks []Mark

const (
	yellow = '.'

	MarkMiss      Mark = 0
	MarkMisplaced Mark = 1
	MarkExact     Mark = 2
)

var ErrInvalidLength = errors.New("secret and guess lengths do not match")

func Check(secret string, guesses ...string) ([]Marks, error) {
	secret = strings.ToLower(secret)
	scores := make([]Marks, len(guesses))
	for n, guess := range guesses {
		if len(secret) != len(guess) {
			log.Error().Str("secret", secret).Str("guess", guess).Msg("score")
			return nil, ErrInvalidLength
		}
		guess = strings.ToLower(guess)
		score := make(Marks, len(secret))
		round := make(map[byte]int, len(secret))
		// first pass checks for exact matches
		for i := range guess {
			if secret[i] == guess[i] {
				score[i] = MarkExact
			} else {
				round[secret[i]]++
			}
		}
		// second pass checks for misplaced matches
		for i := range guess {
			if score[i] != MarkExact {
				m := guess[i]
				switch round[m] {
				case 0:
					// this letter doesn't exist in the secret
					score[i] = MarkMiss
				default:
					round[m]--
					score[i] = MarkMisplaced
				}
			}
		}
		scores[n] = score
	}
	return scores, nil
}

func Score(secret string, guesses ...string) ([]string, error) {
	checks, err := Check(secret, guesses...)
	if err != nil {
		return nil, err
	}
	scores := make([]string, len(checks))
	for i := range checks {
		var pattern []rune
		score, guess := checks[i], []rune(guesses[i])
		for j := range score {
			switch score[j] {
			case MarkExact:
				pattern = append(pattern, unicode.ToUpper(guess[j]))
			case MarkMiss:
				pattern = append(pattern, unicode.ToLower(guess[j]))
			case MarkMisplaced:
				pattern = append(pattern, yellow, unicode.ToLower(guess[j]))
			}
		}
		scores[i] = string(pattern)
	}
	return scores, nil
}

func CommandScore() *cli.Command {
	return &cli.Command{
		Name:      "score",
		Category:  "wordle",
		Usage:     "Score the guesses against the secret",
		ArgsUsage: "<secret> <guess> [, <guess>]",
		Action: func(c *cli.Context) error {
			scores, err := Score(c.Args().First(), c.Args().Tail()...)
			if err != nil {
				return err
			}
			return Runtime(c).Encoder.Encode(scores)
		},
	}
}
