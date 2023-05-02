package qordle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/bzimmer/qordle"
)

func TestOp(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	a.Equal("+", qordle.OpAdd.String())
	a.Equal("*", qordle.OpMultiply.String())
	a.Equal("/", qordle.OpDivide.String())
	a.Equal("-", qordle.OpSubtract.String())
}

func TestOeration(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	o := qordle.Operation{Op: qordle.OpAdd, LHS: 1, RHS: 3, Val: 4}
	a.Equal("0001 + 0003 = 0004", o.String())
}

func TestOerations(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	o := []qordle.Operation{
		{Op: qordle.OpAdd, LHS: 1, RHS: 3, Val: 4},
		{Op: qordle.OpSubtract, LHS: 10, RHS: 3, Val: 7},
	}
	os := qordle.Operations(o)
	a.Equal("0001 + 0003 = 0004, 0010 - 0003 = 0007", os.String())
}

func TestDigits(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	var digits qordle.Digits
	var candidates []qordle.Candidate
	for c := range digits.Play(context.Background(), qordle.Board{1, 2, 6, 3}, 5) {
		candidates = append(candidates, c)
	}
	a.Len(candidates, 13)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	cancel()

	candidates = []qordle.Candidate{}
	for c := range digits.Play(ctx, qordle.Board{1, 2, 6, 3}, 5) {
		candidates = append(candidates, c)
	}
	a.LessOrEqual(len(candidates), 0)
}

func TestDigitsCommand(t *testing.T) {
	for _, tt := range []harness{
		{
			name: "digits",
			args: []string{"digits", "-t", "5", "1", "2", "6", "3"},
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
