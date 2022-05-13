package qordle

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Game struct {
	start      string
	strategy   Strategy
	dictionary Dictionary
}

// Option provides a configuration mechanism for a Game
type Option func(*Game)

// NewGame creates a new client and applies all provided Options
func NewGame(opts ...Option) *Game {
	g := new(Game)
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// WithStart is the first word to use
func WithStart(start string) Option {
	return func(g *Game) {
		g.start = start
	}
}

// WithDictionary is the dictionary to use
func WithDictionary(dictionary Dictionary) Option {
	return func(g *Game) {
		g.dictionary = dictionary
	}
}

// WithStrategy is the strategy to use
func WithStrategy(strategy Strategy) Option {
	return func(g *Game) {
		g.strategy = strategy
	}
}

// Play the game for the secret
func (g *Game) Play(secret string) ([]string, error) {
	if g.strategy == nil {
		return nil, errors.New("missing strategy")
	}
	if len(g.dictionary) == 0 {
		return nil, errors.New("missing dictionary")
	}
	var words []string
	switch g.start {
	case "":
		words = []string{g.strategy.Apply(g.dictionary)[0]}
	default:
		words = strings.Split(g.start, ",")
	}
	return game(g.dictionary, g.strategy, words, secret)
}

func game(dictionary Dictionary, strategy Strategy, words []string, secret string) ([]string, error) {
	upper := cases.Upper(language.English)
	fns := []FilterFunc{Length(len(secret)), IsLower()}
	for {
		scores, err := Score(secret, words...)
		if err != nil {
			return nil, err
		}
		guesses, err := Guesses(scores...)
		if err != nil {
			return nil, err
		}
		dictionary = strategy.Apply(Filter(dictionary, append(fns, guesses)...))
		log.Info().
			Int("dictionary", len(dictionary)).
			Str("secret", secret).
			Str("next", func() string {
				switch {
				case len(dictionary) == 0:
					return ""
				default:
					return dictionary[0]
				}
			}()).
			Strs("scores", scores).
			Strs("words", words).
			Msg("play")
		switch {
		case len(dictionary) == 0:
			return nil, fmt.Errorf("failed to find secret")
		case dictionary[0] == secret:
			if len(scores) == 1 {
				return scores, nil
			}
			return append(scores, upper.String(dictionary[0])), nil
		}
		words = append(words, dictionary[0])
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
				t, err := DictionaryEm(wordlist)
				if err != nil {
					return err
				}
				words = append(words, t...)
			}
			st, err := NewStrategy(c.String("strategy"))
			if err != nil {
				return err
			}
			game := NewGame(
				WithStrategy(st),
				WithDictionary(words),
				WithStart(c.String("start")))
			enc := json.NewEncoder(c.App.Writer)
			for _, secret := range c.Args().Slice() {
				words, err := game.Play(secret)
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
