package qordle

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func play(c echo.Context) error {
	dictionary, err := DictionaryEm("solutions")
	if err != nil {
		return err
	}
	st, err := NewStrategy(c.QueryParam("strategy"))
	if err != nil {
		return err
	}
	secret := c.Param("secret")
	start := c.QueryParam("start")
	if start == "" {
		start = "brain"
	}
	game := NewGame(
		WithDictionary(dictionary),
		WithStart(start),
		WithStrategy(st))

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
	dictionary, err := DictionaryEm("solutions")
	if err != nil {
		return err
	}
	gss := strings.Split(c.Param("guesses"), ",")
	ff, err := Guesses(gss...)
	if err != nil {
		return err
	}
	dictionary = Filter(dictionary, ff)
	st, err := NewStrategy(c.QueryParam("strategy"))
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
	engine := echo.New()
	engine.Pre(middleware.RemoveTrailingSlash())
	engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339} method=${method} uri=${uri} path=${path} status=${status}\n",
	}))
	engine.HTTPErrorHandler = func(err error, c echo.Context) {
		engine.DefaultHTTPErrorHandler(err, c)
		log.Error().Err(err).Msg("error")
	}

	methods := []string{http.MethodGet, http.MethodPost}
	engine.GET("/play/:secret", play)
	group := engine.Group("/suggest")
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
	log.Info().Str("address", address).Msg("http server")
	return http.ListenAndServe(address, engine)
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

func ActionAPI() cli.ActionFunc {
	return func(c *cli.Context) error {
		if c.IsSet("port") {
			return serve(c)
		}
		return function(c)
	}
}
