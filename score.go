package qordle

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Score generates the scoreboard after each guess of the secret word
func Score(secret string, guesses ...string) ([]string, error) {
	log.Debug().Str("secret", secret).Strs("guesses", guesses).Msg("score")
	scores := make([]string, 0)
	for _, guess := range guesses {
		if len(secret) != len(guess) {
			return nil, errors.New("secret and guess lengths do not match")
		}
		var score strings.Builder
		guess = strings.ToLower(guess)
		for index := range guess {
			switch {
			case secret[index] == guess[index]:
				if _, err := score.WriteString(strings.ToUpper(string(guess[index]))); err != nil {
					return nil, err
				}
			case strings.Contains(secret, string(guess[index])):
				if _, err := score.WriteString(fmt.Sprintf("%c%c", yellow, guess[index])); err != nil {
					return nil, err
				}
			default:
				if err := score.WriteByte(guess[index]); err != nil {
					return nil, err
				}
			}
		}
		scores = append(scores, score.String())
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
