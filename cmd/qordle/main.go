package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

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
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Stack().Err(err).Msg(c.App.Name)
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

			c.App.Metadata = map[string]any{
				qordle.RuntimeKey: &qordle.Rt{
					Grab:    &http.Client{Timeout: 2 * time.Second},
					Encoder: json.NewEncoder(c.App.Writer),
					Start:   time.Now(),
				},
			}

			return nil
		},
		Commands: []*cli.Command{
			qordle.CommandBee(),
			qordle.CommandLetterBox(),
			qordle.CommandPlay(),
			qordle.CommandScore(),
			qordle.CommandSuggest(),
			qordle.CommandWordlists(),
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
			}
			os.Exit(1)
		}
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	err = app.RunContext(context.Background(), os.Args)
}
