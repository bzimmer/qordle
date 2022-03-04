package qordle

import (
	"github.com/armon/go-metrics"
)

// RuntimeKey in app metadata
const RuntimeKey = "github.com/bzimmer/qordle#RuntimeKey"

// Runtime for access to runtime components
type Runtime struct {
	// Sink for metrics
	Sink *metrics.InmemSink
	// Metrics for capturing metrics
	Metrics *metrics.Metrics
}
