package main

import (
	"context"
	"os"
	"time"

	"github.com/bzimmer/qordle"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "qordled",
		HelpName:    "qordled",
		Usage:       "daemon for guessing wordle words",
		Description: "daemon for guessing wordle words",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Value: 0,
				Usage: "port on which to run",
			},
			&cli.StringFlag{
				Name:    "base-url",
				Value:   "http://localhost",
				Usage:   "Base URL",
				EnvVars: []string{"BASE_URL"},
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug log level",
				Value: false,
			},
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Stack().Err(err).Msg(c.App.Name)
		},
		Action: qordle.ActionAPI(),
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
					NoColor:    false,
					TimeFormat: time.RFC3339,
				},
			)
			return nil
		},
	}
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
