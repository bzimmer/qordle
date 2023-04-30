package qordle

import (
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
	return fmt.Sprintf("%04d %s %04d = %04d", o.LHS, o.Op, o.RHS, o.Val)
}

type Operations []Operation

func (o Operations) String() string {
	var val []string
	for _, x := range o {
		val = append(val, x.String())
	}
	return strings.Join(val, ",")
}

type candidate struct {
	Board Board      `json:"board"`
	Ops   Operations `json:"ops"`
}

func operations(board Board, ops Operations) []candidate {
	if len(board) == 1 {
		return nil
	}

	var candidates []candidate
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
			for _, op := range []Op{OpAdd, OpMultiply, OpSubtract, OpDivide} {
				if !op.valid(lhs, rhs) {
					continue
				}
				op := Operation{
					Op:  op,
					LHS: lhs,
					RHS: rhs,
					Val: op.apply(lhs, rhs),
				}
				candidates = append(candidates, candidate{
					Board: append(next, op.Val),
					Ops:   append(ops, op),
				})
			}
		}
	}
	return candidates
}

func digits(c *cli.Context) error {
	target := c.Int("target")
	var board []int

	for i := 0; i < c.NArg(); i++ {
		val, err := strconv.Atoi(c.Args().Get(i))
		if err != nil {
			return err
		}
		board = append(board, val)
	}

	enc := Runtime(c).Encoder
	queue := lane.NewPriorityQueue[candidate](lane.Minimum[int])
	queue.Push(candidate{Board: Board(board), Ops: nil}, 0)
	for !queue.Empty() {
		val, steps, _ := queue.Pop()
	loop:
		for _, candidate := range operations(val.Board, val.Ops) {
			for i := 0; i < len(candidate.Ops); i++ {
				if candidate.Ops[i].Val == target {
					if err := enc.Encode([]any{candidate.Board, candidate.Ops}); err != nil {
						return err
					}
					break loop
				}
			}
			queue.Push(candidate, steps+1)
		}
	}
	return nil
}

func CommandDigits() *cli.Command {
	return &cli.Command{
		Name:     "digits",
		Category: "wordle",
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
