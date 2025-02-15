package qordle_test

import (
	"encoding/json"
	"io"
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

func TestOperation(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	o := qordle.Operation{Op: qordle.OpAdd, LHS: 1, RHS: 3, Val: 4}
	a.Equal("1 + 3 = 4", o.String())
}

func TestOperations(t *testing.T) {
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
				a.Len(candidates, 9)
			},
		},
		{
			name: "single answer",
			// without hashing results all combinations are produced
			// $ qordle digits -t 2125 1 2 3 4 18 15 | sort | uniq -c | sort -nr
			//    1 {"board":[2125],"ops":[{"op":0,"lhs":4,"rhs":3,"val":7},{"op":1,"lhs":18,"rhs":7,"val":126},
			//		{"op":2,"lhs":126,"rhs":1,"val":125},{"op":0,"lhs":15,"rhs":2,"val":17},{"op":1,"lhs":125,"rhs":17,"val":2125}]}
			//    1 {"board":[2125],"ops":[{"op":0,"lhs":4,"rhs":3,"val":7},{"op":1,"lhs":18,"rhs":7,"val":126},
			//		{"op":0,"lhs":15,"rhs":2,"val":17},{"op":2,"lhs":126,"rhs":1,"val":125},{"op":1,"lhs":125,"rhs":17,"val":2125}]}
			f: func(a *assert.Assertions, digits qordle.Digits) {
				var candidates []qordle.Candidate
				for c := range digits.Play(context.Background(), qordle.Board{1, 2, 3, 4, 18, 15}, 2125) {
					candidates = append(candidates, c)
				}
				a.Len(candidates, 1)
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
	decode := func(c *cli.Context) []qordle.Candidate {
		a := assert.New(t)
		var res []qordle.Candidate
		decoder := json.NewDecoder(c.App.Writer.(io.Reader))
		for decoder.More() {
			var candidate qordle.Candidate
			err := decoder.Decode(&candidate)
			a.NoError(err)
			res = append(res, candidate)
		}
		return res
	}

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
		{
			name: "equation",
			args: []string{"digits", "-e", "-t", "5", "1", "2", "6", "3"},
		},
		{
			name: "count four",
			args: []string{"digits", "-t", "124", "-N", "4", "12", "2", "6", "3", "8"},
			after: func(c *cli.Context) error {
				a := assert.New(t)
				res := decode(c)
				a.Len(res, 4)
				return nil
			},
		},
		{
			name: "count one",
			args: []string{"digits", "-t", "124", "-N", "1", "12", "2", "6", "3", "8"},
			after: func(c *cli.Context) error {
				a := assert.New(t)
				res := decode(c)
				a.Len(res, 1)
				return nil
			},
		},
		{
			name: "count negative one",
			args: []string{"digits", "-t", "124", "-N", "-1", "12", "2", "6", "3", "8"},
			after: func(c *cli.Context) error {
				a := assert.New(t)
				res := decode(c)
				a.Len(res, 10)
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, qordle.CommandDigits)
		})
	}
}
