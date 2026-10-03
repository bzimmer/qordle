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

// strategyDescriptions returns a map of each strategy name to a human-readable description.
func strategyDescriptions() map[string]string {
	return map[string]string{
		"alpha":       "Sort the word list alphabetically",
		"bigram":      "Rank words by bigram frequency of their letters",
		"elimination": "Rank words by how many candidates each guess eliminates",
		"frequency":   "Rank words by the frequency of their letters in the remaining list",
		"position":    "Rank words by how often each letter appears in its position",
	}
}

// server holds state shared across requests; the dictionary is read-only
// and every strategy is stateless, so one instance serves concurrent requests.
type server struct {
	dictionary qordle.Dictionary
	registry   *qordle.Trie[qordle.Strategy]
}

func newServer() (*server, error) {
	dictionary, err := qordle.Read("solutions")
	if err != nil {
		return nil, err
	}
	registry := &qordle.Trie[qordle.Strategy]{}
	for _, s := range []qordle.Strategy{
		new(qordle.Alpha),
		new(qordle.Bigram),
		new(qordle.Elimination),
		new(qordle.Frequency),
		new(qordle.Position),
	} {
		registry.Add(s.String(), s)
	}
	return &server{dictionary: dictionary, registry: registry}, nil
}

// buildStrategy constructs a strategy from the given names, chaining them
// when more than one is provided. Falls back to frequency+position when
// no names are supplied.
func (s *server) buildStrategy(names []string) (qordle.Strategy, error) {
	if len(names) == 0 {
		names = []string{"frequency", "position"}
	}
	strategies := make([]qordle.Strategy, 0, len(names))
	for _, name := range names {
		strategy := s.registry.Value(name)
		if strategy == nil {
			return nil, fmt.Errorf("unknown strategy %q", name)
		}
		strategies = append(strategies, strategy)
	}
	if len(strategies) == 1 {
		return strategies[0], nil
	}
	return qordle.NewChain(strategies...), nil
}

func (s *server) strategies(c echo.Context) error {
	names := s.registry.Strings()
	sort.Strings(names)
	descs := strategyDescriptions()
	result := make(map[string]string, len(names))
	for _, name := range names {
		result[name] = descs[name]
	}
	return c.JSONPretty(http.StatusOK, result, " ")
}

func (s *server) play(c echo.Context) error {
	strategy, err := s.buildStrategy(c.QueryParams()["strategy"])
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	strategy = qordle.NewSpeculator(s.dictionary, strategy)
	game := qordle.NewGame(
		qordle.WithDictionary(s.dictionary),
		qordle.WithStart(c.QueryParam("start")),
		qordle.WithStrategy(strategy))
	scoreboard, err := game.Play(c.Param("secret"))
	if err != nil {
		return err
	}
	return c.JSONPretty(http.StatusOK, scoreboard, " ")
}

func (s *server) suggest(c echo.Context) error {
	strategy, err := s.buildStrategy(c.QueryParams()["strategy"])
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if c.QueryParam("speculate") == "true" {
		strategy = qordle.NewSpeculator(s.dictionary, strategy)
	}
	// Fields rather than Split so an empty path yields no guesses instead of
	// a single empty guess matching nothing
	guesser, err := qordle.Guess(strings.Fields(c.Param("guesses"))...)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// Filter returns a fresh slice so the shared dictionary stays untouched
	words := strategy.Apply(qordle.Filter(s.dictionary, guesser))
	return c.JSONPretty(http.StatusOK, words, " ")
}

func newEngine() (*echo.Echo, error) {
	srv, err := newServer()
	if err != nil {
		return nil, err
	}
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
	base.GET("/strategies", srv.strategies)
	base.GET("/play/:secret", srv.play)
	group := base.Group("/suggest")
	group.Match(methods, "", srv.suggest)
	group.Match(methods, "/:guesses", srv.suggest)
	return engine, nil
}

func serve(c *cli.Context) error {
	engine, err := newEngine()
	if err != nil {
		return err
	}
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
