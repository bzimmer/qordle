package qordle

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

const yellow = '~'

func Score(secret string, guesses ...string) ([]string, error) {
	secret = strings.ToLower(secret)
	scores := make([]string, len(guesses))
	for n, guess := range guesses {
		if len(secret) != len(guess) {
			return nil, errors.New("secret and guess lengths do not match")
		}
		pass := map[string]int{}
		guess = strings.ToLower(guess)
		score := make([]string, len(secret))
		// first pass checks for exact matches
		for i := range guess {
			m := string(guess[i])
			if secret[i] == guess[i] {
				score[i] = strings.ToUpper(m)
			} else {
				pass[string(secret[i])]++
			}
		}
		// second pass checks for inexact matches
		for i := range guess {
			if score[i] != "" {
				continue
			}
			m := string(guess[i])
			switch val := pass[m]; val {
			case 0:
				// this letter doesn't exist in the secret
				score[i] = m
			default:
				pass[m]--
				score[i] = fmt.Sprintf("%c%s", yellow, m)
			}
		}
		// construct the score
		scores[n] = strings.Join(score, "")
	}
	return scores, nil
}

func CommandScore() *cli.Command {
	return &cli.Command{
		Name:      "score",
		Usage:     "score the guesses against the secret",
		ArgsUsage: "<secret> <guess> [, <guess>]",
		Action: func(c *cli.Context) error {
			scores, err := Score(c.Args().First(), c.Args().Tail()...)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(scores)
		},
	}
}
