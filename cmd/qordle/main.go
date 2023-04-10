package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/manual"
	"github.com/bzimmer/qordle"
)

type strategies struct {
	strategies qordle.Trie[qordle.Strategy]
}

func (s strategies) Strategy(prefix string) (qordle.Strategy, error) {
	strategy := s.strategies.Value(prefix)
	if strategy != nil {
		return strategy, nil
	}
	return nil, fmt.Errorf("unknown strategy `%s`", prefix)
}

func (s strategies) Strategies() []string {
	return s.strategies.Strings()
}

func main() {
	app := &cli.App{
		Name:        "qordle",
		HelpName:    "qordle",
		Usage:       "CLI for guessing wordle words",
		Description: "CLI for guessing wordle words",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug log level",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "monochrome",
				Usage: "disable color output",
				Value: false,
			},
		},
		Before: func(c *cli.Context) error {
			level := zerolog.InfoLevel
			if c.Bool("debug") {
				level = zerolog.DebugLevel
			}
			zerolog.SetGlobalLevel(level)
			zerolog.DurationFieldUnit = time.Millisecond
			zerolog.DurationFieldInteger = false
			zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
			log.Logger = log.Output(
				zerolog.ConsoleWriter{
					Out:        c.App.ErrWriter,
					NoColor:    c.Bool("monochrome"),
					TimeFormat: time.RFC3339,
				},
			)

			trie := qordle.Trie[qordle.Strategy]{}
			for _, strategy := range []qordle.Strategy{
				new(qordle.Alpha),
				new(qordle.Bigram),
				new(qordle.Elimination),
				new(qordle.Frequency),
				new(qordle.Position),
			} {
				trie.Add(strategy.String(), strategy)
			}

			c.App.Metadata = map[string]any{
				qordle.RuntimeKey: &qordle.Rt{
					Grab:       &http.Client{Timeout: 2 * time.Second},
					Encoder:    json.NewEncoder(c.App.Writer),
					Start:      time.Now(),
					Strategies: &strategies{strategies: trie},
				},
			}

			return nil
		},
		Commands: []*cli.Command{
			qordle.CommandBee(),
			qordle.CommandLetterBoxed(),
			qordle.CommandOrder(),
			qordle.CommandPlay(),
			qordle.CommandRanks(),
			qordle.CommandScore(),
			qordle.CommandStrategies(),
			qordle.CommandSuggest(),
			qordle.CommandValidate(),
			qordle.CommandVersion(),
			qordle.CommandWordlists(),
			manual.Manual(),
		},
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				log.Error().Err(v).Msg(app.Name)
			case string:
				log.Error().Err(errors.New(v)).Msg(app.Name)
			default:
				log.Error().Err(fmt.Errorf("%v", v)).Msg(app.Name)
			}
			os.Exit(1)
		}
		if err != nil {
			log.Error().Err(err).Msg(app.Name)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	err = app.RunContext(context.Background(), os.Args)
}
