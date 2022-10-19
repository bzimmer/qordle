package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func play(c echo.Context) error {
	dictionary, err := qordle.DictionaryEm("solutions")
	if err != nil {
		return err
	}
	st, err := qordle.NewStrategy(c.QueryParam("strategy"))
	if err != nil {
		return err
	}
	secret := c.Param("secret")
	start := c.QueryParam("start")
	if start == "" {
		start = "brain"
	}
	game := qordle.NewGame(
		qordle.WithDictionary(dictionary),
		qordle.WithStart(start),
		qordle.WithStrategy(st))

	dictionary, err = game.Play(secret)
	if err != nil {
		return err
	}
	log.Debug().
		Str("secret", secret).
		Str("start", start).
		Str("strategy", st.String()).
		Strs("dictionary", dictionary).Msg("play")
	return c.JSONPretty(http.StatusOK, dictionary, " ")
}

func suggest(c echo.Context) error {
	dictionary, err := qordle.DictionaryEm("solutions")
	if err != nil {
		return err
	}
	gss := strings.Split(c.Param("guesses"), ",")
	ff, err := qordle.Guesses(gss...)
	if err != nil {
		return err
	}
	dictionary = qordle.Filter(dictionary, ff)
	st, err := qordle.NewStrategy(c.QueryParam("strategy"))
	if err != nil {
		return err
	}
	dictionary = st.Apply(dictionary)
	log.Debug().
		Str("strategy", st.String()).
		Strs("dictionary", dictionary).Msg("play")
	return c.JSONPretty(http.StatusOK, dictionary, " ")
}

func newEngine(c *cli.Context) (*echo.Echo, error) {
	baseURL := c.String("base-url")
	log.Info().Str("baseURL", baseURL).Msg("found baseURL")
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	log.Info().Str("path", u.Path).Msg("root path")

	engine := echo.New()
	engine.Pre(middleware.RemoveTrailingSlash())
	engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339} method=${method} uri=${uri} path=${path} status=${status}\n",
	}))
	engine.HTTPErrorHandler = func(err error, c echo.Context) {
		engine.DefaultHTTPErrorHandler(err, c)
		log.Error().Err(err).Msg("error")
	}

	base := engine.Group(u.Path)
	methods := []string{http.MethodGet, http.MethodPost}
	base.GET("/play/:secret", play)
	group := base.Group("/suggest")
	group.Match(methods, "", suggest)
	group.Match(methods, "/:guesses", suggest)
	return engine, nil
}

func serve(c *cli.Context) error {
	engine, err := newEngine(c)
	if err != nil {
		return err
	}
	engine.Static("/", "public")
	address := fmt.Sprintf(":%d", c.Int("port"))
	log.Info().Str("address", "http://localhost"+address).Msg("http server")
	return engine.Start(address)
}

func function(c *cli.Context) error {
	engine, err := newEngine(c)
	if err != nil {
		return err
	}
	log.Info().Msg("lambda function")
	gl := echoadapter.New(engine)
	lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return gl.ProxyWithContext(ctx, req)
	})
	return nil
}

func action(c *cli.Context) error {
	if c.IsSet("port") {
		return serve(c)
	}
	return function(c)
}

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
		Action: action,
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
