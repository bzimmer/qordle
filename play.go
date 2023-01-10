package qordle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const auto = "auto"

type Scoreboard struct {
	Secret   string   `json:"secret"`
	Strategy string   `json:"strategy"`
	Rounds   []*Round `json:"rounds"`
	Elapsed  int64    `json:"elapsed"`
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
func (g *Game) Play(ctx context.Context, secret string) (*Scoreboard, error) {
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
	return g.play(ctx, secret, words)
}

func (g *Game) play(ctx context.Context, secret string, words []string) (*Scoreboard, error) {
	scoreboard := &Scoreboard{
		Secret:   secret,
		Strategy: g.strategy.String(),
	}
	defer func(t time.Time) {
		scoreboard.Elapsed = time.Since(t).Milliseconds()
	}(time.Now())
	dictionary := g.dictionary
	upper := cases.Upper(language.English)
	fns := []FilterFunc{Length(len(secret)), IsLower()}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
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

func secrets(c *cli.Context) ([]string, error) {
	if c.Bool(auto) {
		log.Info().Msg("reading from stdin")
		return read(c.App.Reader)
	}
	return c.Args().Slice(), nil
}

func play(c *cli.Context) error {
	dictionary, err := wordlists(c, "possible", "solutions")
	if err != nil {
		return err
	}
	st, err := NewStrategy(c.String("strategy"))
	if err != nil {
		return err
	}
	sts, err := secrets(c)
	if err != nil {
		return err
	}
	if c.Bool("speculate") {
		st = NewSpeculator(dictionary, st)
	}
	game := NewGame(
		WithStrategy(st),
		WithDictionary(dictionary),
		WithStart(c.String("start")))
	enc := json.NewEncoder(c.App.Writer)
	for _, secret := range sts {
		if err = func() error {
			var board *Scoreboard
			ctx, cancel := context.WithTimeout(c.Context, time.Second*2)
			defer cancel()
			board, err = game.Play(ctx, secret)
			if err != nil {
				if !c.Bool(auto) {
					return err
				}
				if errors.Is(err, context.DeadlineExceeded) {
					log.Error().Str("secret", secret).Msg("game never converged")
					return nil
				}
				return err
			}
			return enc.Encode(board)
		}(); err != nil {
			return err
		}
	}
	return nil
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
			&cli.BoolFlag{
				Name:    "speculate",
				Aliases: []string{"S"},
				Usage:   "speculate if necessary",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    auto,
				Aliases: []string{"A"},
				Usage:   "auto play from stdin",
				Value:   false,
			},
			wordlistFlag(),
		},
		Before: func(c *cli.Context) error {
			if !c.Bool(auto) && c.NArg() == 0 {
				return fmt.Errorf("expected at least one word to play")
			}
			return nil
		},
		Action: play,
	}
}
