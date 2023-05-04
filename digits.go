package qordle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/oleiade/lane/v2"
	"github.com/urfave/cli/v2"
)

// Op represents an operation
type Op int

const (
	OpAdd      Op = 0
	OpMultiply Op = 1
	OpSubtract Op = 2
	OpDivide   Op = 3
)

func (o Op) String() string {
	var sign string
	switch o {
	case OpAdd:
		sign = "+"
	case OpSubtract:
		sign = "-"
	case OpMultiply:
		sign = "*"
	case OpDivide:
		sign = "/"
	}
	return sign
}

func (o Op) valid(lhs, rhs int) bool {
	switch o {
	case OpAdd, OpMultiply:
	case OpSubtract:
		return lhs >= rhs
	case OpDivide:
		return rhs > 0 && lhs%rhs == 0
	}
	return true
}

func (o Op) apply(lhs, rhs int) int {
	var val int
	switch o {
	case OpAdd:
		val = lhs + rhs
	case OpSubtract:
		val = lhs - rhs
	case OpMultiply:
		val = lhs * rhs
	case OpDivide:
		val = lhs / rhs
	}
	return val
}

type Board []int

type Operation struct {
	Op  Op  `json:"op"`
	LHS int `json:"lhs"`
	RHS int `json:"rhs"`
	Val int `json:"val"`
}

func (o Operation) String() string {
	return fmt.Sprintf("%d %s %d = %d", o.LHS, o.Op, o.RHS, o.Val)
}

type Operations []Operation

func (o Operations) String() string {
	var val []string
	for _, x := range o {
		val = append(val, x.String())
	}
	return strings.Join(val, ", ")
}

type Candidate struct {
	Board Board      `json:"board"`
	Ops   Operations `json:"ops"`
}

type Digits struct{}

func (d Digits) operations(board Board, ops Operations, target int) ([]Candidate, []Candidate) {
	if len(board) == 1 {
		return nil, nil
	}

	operations := []Op{OpAdd, OpMultiply, OpSubtract, OpDivide}

	// rows := make([][]string, len(board)-1)
	// for i := 0; i < len(rows); i++ {
	// 	rows[i] = make([]string, len(operations))
	// }

	var solutions, candidates []Candidate
	for i := 0; i < len(board); i++ {
		for j := i + 1; j < len(board); j++ {
			lhs, rhs := board[i], board[j]
			if lhs < rhs {
				lhs, rhs = rhs, lhs
			}
			next := make(Board, 0)
			next = append(next, board[0:i]...)
			next = append(next, board[i+1:j]...)
			next = append(next, board[j+1:]...)
			for _, op := range operations {
				// rows[i][k] = "-"
				if !op.valid(lhs, rhs) {
					continue
				}
				operation := Operation{
					Op:  op,
					LHS: lhs,
					RHS: rhs,
					Val: op.apply(lhs, rhs),
				}
				candidate := Candidate{
					Board: append(next, operation.Val),
					Ops:   append(ops, operation),
				}
				switch {
				case operation.Val == target:
					solutions = append(solutions, candidate)
				default:
					candidates = append(candidates, candidate)
				}
				// rows[i][k] = fmt.Sprintf("%s => %v", operation.String(), append(next, operation.Val))
			}
		}
	}
	// table := tablewriter.NewWriter(os.Stderr)
	// table.AppendBulk(rows)
	// table.SetColWidth(75)
	// table.Render()
	// fmt.Fprintln(os.Stderr)

	return solutions, candidates
}

func (d Digits) Play(ctx context.Context, board Board, target int) <-chan Candidate {
	c := make(chan Candidate)
	go func() {
		defer close(c)
		queue := lane.NewPriorityQueue[Candidate](lane.Minimum[int])
		queue.Push(Candidate{Board: board, Ops: nil}, 0)
		for !queue.Empty() {
			val, steps, _ := queue.Pop()
			if ctx.Err() != nil {
				return
			}
			// fmt.Fprintf(os.Stderr, "step: %d, board: %v, ops: %v\n", steps, val.Board, val.Ops)
			solutions, candidates := d.operations(val.Board, val.Ops, target)
			for _, solution := range solutions {
				select {
				case <-ctx.Done():
					return
				case c <- solution:
				}
			}
			for _, candidate := range candidates {
				queue.Push(candidate, steps+1)
			}
		}
	}()
	return c
}

func digits(c *cli.Context) error {
	var board Board
	var digits Digits
	for i := 0; i < c.NArg(); i++ {
		val, err := strconv.Atoi(c.Args().Get(i))
		if err != nil {
			return fmt.Errorf("failed to convert '%s'", c.Args().Get(i))
		}
		board = append(board, val)
	}

	enc := Runtime(c).Encoder
	for candidate := range digits.Play(c.Context, board, c.Int("target")) {
		if err := enc.Encode(candidate); err != nil {
			return err
		}
	}

	return nil
}

func CommandDigits() *cli.Command {
	return &cli.Command{
		Name:     "digits",
		Category: "puzzles",
		Usage:    "Play digits automatically",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "target",
				Aliases: []string{"t"},
				Value:   0,
			},
		},
		Action: digits,
	}
}

/*
-t 78132 1 3 5 9 17 34
{
	"board": [
	  78132
	],
	"ops": [
	  {
		"op": 1,
		"lhs": 34,
		"rhs": 3,
		"val": 102
	  },
	  {
		"op": 1,
		"lhs": 9,
		"rhs": 5,
		"val": 45
	  },
	  {
		"op": 1,
		"lhs": 45,
		"rhs": 17,
		"val": 765
	  },
	  {
		"op": 2,
		"lhs": 765,
		"rhs": 102,
		"val": 663
	  },
	  {
		"op": 1,
		"lhs": 766,
		"rhs": 102,
		"val": 78132
	  }
	]
  }
*/

/*
-t 781 1 3 5 9 17 34 | jq -s "first"
{
  "board": [
    5,
    781
  ],
  "ops": [
    {
      "op": 0,
      "lhs": 9,
      "rhs": 3,
      "val": 12
    },
    {
      "op": 0,
      "lhs": 34,
      "rhs": 12,
      "val": 46
    },
    {
      "op": 1,
      "lhs": 46,
      "rhs": 17,
      "val": 782
    },
    {
      "op": 2,
      "lhs": 782,
      "rhs": 5,
      "val": 777
    }
  ]
}*/
