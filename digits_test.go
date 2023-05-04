package qordle_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
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
	a.Equal("1 + 3 = 4", o.String())
}

func TestOerations(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	o := []qordle.Operation{
		{Op: qordle.OpAdd, LHS: 1, RHS: 3, Val: 4},
		{Op: qordle.OpSubtract, LHS: 10, RHS: 3, Val: 7},
	}
	os := qordle.Operations(o)
	a.Equal("1 + 3 = 4, 10 - 3 = 7", os.String())
}

func TestDigits(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		f    func(*assert.Assertions, qordle.Digits)
	}{
		{
			name: "success",
			f: func(a *assert.Assertions, digits qordle.Digits) {
				var candidates []qordle.Candidate
				for c := range digits.Play(context.Background(), qordle.Board{1, 2, 6, 3}, 5) {
					candidates = append(candidates, c)
				}
				a.Len(candidates, 6)
			},
		},
		{
			name: "large number",
			f: func(a *assert.Assertions, digits qordle.Digits) {
				var candidates []qordle.Candidate
				for c := range digits.Play(context.Background(), qordle.Board{1, 3, 5, 9, 17, 34}, 78132) {
					candidates = append(candidates, c)
				}
				a.Len(candidates, 18)
			},
		},
		{
			name: "exit with deadline",
			f: func(a *assert.Assertions, digits qordle.Digits) {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				defer cancel()

				var candidates []qordle.Candidate
				candidates = []qordle.Candidate{}
				for c := range digits.Play(ctx, qordle.Board{1, 2, 6, 3}, 5) {
					candidates = append(candidates, c)
				}
				a.Len(candidates, 0)
			},
		},
		{
			name: "exit with cancel",
			f: func(a *assert.Assertions, digits qordle.Digits) {
				ctx, cancel := context.WithCancel(context.Background())
				c := digits.Play(ctx, qordle.Board{2, 3, 8}, 5)
				_, ok := <-c
				a.True(ok)
				cancel()
				_, ok = <-c
				a.False(ok)
				a.Error(ctx.Err())
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var digits qordle.Digits
			tt.f(assert.New(t), digits)
		})
	}
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
		{
			name: "encoding error",
			args: []string{"digits", "-t", "5", "1", "2", "6", "3"},
			before: func(c *cli.Context) error {
				qordle.Runtime(c).Encoder = json.NewEncoder(new(errWriter))
				return nil
			},
			err: ErrEncoding.Error(),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandDigits)
		})
	}
}
