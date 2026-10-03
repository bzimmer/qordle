package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

// strategies lists every available strategy with a human-readable description.
var strategies = []struct { //nolint:gochecknoglobals // read-only table
	strategy    qordle.Strategy
	description string
}{
	{new(qordle.Alpha), "Sort the word list alphabetically"},
	{new(qordle.Bigram), "Rank words by bigram frequency of their letters"},
	{new(qordle.Elimination), "Rank words by how many candidates each guess eliminates"},
	{new(qordle.Frequency), "Rank words by the frequency of their letters in the remaining list"},
	{new(qordle.Position), "Rank words by how often each letter appears in its position"},
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
	for _, s := range strategies {
		registry.Add(s.strategy.String(), s.strategy)
	}
	return &server{dictionary: dictionary, registry: registry}, nil
}

// strategy constructs a strategy from the given names, chaining them when
// more than one is provided. Falls back to frequency+position when no names
// are supplied.
func (s *server) strategy(names []string) (qordle.Strategy, error) {
	if len(names) == 0 {
		names = []string{"frequency", "position"}
	}
	chain := make([]qordle.Strategy, 0, len(names))
	for _, name := range names {
		strategy := s.registry.Value(name)
		if strategy == nil {
			return nil, fmt.Errorf("unknown strategy %q", name)
		}
		chain = append(chain, strategy)
	}
	if len(chain) == 1 {
		return chain[0], nil
	}
	return qordle.NewChain(chain...), nil
}

func (*server) strategies(w http.ResponseWriter, _ *http.Request) {
	result := make(map[string]string, len(strategies))
	for _, s := range strategies {
		result[s.strategy.String()] = s.description
	}
	encode(w, http.StatusOK, result)
}

func (s *server) play(w http.ResponseWriter, r *http.Request) {
	strategy, err := s.strategy(r.URL.Query()["strategy"])
	if err != nil {
		badRequest(w, err)
		return
	}
	game := qordle.NewGame(
		qordle.WithDictionary(s.dictionary),
		qordle.WithStart(r.URL.Query().Get("start")),
		qordle.WithStrategy(qordle.NewSpeculator(s.dictionary, strategy)))
	scoreboard, err := game.Play(r.PathValue("secret"))
	if err != nil {
		badRequest(w, err)
		return
	}
	encode(w, http.StatusOK, scoreboard)
}

func (s *server) suggest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	strategy, err := s.strategy(query["strategy"])
	if err != nil {
		badRequest(w, err)
		return
	}
	if query.Get("speculate") == "true" {
		strategy = qordle.NewSpeculator(s.dictionary, strategy)
	}
	// Fields rather than Split so an empty path yields no guesses instead of
	// a single empty guess matching nothing
	guesser, err := qordle.Guess(strings.Fields(r.PathValue("guesses"))...)
	if err != nil {
		badRequest(w, err)
		return
	}
	// Filter returns a fresh slice so the shared dictionary stays untouched
	encode(w, http.StatusOK, strategy.Apply(qordle.Filter(s.dictionary, guesser)))
}

func encode(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		log.Error().Err(err).Msg("encode")
	}
}

func badRequest(w http.ResponseWriter, err error) {
	log.Error().Err(err).Msg("bad request")
	encode(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
}

// statusWriter records the response status for the request log.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Int("status", sw.status).
			Dur("elapsed", time.Since(start)).
			Msg("request")
	})
}

// newHandler routes the API under /qordle and, when public is set, serves
// the static site from that directory.
func newHandler(public string) (http.Handler, error) {
	srv, err := newServer()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /qordle/strategies", srv.strategies)
	mux.HandleFunc("GET /qordle/play/{secret}", srv.play)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		mux.HandleFunc(method+" /qordle/suggest", srv.suggest)
		mux.HandleFunc(method+" /qordle/suggest/{guesses...}", srv.suggest)
	}
	if public != "" {
		mux.Handle("GET /", http.FileServer(http.Dir(public)))
	}
	return logged(mux), nil
}

func serve(c *cli.Context) error {
	handler, err := newHandler("public")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(c.Context, os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", c.Int("port")),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if serr := server.Shutdown(shutdown); serr != nil { //nolint:contextcheck // parent is already done
			log.Error().Err(serr).Msg("shutdown")
		}
	}()

	log.Info().Str("address", "http://localhost"+server.Addr).Msg("http server")
	if err = server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
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
}
