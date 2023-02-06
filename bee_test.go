package qordle_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/bzimmer/qordle"
)

type grab struct {
	err    error
	status int
}

func (g grab) Do(req *http.Request) (*http.Response, error) {
	if g.err != nil {
		return nil, g.err
	}
	w := httptest.NewRecorder()
	w.WriteHeader(g.status)
	switch g.status {
	case http.StatusOK:
		file := "testdata/spelling-bee.html"
		http.ServeFile(w, req, file)
	default:
	}
	return w.Result(), nil
}

func TestBeeCommand(t *testing.T) {
	for _, tt := range []harness{
		{
			name: "today",
			args: []string{"bee", "today"},
			before: func(c *cli.Context) error {
				qordle.Runtime(c).Grab = grab{
					status: http.StatusOK,
				}
				return nil
			},
		},
		{
			name: "not found",
			args: []string{"bee", "today"},
			err:  "status: 404",
			before: func(c *cli.Context) error {
				qordle.Runtime(c).Grab = grab{
					status: http.StatusNotFound,
				}
				return nil
			},
		},
		{
			name: "grab error",
			args: []string{"bee", "today"},
			err:  "something failed",
			before: func(c *cli.Context) error {
				qordle.Runtime(c).Grab = grab{
					err: errors.New("something failed"),
				}
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandBee)
		})
	}
}
