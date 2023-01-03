package qordle

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func CommandSuggest() *cli.Command {
	return &cli.Command{
		Name:        "suggest",
		Usage:       "suggest the next guess",
		Description: "suggest the next guess",
		Flags: []cli.Flag{
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
			&cli.StringFlag{
				Name:    "strategy",
				Aliases: []string{"s"},
				Usage:   "use the specified strategy",
				Value:   "frequency",
			},
			wordlistFlag(),
		},
		Action: func(c *cli.Context) error {
			t := time.Now()
			pattern, err := Pattern(c.String("pattern"))
			if err != nil {
				return err
			}
			guesses, err := Guesses(c.Args().Slice()...)
			if err != nil {
				return err
			}
			fns := []FilterFunc{
				IsLower(),
				Length(c.Int("length")),
				Hits(c.String("hits")),
				Misses(c.String("misses")),
				pattern,
				guesses,
			}
			dictionary, err := wordlists(c, "possible", "solutions")
			if err != nil {
				return err
			}
			n := len(dictionary)
			dictionary = Filter(dictionary, fns...)
			q := len(dictionary)
			log.Debug().
				Dur("elapsed", time.Since(t)).
				Int("master", n).
				Int("filtered", q).
				Msg("dictionary")

			st, err := NewStrategy(c.String("strategy"))
			if err != nil {
				return err
			}
			dictionary = st.Apply(dictionary)
			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(dictionary)
		},
	}
}
