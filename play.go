package qordle

import (
	"errors"
	"io"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/urfave/cli/v2"
)

// rounds is the multiplier for the number of rounds to attempt
const rounds int = 3

type Scoreboard struct {
	Target     string   `json:"target"`
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
	dictionary := Filter(g.dictionary, Length(len(secret)), IsLower())
	if len(dictionary) == 0 {
		return nil, errors.New("empty dictionary")
	}
	start := g.start
	if start == "" {
		start = g.strategy.Apply(dictionary)[0]
	}
	return g.play(dictionary, secret, []string{start})
}

func (g *Game) play(dictionary Dictionary, secret string, words []string) (*Scoreboard, error) {
	scoreboard := &Scoreboard{
		Target:     secret,
		Strategy:   g.strategy.String(),
		Dictionary: len(dictionary),
	}
	defer func(t time.Time) {
		scoreboard.Elapsed = time.Since(t).Milliseconds()
	}(time.Now())

	r := g.rounds
	if r <= 0 {
		r = rounds
	}
	n := len(secret) * r
	for len(scoreboard.Rounds) < n {
		scores, err := Score(secret, words...)
		if err != nil {
			return nil, err
		}
		guess, err := Guess(scores[len(scores)-1])
		if err != nil {
			return nil, err
		}
		dictionary = g.strategy.Apply(Filter(dictionary, guess))

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
	dictionary, strategy, err := prepare(c, "possible", "solutions")
	if err != nil {
		return err
	}
	secrets := c.Args().Slice()
	if len(secrets) == 0 {
		secrets, err = read(c.App.Reader)
		if err != nil {
			return err
		}
	}
	enc := Runtime(c).Encoder

	game := NewGame(
		WithStrategy(strategy),
		WithDictionary(dictionary),
		WithStart(c.String("start")),
		WithRounds(c.Int("rounds")))

	writer := io.Discard
	if c.Bool("progress") {
		writer = c.App.ErrWriter
	}
	bar := pb.New(len(secrets)).SetWriter(writer).Start()
	defer bar.Finish()

	var board *Scoreboard
	for i := range secrets {
		bar.Increment()
		board, err = game.Play(secrets[i])
		if err != nil {
			return err
		}
		if err = enc.Encode(board); err != nil {
			return err
		}
	}
	return nil
}

func CommandPlay() *cli.Command {
	return &cli.Command{
		Name:     "play",
		Category: "wordle",
		Usage:    "Play wordle automatically",
		Flags: append(
			[]cli.Flag{
				&cli.StringFlag{
					Name:    "start",
					Aliases: []string{"t"},
					Value:   "",
				},
				&cli.BoolFlag{
					Name:    "progress",
					Aliases: []string{"B"},
					Usage:   "display a progress bar",
					Value:   false,
				},
				&cli.IntFlag{
					Name:    "rounds",
					Aliases: []string{"r"},
					Usage:   "max rounds not to exceed `rounds` * len(secret)",
					Value:   rounds,
				},
			},
			append(wordlistFlags(), strategyFlags()...)...,
		),
		Action: play,
	}
}
