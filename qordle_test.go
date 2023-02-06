package qordle_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

type harness struct {
	name, err string
	args      []string
	before    cli.BeforeFunc
	after     cli.AfterFunc
	context   func(context.Context) context.Context
}

func newTestApp(tt *harness, cmd *cli.Command) *cli.App {
	name := strings.ReplaceAll(tt.name, " ", "-")
	return &cli.App{
		Name:      name,
		HelpName:  name,
		Reader:    new(bytes.Buffer),
		Writer:    new(bytes.Buffer),
		ErrWriter: io.Discard,
		Before: func(c *cli.Context) error {
			c.App.Metadata = map[string]any{
				qordle.RuntimeKey: &qordle.Rt{
					Encoder: json.NewEncoder(c.App.Writer),
					Start:   time.Now(),
				},
			}
			return nil
		},
		Commands: []*cli.Command{cmd},
	}
}

func run(t *testing.T, tt *harness, cmd func() *cli.Command) {
	a := assert.New(t)

	app := newTestApp(tt, cmd())

	if tt.before != nil {
		f := app.Before
		app.Before = func(c *cli.Context) error {
			for _, f := range []cli.BeforeFunc{f, tt.before} {
				if f != nil {
					if err := f(c); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
	if tt.after != nil {
		f := app.After
		app.After = func(c *cli.Context) error {
			for _, f := range []cli.AfterFunc{f, tt.after} {
				if f != nil {
					if err := f(c); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}

	ctx := context.Background()
	if tt.context != nil {
		ctx = tt.context(ctx)
	}
	err := app.RunContext(ctx, append([]string{app.Name}, tt.args...))
	if tt.err == "" {
		a.NoError(err)
		return
	}
	a.Error(err)
	if err != nil { // avoids a panic if err is nil
		a.Contains(err.Error(), tt.err)
	}
}
