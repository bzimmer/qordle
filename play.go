package qordle

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Scoreboard struct {
	Secret   string   `json:"secret"`
	Strategy string   `json:"strategy"`
	Rounds   []*Round `json:"rounds"`
}

type Round struct {
	Dictionary int      `json:"dictionary"`
	Next       string   `json:"next,omitempty"`
	Scores     []string `json:"scores"`
	Words      []string `json:"words"`
	Success    bool     `json:"success"`
}

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
func (g *Game) Play(secret string) (*Scoreboard, error) {
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
	return g.play(secret, words)
}

func (g *Game) play(secret string, words []string) (*Scoreboard, error) {
	scoreboard := &Scoreboard{
		Secret:   secret,
		Strategy: g.strategy.String(),
	}
	dictionary := g.dictionary
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
		round := &Round{
			Dictionary: len(dictionary),
			Scores:     scores,
			Words:      words,
		}
		dictionary = g.strategy.Apply(Filter(dictionary, append(fns, guesses)...))
		scoreboard.Rounds = append(scoreboard.Rounds, round)
		switch {
		case len(dictionary) == 0:
			return scoreboard, nil
		case dictionary[0] == secret:
			switch {
			case len(scores) == 1 && len(dictionary) == 1:
				// guessed the word immediately!
			default:
				scoreboard.Rounds = append(scoreboard.Rounds, &Round{
					Dictionary: len(dictionary),
					Scores:     append(scores, upper.String(dictionary[0])),
					Words:      append(words, dictionary[0]),
					Success:    true,
				})
			}
			return scoreboard, nil
		default:
			round.Next = dictionary[0]
			words = append(words, dictionary[0])
		}
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
			&cli.StringFlag{
				Name:    "strategy",
				Aliases: []string{"s"},
				Usage:   "use the specified strategy",
				Value:   "frequency",
			},
			wordlistFlag(),
		},
		Before: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("expected at least one word to play")
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			dictionary, err := wordlists(c, "possible", "solutions")
			if err != nil {
				return err
			}
			st, err := NewStrategy(c.String("strategy"))
			if err != nil {
				return err
			}
			game := NewGame(
				WithStrategy(st),
				WithDictionary(dictionary),
				WithStart(c.String("start")))
			enc := json.NewEncoder(c.App.Writer)
			for _, secret := range c.Args().Slice() {
				var rounds *Scoreboard
				rounds, err = game.Play(secret)
				if err != nil {
					return err
				}
				if err = enc.Encode(rounds); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
