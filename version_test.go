package qordle_test

import (
	"testing"

	"github.com/bzimmer/qordle"
)

func TestVersion(t *testing.T) {
	tests := []harness{
		{
			name: "version",
			args: []string{"version"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandVersion)
		})
	}
}
