package qordle

import (
	"github.com/urfave/cli/v2"
)

func CommandSuggest() *cli.Command {
	return &cli.Command{
		Name:        "suggest",
		Usage:       "suggest the next guess",
		Description: "suggest the next guess",
		Flags: append(
			[]cli.Flag{
				&cli.IntFlag{
					Name:  "length",
					Usage: "word length",
					Value: 5,
				},
			},
			append(wordlistFlags(), strategyFlags()...)...,
		),
		Action: func(c *cli.Context) error {
			guess, err := Guess(c.Args().Slice()...)
			if err != nil {
				return err
			}
			dictionary, strategy, err := prepare(c, "possible", "solutions")
			if err != nil {
				return err
			}
			dictionary = Filter(dictionary, IsLower(), Length(c.Int("length")), guess)
			return Runtime(c).Encoder.Encode(strategy.Apply(dictionary))
		},
	}
}

func CommandValidate() *cli.Command {
	return &cli.Command{
		Name:      "validate",
		Usage:     "validate the word against the pattern",
		ArgsUsage: "<word to validate> <scored word> [, <scored word>]",
		Action: func(c *cli.Context) error {
			word := c.Args().First()
			scores := c.Args().Tail()
			guess, err := Guess(scores...)
			if err != nil {
				return err
			}
			return Runtime(c).Encoder.Encode(map[string]any{
				"ok":     guess(word),
				"scores": scores,
				"word":   word,
			})
		},
	}
}

func CommandBooks() *cli.Command {
	return &cli.Command{
		Name:  "books",
		Usage: "detailed information from google books",
		Description: `Sum all the percentages for letters in position 2:

	$ qordle books | jq '.positions | flatten | map(."2") | add'

Compute the score for a word:

	$ qordle books brown | jq .words
	{
		"brown": {
		  "frequencies": {
			"ranks": {
			  "0": 1.48,
			  "1": 6.28,
			  "2": 7.64,
			  "3": 1.68,
			  "4": 7.23
			},
			"total": 24.31
		  },
		  "positions": {
			"ranks": {
			  "0": 4.434,
			  "1": 5.891,
			  "2": 5.97,
			  "3": 1.107,
			  "4": 6.536
			},
			"total": 23.937999999999995
		  }
		}
	}`,
		Action: func(c *cli.Context) error {
			words := make(map[string]any, c.NArg())
			for i := 0; i < c.NArg(); i++ {
				var pt, ft float64
				w := []rune(c.Args().Get(i))
				p := make(map[int]float64, len(w))
				f := make(map[int]float64, len(w))
				for j := range w {
					p[j] = positions[w[j]][j]
					pt += p[j]
					f[j] = frequencies[w[j]]
					ft += f[j]
				}
				words[string(w)] = map[string]any{
					"positions":   map[string]any{"ranks": p, "total": pt},
					"frequencies": map[string]any{"ranks": f, "total": ft},
				}
			}
			return Runtime(c).Encoder.Encode(map[string]any{
				"bigrams":     bigrams,
				"frequencies": frequencies,
				"positions":   positions,
				"words":       words,
			})
		},
	}
}
