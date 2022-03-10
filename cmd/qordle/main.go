package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/bzimmer/qordle"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

func flags() []cli.Flag {
	return []cli.Flag{
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
			Usage: "match against a known pattern (all letters green, use '.' for all unknown letters)",
		},
		&cli.StringFlag{
			Name:  "dictionary",
			Usage: "dictionary of possible words",
			Value: "/usr/share/dict/words",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug log level",
			Value: false,
		},
	}
}

func initLogging(c *cli.Context) error {
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
			NoColor:    false,
			TimeFormat: time.RFC3339,
		},
	)
	return nil
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
		Before: initLogging,
		Action: func(c *cli.Context) error {
			fns := []qordle.FilterFunc{qordle.Lower(), qordle.Length(c.Int("length"))}
			switch c.NArg() {
			case 0:
				if c.IsSet("pattern") {
					f, err := qordle.Pattern(c.String("pattern"))
					if err != nil {
						return err
					}
					fns = append(fns, f)
				}
				if c.IsSet("hits") {
					fns = append(fns, qordle.Hits(c.String("hits")))
				}
				if c.IsSet("misses") {
					fns = append(fns, qordle.Misses(c.String("misses")))
				}
			default:
				fs, err := qordle.Guesses(c.Args().Slice()...)
				if err != nil {
					return err
				}
				fns = append(fns, fs)
			}
			var err error
			var source string
			var dictionary qordle.Dictionary
			switch c.IsSet("dictionary") {
			case true:
				source = c.String("dictionary")
				dictionary, err = qordle.DictionaryFs(afero.NewOsFs(), source)
			case false:
				source = "embedded"
				dictionary, err = qordle.DictionarySlice(qordle.Words)
			}
			if err != nil {
				return err
			}
			log.Debug().Str("source", source).Msg("dictionary")
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
