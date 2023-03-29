package qordle

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type Mark int
type Marks []Mark

const (
	yellow = '.'

	MarkMiss      Mark = 0
	MarkMisplaced Mark = 1
	MarkExact     Mark = 2
)

func Check(secret string, guesses ...string) []Marks {
	secret = strings.ToLower(secret)
	scores := make([]Marks, len(guesses))
	for n, guess := range guesses {
		if len(secret) != len(guess) {
			log.Error().Str("secret", secret).Str("guess", guess).Msg("score")
			panic("secret and guess lengths do not match")
		}
		guess = strings.ToLower(guess)
		score := make(Marks, len(secret))
		round := make(map[string]int, len(secret))
		// first pass checks for exact matches
		for i := range guess {
			switch {
			case secret[i] == guess[i]:
				score[i] = MarkExact
			default:
				round[string(secret[i])]++
			}
		}
		// second pass checks for misplaced matches
		for i := range guess {
			if score[i] != MarkExact {
				m := string(guess[i])
				switch val := round[m]; val {
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
	return scores
}

func Score(secret string, guesses ...string) []string {
	checks := Check(secret, guesses...)
	scores := make([]string, len(checks))
	for i := range checks {
		var pattern string
		score, guess := checks[i], guesses[i]
		for j := range score {
			switch score[j] {
			case MarkExact:
				pattern += strings.ToUpper(string(guess[j]))
			case MarkMiss:
				pattern += strings.ToLower(string(guess[j]))
			case MarkMisplaced:
				pattern += fmt.Sprintf("%c%s", yellow, strings.ToLower(string(guess[j])))
			}
		}
		scores[i] = pattern
	}
	return scores
}

func CommandScore() *cli.Command {
	return &cli.Command{
		Name:      "score",
		Usage:     "score the guesses against the secret",
		ArgsUsage: "<secret> <guess> [, <guess>]",
		Action: func(c *cli.Context) error {
			scores := Score(c.Args().First(), c.Args().Tail()...)
			return Runtime(c).Encoder.Encode(scores)
		},
	}
}
