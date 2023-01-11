package qordle

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"
)

type Scoreboard struct {
	Secret     string   `json:"secret"`
	Strategy   string   `json:"strategy"`
	Dictionary int      `json:"dictionary"`
	Rounds     []*Round `json:"rounds"`
	Elapsed    int64    `json:"elapsed"`
}

type Round struct {
	Dictionary int      `json:"dictionary"`
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
	start := g.start
	if start == "" {
		start = g.strategy.Apply(g.dictionary)[0]
	}
	return g.play(ctx, secret, start)
}

func (g *Game) play(ctx context.Context, secret string, start string) (*Scoreboard, error) {
	words := []string{start}
	dictionary := g.dictionary
	scoreboard := &Scoreboard{
		Secret:     secret,
		Strategy:   g.strategy.String(),
		Dictionary: len(dictionary),
	}
	defer func(t time.Time) {
		scoreboard.Elapsed = time.Since(t).Milliseconds()
	}(time.Now())
	fns := []FilterFunc{Length(len(secret)), IsLower()}

	for {
		select {
		case <-ctx.Done():
			return scoreboard, ctx.Err()
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
		dictionary = g.strategy.Apply(Filter(dictionary, append(fns, guesses)...))

		round := &Round{
			Dictionary: len(dictionary),
			Scores:     scores,
			Words:      words,
		}
		scoreboard.Rounds = append(scoreboard.Rounds, round)

		switch {
		case round.Dictionary == 0:
			return scoreboard, nil
		case round.Words[len(round.Words)-1] == secret:
			round.Success = true
			return scoreboard, nil
		default:
			for _, w := range dictionary {
				for j := 0; j < len(words) && w != ""; j++ {
					if words[j] == w {
						w = ""
					}
				}
				if w != "" {
					words = append(words, w)
					break
				}
			}
		}
	}
}

func secrets(c *cli.Context) ([]string, error) {
	if c.NArg() > 0 {
		return c.Args().Slice(), nil
	}
	return read(c.App.Reader)
}

func play(c *cli.Context) error {
	dictionary, err := wordlists(c, "possible", "solutions")
	if err != nil {
		return err
	}
	strategy, err := NewStrategy(c.String("strategy"))
	if err != nil {
		return err
	}
	secrets, err := secrets(c)
	if err != nil {
		return err
	}
	if c.Bool("speculate") {
		strategy = NewSpeculator(dictionary, strategy)
	}
	game := NewGame(
		WithStrategy(strategy),
		WithDictionary(dictionary),
		WithStart(c.String("start")))
	enc := json.NewEncoder(c.App.Writer)

	var bar *pb.ProgressBar
	if c.Bool("auto") {
		bar = pb.New(len(secrets)).SetWriter(c.App.ErrWriter).Start()
		defer bar.Finish()
	}
	dur := c.Duration("timeout")
	for i := range secrets {
		if bar != nil {
			bar.Increment()
		}
		if err := func() error {
			ctx, cancel := context.WithTimeout(c.Context, dur)
			defer cancel()
			var board *Scoreboard
			board, err = game.Play(ctx, secrets[i])
			if err != nil {
				if !errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				log.Error().Str("secret", secrets[i]).Msg("failed to converge")
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
				Name:    "auto",
				Aliases: []string{"A"},
				Usage:   "auto play from stdin",
				Value:   false,
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Aliases: []string{},
				Usage:   "timeout per iteration in auto play mode",
				Value:   time.Millisecond * 10,
				Hidden:  true,
			},
			wordlistFlag(),
		},
		Action: play,
	}
}
