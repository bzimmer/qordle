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

func (o Operations) simplify() Operations {
	// [432,24]
	// 	{"op":0,"lhs": 25,"rhs": 23,"val": 48}
	// 	{"op":0,"lhs": 20,"rhs": 11,"val": 31}
	// 	{"op":1,"lhs": 48,"rhs":  9,"val":432}
	// 	{"op":2,"lhs":432,"rhs": 19,"val":413}

	// [413]
	//  {"op":1,"lhs": 20,"rhs": 19,"val":380}
	//  {"op":0,"lhs": 25,"rhs": 23,"val": 48}
	//  {"op":2,"lhs":  5,"rhs":  2,"val":  3}
	//  {"op":2,"lhs":380,"rhs":  3,"val":377}
	//  {"op":0,"lhs":380,"rhs": 33,"val":413}

	// [413]
	//  {"op":1,"lhs": 25,"rhs": 11,"val":275}
	//  {"op":0,"lhs":275,"rhs": 23,"val":298}
	//  {"op":1,"lhs": 19,"rhs":  5,"val": 95}
	//  {"op":2,"lhs":298,"rhs": 95,"val":203}
	//  {"op":0,"lhs":298,"rhs":115,"val":413}
	return o
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
	operations := []Op{OpAdd, OpMultiply, OpSubtract, OpDivide}
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
				candidates = append(candidates, candidate{
					Board: append(next, operation.Val),
					Ops:   append(ops, operation),
				})
			}
		}
	}

	return candidates
}

func digits(c *cli.Context) error {
	var board Board
	for i := 0; i < c.NArg(); i++ {
		val, err := strconv.Atoi(c.Args().Get(i))
		if err != nil {
			return err
		}
		board = append(board, val)
	}

	enc := Runtime(c).Encoder
	target := c.Int("target")
	queue := lane.NewPriorityQueue[candidate](lane.Minimum[int])
	queue.Push(candidate{Board: board, Ops: nil}, 0)
	for !queue.Empty() {
		val, steps, _ := queue.Pop()
	loop:
		for _, candidate := range operations(val.Board, val.Ops) {
			for i := 0; i < len(candidate.Ops); i++ {
				if candidate.Ops[i].Val == target {
					if err := enc.Encode([]any{candidate.Board, candidate.Ops.simplify()}); err != nil {
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
