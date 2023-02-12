package qordle

import (
	"io"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// rounds is the multiplier for the number of rounds to attempt
const rounds int = 3

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
	rounds     int
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

func WithRounds(rounds int) Option {
	return func(g *Game) {
		g.rounds = rounds
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
	start := g.start
	if start == "" {
		start = g.strategy.Apply(g.dictionary)[0]
	}
	if g.rounds <= 0 {
		g.rounds = rounds
	}
	return g.play(secret, start)
}

func (g *Game) play(secret string, start string) (*Scoreboard, error) {
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

	n := len(secret) * g.rounds
	fns := []FilterFunc{Length(len(secret)), IsLower()}
	for len(scoreboard.Rounds) < n {
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
			words = append(words, dictionary[0])
		}
	}
	return scoreboard, nil
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
	if c.Bool("speculate") {
		strategy = NewSpeculator(dictionary, strategy)
	}
	secrets := c.Args().Slice()
	if len(secrets) == 0 {
		secrets, err = read(c.App.Reader)
		if err != nil {
			return err
		}
	}
	game := NewGame(
		WithStrategy(strategy),
		WithDictionary(dictionary),
		WithStart(c.String("start")),
		WithRounds(c.Int("rounds")))
	enc := Runtime(c).Encoder
	writer := io.Discard
	if c.Bool("auto") {
		writer = c.App.ErrWriter
	}
	bar := pb.New(len(secrets)).SetWriter(writer).Start()
	defer bar.Finish()

	var board *Scoreboard
	for i := range secrets {
		bar.Increment()
		board, err = game.Play(secrets[i])
		if err != nil {
			return errors.Wrapf(err, "secret: %s ", secrets[i])
		}
		if err = enc.Encode(board); err != nil {
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
			&cli.IntFlag{
				Name:    "rounds",
				Aliases: []string{"r"},
				Usage:   "max rounds not to exceed `rounds` * len(secret)",
				Value:   rounds,
			},
			wordlistFlag(),
		},
		Action: play,
	}
}
