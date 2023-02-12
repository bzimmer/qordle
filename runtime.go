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
}

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v any) error
}

func Runtime(c *cli.Context) *Rt {
	return c.App.Metadata[RuntimeKey].(*Rt)
}
