package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/qordle"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "begins",
			Usage: "word begins with",
		},
		&cli.StringFlag{
			Name:  "ends",
			Usage: "word ends with",
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
			Name:  "dictionary",
			Usage: "dictionary of possible words",
			Value: "/usr/share/dict/words",
		},
	}
}

func initLogging(c *cli.Context) {
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
}

func main() {
	app := &cli.App{
		Name:        "qordle",
		HelpName:    "qordle",
		Usage:       "CLI for guessing wordle words",
		Description: "CLI for guessing wordle words",
		Flags:       flags(),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Stack().Err(err).Msg(c.App.Name)
		},
		Before: func(c *cli.Context) error {
			initLogging(c)

			cfg := metrics.DefaultConfig(c.App.Name)
			cfg.EnableRuntimeMetrics = false
			cfg.TimerGranularity = time.Second
			sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
			metric, err := metrics.New(cfg, sink)
			if err != nil {
				return err
			}

			c.App.Metadata = map[string]interface{}{
				qordle.RuntimeKey: &qordle.Runtime{
					Sink:    sink,
					Metrics: metric,
				},
			}

			return nil
		},
		Action: func(c *cli.Context) error {
			fns := []qordle.SolveFunc{qordle.Lower(), qordle.Length(5)}
			if c.IsSet("begins") {
				fns = append(fns, qordle.Begins(c.String("begins")))
			}
			if c.IsSet("ends") {
				fns = append(fns, qordle.Ends(c.String("ends")))
			}
			if c.IsSet("hits") {
				fns = append(fns, qordle.Hits(c.String("hits")))
			}
			if c.IsSet("misses") {
				fns = append(fns, qordle.Misses(c.String("misses")))
			}

			dictionary, err := qordle.DictionaryFs(afero.NewOsFs(), c.String("dictionary"))
			if err != nil {
				return err
			}
			dictionary = qordle.Solve(dictionary, fns...)

			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(dictionary.Words())
		},
	}
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
