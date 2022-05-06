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
				if _, err := score.WriteString(fmt.Sprintf("%c%c", YellowPrefix, guess[index])); err != nil {
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

type game struct {
	start    string
	words    Dictionary
	strategy Strategy
}

func (g *game) play(secret string) ([]string, error) {
	dict := g.words
	words := []string{g.start}
	fns := []FilterFunc{Length(len(secret)), Lower()}
	for {
		scores, err := Score(secret, words...)
		if err != nil {
			return nil, err
		}
		guesses, err := Guesses(scores...)
		if err != nil {
			return nil, err
		}
		dict = g.strategy.Apply(Filter(dict, append(fns, guesses)...))
		log.Info().
			Int("dict", len(dict)).
			Str("secret", secret).
			Str("next", func() string {
				switch {
				case len(dict) == 0:
					return ""
				default:
					return dict[0]
				}
			}()).
			Strs("scores", scores).
			Strs("words", words).
			Msg("play")
		switch {
		case len(dict) == 0:
			return nil, fmt.Errorf("failed to find secret")
		case dict[0] == secret:
			if len(scores) == 1 {
				return scores, nil
			}
			return append(scores, strings.ToUpper(dict[0])), nil
		}
		words = append(words, dict[0])
	}
}

func CommandPlay() *cli.Command {
	return &cli.Command{
		Name:  "play",
		Usage: "play wordle automatically",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "start",
				Aliases: []string{"t"},
				Value:   "soare",
			},
			&cli.StringSliceFlag{
				Name:    "wordlist",
				Aliases: []string{"w"},
				Usage:   "use the specified embedded word list",
				Value:   nil,
			},
			&cli.StringFlag{
				Name:    "strategy",
				Aliases: []string{"s"},
				Usage:   "use the specified strategy",
				Value:   "frequency",
			},
		},
		Before: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("expected at least one word to play")
			}
			if !c.IsSet("wordlist") || len(c.StringSlice("wordlist")) == 0 {
				if err := c.Set("wordlist", "possible"); err != nil {
					return err
				}
				if err := c.Set("wordlist", "solutions"); err != nil {
					return err
				}
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			var words Dictionary
			for _, wordlist := range c.StringSlice("wordlist") {
				t, err := DictionaryEm(fmt.Sprintf("%s.txt", wordlist))
				if err != nil {
					return err
				}
				words = append(words, t...)
			}
			st, err := strategy(c.String("strategy"))
			if err != nil {
				return err
			}
			gm := &game{
				start:    c.String("start"),
				words:    words,
				strategy: st,
			}
			enc := json.NewEncoder(c.App.Writer)
			for _, secret := range c.Args().Slice() {
				words, err := gm.play(secret)
				if err != nil {
					return err
				}
				if err := enc.Encode(words); err != nil {
					return err
				}
			}
			return nil
		},
	}
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
