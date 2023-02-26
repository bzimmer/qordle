package qordle

import (
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
)

// RuntimeKey in app metadata
const RuntimeKey = "github.com/bzimmer/qordle#RuntimeKey"

type Grab interface {
	Do(*http.Request) (*http.Response, error)
}

// Rt for access to runtime components
type Rt struct {
	// Encoder encodes a struct
	Encoder Encoder
	// Grab for querying http endpoints
	Grab Grab
	// Start time of the execution
	Start time.Time
	// Strategy returns the strategy for the code
	Strategy func(string) (Strategy, error)
}

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v any) error
}

func Runtime(c *cli.Context) *Rt {
	return c.App.Metadata[RuntimeKey].(*Rt)
}

func prepare(c *cli.Context, wordlist ...string) (Dictionary, Strategy, error) {
	dictionary, err := wordlists(c, wordlist...)
	if err != nil {
		return nil, nil, err
	}
	var strategy Strategy
	strategies := c.StringSlice("strategy")
	switch n := len(strategies); n {
	case 1:
		strategy, err = Runtime(c).Strategy(strategies[0])
		if err != nil {
			return nil, nil, err
		}
	default:
		s := make([]Strategy, n)
		for i := range strategies {
			strategy, err = Runtime(c).Strategy(strategies[i])
			if err != nil {
				return nil, nil, err
			}
			s[i] = strategy
		}
		strategy = NewChain(s...)
	}
	if c.Bool("speculate") {
		strategy = NewSpeculator(dictionary, strategy)
	}
	return dictionary, strategy, nil
}
