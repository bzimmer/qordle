package qordle

import (
	"encoding/json"
	"fmt"
	"io"
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
	switch o {
	case OpAdd:
		return "+"
	case OpSubtract:
		return "-"
	case OpMultiply:
		return "*"
	case OpDivide:
		return "/"
	default:
		return "!!!!"
	}
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
	var candidates []candidate
	if len(board) == 1 {
		return candidates
	}
	enc := json.NewEncoder(io.Discard)
	enc.Encode([]any{nil, candidate{Board: board}})

	for i := 0; i < len(board); i++ {
		for j := i + 1; j < len(board); j++ {
			lhs, rhs := board[i], board[j]
			if lhs < rhs {
				lhs, rhs = rhs, lhs
			}
			// var cb Board
			cb := make(Board, i)
			copy(cb, board[0:i])
			cb = append(cb, board[i+1:j]...)
			cb = append(cb, board[j+1:]...)
			// addition
			op := Operation{
				Op:  OpAdd,
				LHS: lhs,
				RHS: rhs,
				Val: lhs + rhs,
			}
			candidates = append(candidates, candidate{
				Board: append(cb, op.Val),
				Ops:   append(ops, op),
			})
			enc.Encode([]any{[]int{i, j}, candidates[len(candidates)-1]})
			// multiply
			op = Operation{
				Op:  OpMultiply,
				LHS: lhs,
				RHS: rhs,
				Val: lhs * rhs,
			}
			candidates = append(candidates, candidate{
				Board: append(cb, op.Val),
				Ops:   append(ops, op),
			})
			enc.Encode([]any{[]int{i, j}, candidates[len(candidates)-1]})
			// subtraction
			op = Operation{
				Op:  OpSubtract,
				LHS: lhs,
				RHS: rhs,
				Val: lhs - rhs,
			}
			candidates = append(candidates, candidate{
				Board: append(cb, op.Val),
				Ops:   append(ops, op),
			})
			enc.Encode([]any{[]int{i, j}, candidates[len(candidates)-1]})
			// division
			if rhs != 0 && lhs%rhs == 0 {
				op = Operation{
					Op:  OpDivide,
					LHS: lhs,
					RHS: rhs,
					Val: lhs / rhs,
				}
				candidates = append(candidates, candidate{
					Board: append(cb, op.Val),
					Ops:   append(ops, op),
				})
				enc.Encode([]any{[]int{i, j}, candidates[len(candidates)-1]})
			}
		}
	}
	return candidates
}

func digits(c *cli.Context) error {
	target := 413
	board := []int{5, 11, 19, 20, 23, 25}

	enc := Runtime(c).Encoder
	queue := lane.NewPriorityQueue[candidate](lane.Minimum[int])
	queue.Push(candidate{Board: Board(board), Ops: nil}, 0)
	for !queue.Empty() {
		val, steps, _ := queue.Pop()
	loop:
		for _, candidate := range operations(val.Board, val.Ops) {
			for i := 0; i < len(candidate.Ops); i++ {
				if candidate.Ops[i].Val == target {
					if err := enc.Encode(candidate.Ops); err != nil {
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
