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

func (b Board) contains(target int) bool {
	for i := 0; i < len(b); i++ {
		if b[i] == target {
			return true
		}
	}
	return false
}

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
	return strings.Join(val, ", ")
}

type Candidate struct {
	Board Board      `json:"board"`
	Ops   Operations `json:"ops"`
}

func (c Candidate) contains(target int) bool {
	return c.Board.contains(target)
}

type Digits struct{}

func (d Digits) operations(board Board, ops Operations, target int) []Candidate {
	if len(board) == 1 {
		return nil
	}

	operations := []Op{OpAdd, OpMultiply, OpSubtract, OpDivide}

	var candidates []Candidate
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
				if !op.valid(lhs, rhs) {
					continue
				}
				operation := Operation{
					Op:  op,
					LHS: lhs,
					RHS: rhs,
					Val: op.apply(lhs, rhs),
				}
				candidates = append(candidates, Candidate{
					Board: append(next, operation.Val),
					Ops:   append(ops, operation),
				})
				if operation.Val == target {
					return candidates
				}
			}
		}
	}

	return candidates
}

func (d Digits) Play(ctx context.Context, board Board, target int) <-chan Candidate {
	c := make(chan Candidate)
	go func() {
		defer close(c)
		queue := lane.NewPriorityQueue[Candidate](lane.Minimum[int])
		queue.Push(Candidate{Board: board, Ops: nil}, 0)
		for !queue.Empty() {
			val, steps, _ := queue.Pop()
			for _, candidate := range d.operations(val.Board, val.Ops, target) {
				switch candidate.contains(target) {
				case true:
					select {
					case <-ctx.Done():
						return
					case c <- candidate:
					}
				default:
					queue.Push(candidate, steps+1)
				}
			}
		}
	}()
	return c
}

func digits(c *cli.Context) error {
	var board Board
	for i := 0; i < c.NArg(); i++ {
		val, err := strconv.Atoi(c.Args().Get(i))
		if err != nil {
			return fmt.Errorf("failed to convert '%s'", c.Args().Get(i))
		}
		board = append(board, val)
	}

	var digits Digits
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
