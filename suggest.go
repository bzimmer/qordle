package qordle

import (
	"github.com/urfave/cli/v2"
)

func CommandSuggest() *cli.Command {
	return &cli.Command{
		Name:     "suggest",
		Category: "wordle",
		Usage:    "Suggest the next guess",
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
		Category:  "wordle",
		Usage:     "Validate the word against the pattern",
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

func CommandRanks() *cli.Command {
	return &cli.Command{
		Name:     "ranks",
		Category: "wordle",
		Usage:    "Detailed rank information from letter frequency tables",
		Description: `Sum all the percentages for letters in position 2:

	$ qordle ranks | jq '.positions | flatten | map(."2") | add'

Compute the score for a word:

	$ qordle ranks brown | jq .words
	{
		"brown": {
		  "frequencies": {
			"ranks": {
			  "0": 0.0183,
			  "1": 0.0704,
			  "2": 0.0720,
			  "3": 0.0065,
			  "4": 0.0718
			},
			"total": 0.2390
		  },
		  "positions": {
			"ranks": {
			  "0": 0.0698,
			  "1": 0.0649,
			  "2": 0.0638,
			  "3": 0.0101,
			  "4": 0.0644
			},
			"total": 0.2730
		  }
		}
	}`,
		Action: func(c *cli.Context) error {
			words := make(map[string]any, c.NArg())
			for i := 0; i < c.NArg(); i++ {
				var bt, ft, pt float64
				word := c.Args().Get(i)
				w := []rune(word)
				p := make(map[int]float64, len(w))
				f := make(map[int]float64, len(w))
				b := make(map[string]float64, len(w))
				for j := range w {
					p[j] = positions[w[j]][j]
					pt += p[j]
					f[j] = frequencies[w[j]]
					ft += f[j]

					k := 0
					for k+2 <= len(word) {
						bt += bigrams[word[k:k+2]]
						b[word[k:k+2]] = bigrams[word[k:k+2]]
						k++
					}
				}
				words[string(w)] = map[string]any{
					"bigrams":     map[string]any{"ranks": b, "total": bt},
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

func CommandOrder() *cli.Command {
	return &cli.Command{
		Name:      "order",
		Category:  "wordle",
		Usage:     "Order the arguments per the strategy",
		ArgsUsage: "word [, word, ...]",
		Flags:     strategyFlags(),
		Action: func(c *cli.Context) error {
			dictionary := Dictionary(c.Args().Slice())
			_, strategy, err := prepare(c)
			if err != nil {
				return err
			}
			return Runtime(c).Encoder.Encode(strategy.Apply(dictionary))
		},
	}
}
