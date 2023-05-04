package qordle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	set "github.com/deckarep/golang-set/v2"
	"github.com/oleiade/lane/v2"
	"github.com/urfave/cli/v2"
)

type Operator int

const (
	OpAdd      Operator = 0
	OpMultiply Operator = 1
	OpSubtract Operator = 2
	OpDivide   Operator = 3
)

func (o Operator) String() string {
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

func (o Operator) valid(lhs, rhs int) bool {
	switch o {
	case OpAdd, OpMultiply:
	case OpSubtract:
		return lhs >= rhs
	case OpDivide:
		return rhs > 0 && lhs%rhs == 0
	}
	return true
}

func (o Operator) apply(lhs, rhs int) int {
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
	Op  Operator `json:"op"`
	LHS int      `json:"lhs"`
	RHS int      `json:"rhs"`
	Val int      `json:"val"`
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

func (d Digits) operations(
	board Board, operations Operations, target int) ([]Candidate, []Candidate) {
	if len(board) == 1 {
		return nil, nil
	}

	seen := set.NewThreadUnsafeSet(board...)
	operators := []Operator{OpAdd, OpMultiply, OpSubtract, OpDivide}

	var solutions, candidates []Candidate
	for i := 0; i < len(board); i++ {
		for j := i + 1; j < len(board); j++ {
			lhs, rhs := board[i], board[j]
			if lhs < rhs {
				lhs, rhs = rhs, lhs
			}
			for _, operator := range operators {
				if !operator.valid(lhs, rhs) {
					continue
				}
				val := operator.apply(lhs, rhs)
				if seen.Contains(val) {
					continue
				}
				operation := Operation{
					Op:  operator,
					LHS: lhs,
					RHS: rhs,
					Val: val,
				}
				seen.Add(operation.Val)
				candidate := Candidate{
					Board: make(Board, 0, len(board)-1),
					Ops:   make(Operations, 0, len(operations)+1),
				}
				candidate.Ops = append(candidate.Ops, operations...)
				candidate.Ops = append(candidate.Ops, operation)
				candidate.Board = append(candidate.Board, board[0:i]...)
				candidate.Board = append(candidate.Board, board[i+1:j]...)
				candidate.Board = append(candidate.Board, board[j+1:]...)
				candidate.Board = append(candidate.Board, operation.Val)
				switch {
				case operation.Val == target:
					solutions = append(solutions, candidate)
				default:
					candidates = append(candidates, candidate)
				}
			}
		}
	}

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
