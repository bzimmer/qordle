package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func flags() []cli.Flag {
	return []cli.Flag{
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
		Commands: []*cli.Command{
			qordle.CommandPlay(),
			qordle.CommandScore(),
			qordle.CommandSuggest(),
			qordle.CommandWordlists(),
		},
	}
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
