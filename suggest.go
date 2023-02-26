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
				&cli.StringFlag{
					Name:  "hits",
					Usage: "letters in the word",
				},
				&cli.StringFlag{
					Name:  "misses",
					Usage: "letters not in the word",
				},
				&cli.StringFlag{
					Name:  "pattern",
					Usage: "match against a known pattern (use '.' for all unknown letters)",
				},
			},
			append(wordlistFlags(), strategyFlags()...)...,
		),
		Action: func(c *cli.Context) error {
			dictionary, strategy, err := prepare(c, "possible", "solutions")
			if err != nil {
				return err
			}
			pattern, err := Pattern(c.String("pattern"))
			if err != nil {
				return err
			}
			guesses, err := Guesses(c.Args().Slice()...)
			if err != nil {
				return err
			}
			dictionary = Filter(
				dictionary,
				IsLower(),
				Length(c.Int("length")),
				Hits(c.String("hits")),
				Misses(c.String("misses")),
				pattern,
				guesses)
			dictionary = strategy.Apply(dictionary)
			return Runtime(c).Encoder.Encode(dictionary)
		},
	}
}
