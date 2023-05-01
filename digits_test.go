package qordle_test

import (
	"testing"

	"github.com/bzimmer/qordle"
)

func TestDigitsCommand(t *testing.T) {
	for _, tt := range []harness{
		{
			name: "digits",
			args: []string{"digits", "-t", "413", "5", "11", "19", "20", "23", "25"},
		},
		{
			name: "invalid number",
			args: []string{"digits", "-t", "413", "20", "23", "aa"},
			err:  `failed to convert 'aa'`,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandDigits)
		})
	}
}
