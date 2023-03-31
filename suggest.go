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
			dictionary, strategy, err := prepare(c, "possible", "solutions")
			if err != nil {
				return err
			}
			guesser, err := Guess(c.Args().Slice()...)
			if err != nil {
				return err
			}
			dictionary = Filter(
				dictionary, IsLower(), Length(c.Int("length")), guesser)
			dictionary = strategy.Apply(dictionary)
			return Runtime(c).Encoder.Encode(dictionary)
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
			guesser, err := Guess(scores...)
			if err != nil {
				return err
			}
			return Runtime(c).Encoder.Encode(map[string]any{
				"word":   word,
				"scores": scores,
				"ok":     guesser(word),
			})
		},
	}
}
