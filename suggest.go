package qordle

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
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
				Name:  "dictionary",
				Usage: "dictionary of possible words (eg /usr/share/dict/words)",
			},
			&cli.StringFlag{
				Name:    "wordlist",
				Aliases: []string{"w"},
				Usage:   "use the specified embedded word list",
				Value:   "solutions",
			},
			&cli.StringFlag{
				Name:    "strategy",
				Aliases: []string{"s"},
				Usage:   "use the specified strategy",
				Value:   "frequency",
			},
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

			var source string
			var dictionary Dictionary
			if c.IsSet("dictionary") {
				source = c.String("dictionary")
				dictionary, err = DictionaryFs(afero.NewOsFs(), source)
			} else {
				dictionary, err = DictionaryEm(c.String("wordlist"))
			}
			if err != nil {
				return err
			}

			n := len(dictionary)
			dictionary = Filter(dictionary, fns...)
			q := len(dictionary)

			st, err := NewStrategy(c.String("strategy"))
			if err != nil {
				return err
			}
			dictionary = st.Apply(dictionary)
			if debug && log.Debug().Enabled() {
				log.Debug().
					Dur("elapsed", time.Since(t)).
					Int("master", n).
					Int("filtered", q).
					Str("source", source).
					Msg("dictionary")
			}
			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(dictionary)
		},
	}
}
