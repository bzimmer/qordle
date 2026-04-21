package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

// registry holds all available strategies keyed by name.
var registry = func() qordle.Trie[qordle.Strategy] {
	t := qordle.Trie[qordle.Strategy]{}
	for _, s := range []qordle.Strategy{
		new(qordle.Alpha),
		new(qordle.Bigram),
		new(qordle.Elimination),
		new(qordle.Frequency),
		new(qordle.Position),
	} {
		t.Add(s.String(), s)
	}
	return t
}()

// strategyNames returns the sorted list of available strategy names.
func strategyNames() []string {
	names := registry.Strings()
	sort.Strings(names)
	return names
}

// buildStrategy constructs a strategy from the given names, chaining them
// when more than one is provided. Falls back to frequency+position when
// no names are supplied.
func buildStrategy(names []string) (qordle.Strategy, error) {
	if len(names) == 0 {
		names = []string{"frequency", "position"}
	}
	strategies := make([]qordle.Strategy, 0, len(names))
	for _, name := range names {
		s := registry.Value(name)
		if s == nil {
			return nil, fmt.Errorf("unknown strategy %q", name)
		}
		strategies = append(strategies, s)
	}
	if len(strategies) == 1 {
		return strategies[0], nil
	}
	return qordle.NewChain(strategies...), nil
}

func strategies(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, strategyNames(), " ")
}

func play(c echo.Context) error {
	dictionary, err := qordle.Read("solutions")
	if err != nil {
		return err
	}
	strategy, err := buildStrategy(c.QueryParams()["strategy"])
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	strategy = qordle.NewSpeculator(dictionary, strategy)
	secret := c.Param("secret")
	game := qordle.NewGame(
		qordle.WithDictionary(dictionary),
		qordle.WithStart(c.QueryParam("start")),
		qordle.WithStrategy(strategy))
	scoreboard, err := game.Play(secret)
	if err != nil {
		return err
	}
	return c.JSONPretty(http.StatusOK, scoreboard, " ")
}

func suggest(c echo.Context) error {
	dictionary, err := qordle.Read("solutions")
	if err != nil {
		return err
	}
	strategy, err := buildStrategy(c.QueryParams()["strategy"])
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	strategy = qordle.NewSpeculator(dictionary, strategy)
	guesser, err := qordle.Guess(strings.Split(c.Param("guesses"), " ")...)
	if err != nil {
		return err
	}
	dictionary = strategy.Apply(qordle.Filter(dictionary, guesser))
	return c.JSONPretty(http.StatusOK, dictionary, " ")
}

func newEngine() *echo.Echo {
	engine := echo.New()
	engine.Pre(middleware.Rewrite(map[string]string{"/qordle/*": "/$1"}))
	engine.Pre(middleware.RemoveTrailingSlash())
	engine.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("time=%s method=%s uri=%s path=%s status=%d\n", //nolint:forbidigo // log
				time.Now().Format(time.RFC3339),
				v.Method,
				v.URI,
				c.Path(),
				v.Status,
			)
			return nil
		},
	}))
	engine.HTTPErrorHandler = func(err error, c echo.Context) {
		engine.DefaultHTTPErrorHandler(err, c)
		log.Error().Err(err).Msg("error")
	}

	base := engine.Group("")
	methods := []string{http.MethodGet, http.MethodPost}
	base.GET("/strategies", strategies)
	base.GET("/play/:secret", play)
	group := base.Group("/suggest")
	group.Match(methods, "", suggest)
	group.Match(methods, "/:guesses", suggest)
	return engine
}

func serve(c *cli.Context) error {
	engine := newEngine()
	engine.Static("/", "public")
	address := fmt.Sprintf(":%d", c.Int("port"))
	log.Info().Str("address", "http://localhost"+address).Msg("http server")
	return engine.Start(address)
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
		Action: serve,
		Before: func(c *cli.Context) error {
			level := zerolog.InfoLevel
			if c.Bool("debug") {
				level = zerolog.DebugLevel
			}
			zerolog.SetGlobalLevel(level)
			zerolog.DurationFieldUnit = time.Millisecond
			zerolog.DurationFieldInteger = false
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
