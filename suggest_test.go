package qordle_test

import (
	"testing"

	"github.com/bzimmer/qordle"
)

func TestSuggestCommand(t *testing.T) {
	for _, tt := range []harness{
		{
			name: "alpha",
			args: []string{"suggest", "--strategy", "a", "raise", "fol.l.y"},
		},
		{
			name: "frequency",
			args: []string{"suggest", "--strategy", "f", "raise", "fol.l.y"},
		},
		{
			name: "position",
			args: []string{"suggest", "--strategy", "p", "raise", "fol.l.y"},
		},
		{
			name: "unknown",
			args: []string{"suggest", "--strategy", "u", "raise", "fol.l.y"},
			err:  "unknown strategy `u`",
		},
		{
			name: "speculate",
			args: []string{"suggest", "-S", "raise", "fol.l.y"},
		},
		{
			name: "bad pattern",
			args: []string{"suggest", "--pattern", "[A-Z", "raise", "fol.l.y"},
			err:  "error parsing regexp: missing closing ]: `[A-Z`",
		},
		{
			name: "bad wordlist",
			args: []string{"suggest", "-w", "foobar", "raise", "fol.l.y"},
			err:  "invalid wordlist `foobar`",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandSuggest)
		})
	}
}
