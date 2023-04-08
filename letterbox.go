package qordle

import (
	"fmt"
	"math"
	sys "runtime"
	"strings"
	"sync"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

const (
	letters      = 12
	defaultSides = "rme-wcl-tgk-api"
)

type Graph map[rune][]string

func or(a *bitmap.Bitmap, b *bitmap.Bitmap) *bitmap.Bitmap {
	c := new(bitmap.Bitmap)
	c.Or(*a, *b)
	return c
}

func bm(word string) *bitmap.Bitmap {
	m := new(bitmap.Bitmap)
	for _, r := range word {
		m.Set(uint32(r))
	}
	return m
}

type Box struct {
	min   int // minimum word length to be a solution
	max   int // maximum word chain length to be a solution
	con   int // number of concurrent goroutines
	sides []string
}

type BoxOption func(*Box)

func WithSides(sides string) BoxOption {
	return func(box *Box) {
		box.sides = strings.Split(sides, "-")
	}
}

func WithMinWordLength(n int) BoxOption {
	return func(box *Box) {
		box.min = n
	}
}

func WithMaxSolutionLength(n int) BoxOption {
	return func(box *Box) {
		box.max = n
	}
}

func WithConcurrent(n int) BoxOption {
	return func(box *Box) {
		box.con = n
	}
}

func NewBox(opts ...BoxOption) *Box {
	box := new(Box)
	for i := range opts {
		opts[i](box)
	}
	return box
}

func (box *Box) words(trie *Trie[any], prefix string, side int) []string {
	var s []string
	for i := 0; i < len(box.sides); i++ {
		if i == side {
			// skip the starting side
			continue
		}
		for j := 0; j < len(box.sides[i]); j++ {
			r := prefix + string(box.sides[i][j])
			if node := trie.Node(r); node != nil {
				if node.Word() && len(r) >= box.min {
					s = append(s, r)
				}
				if node.Prefix() {
					s = append(s, box.words(trie, r, i)...)
				}
			}
		}
	}
	return s
}

func (box *Box) Words(trie *Trie[any]) []string {
	var s []string
	for i := 0; i < len(box.sides); i++ {
		for j := 0; j < len(box.sides[i]); j++ {
			r := string(box.sides[i][j])
			s = append(s, box.words(trie, r, i)...)
		}
	}
	return s
}

func (box *Box) graph(words []string) Graph {
	graph := make(map[rune][]string)
	for _, word := range words {
		first := rune(word[0])
		graph[first] = append(graph[first], word)
	}
	return graph
}

type elem struct {
	f rune
	s []string
	b *bitmap.Bitmap
}

func (box *Box) solutions(solutions chan<- []string, graph Graph, e elem) {
	stack := []elem{e}
	for len(stack) > 0 {
		e = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(e.s) > box.max {
			continue
		}
		c := e.b.Count()
		if c == letters {
			solutions <- e.s
			continue
		}
		for _, next := range graph[e.f] {
			union := or(e.b, bm(next))
			if c == union.Count() {
				continue
			}
			last := rune(next[len(next)-1])
			soln := append(slices.Clone(e.s), next)
			stack = append(stack, elem{b: union, f: last, s: soln})
		}
	}
}

func (box *Box) Solutions(words []string) <-chan []string {
	start := time.Now()
	wc := make(chan string)
	go func() {
		defer close(wc)
		for i := range words {
			wc <- words[i]
		}
	}()
	var wg sync.WaitGroup
	g := box.graph(words)
	n := int(math.Max(1, float64(box.con)))
	solutions := make(chan []string, 3*n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range wc {
				e := elem{b: bm(word), s: []string{word}, f: rune(word[len(word)-1])}
				box.solutions(solutions, g, e)
			}
		}()
	}
	go func() {
		defer func(t time.Time) {
			log.Info().Dur("elapsed", time.Since(t)).Int("concurrency", n).Msg("find solutions")
		}(start)
		defer close(solutions)
		wg.Wait()
	}()
	return solutions
}

func box(c *cli.Context) (*Box, error) {
	sides := defaultSides
	switch c.NArg() {
	case 0:
		// use default
	case 1:
		sides = c.Args().First()
	case 4:
		sides = strings.Join(c.Args().Slice(), "-")
	default:
		return nil, fmt.Errorf("found %d sides", c.NArg())
	}
	log.Info().Str("sides", sides).Msg("using")
	return NewBox(
		WithSides(sides),
		WithConcurrent(c.Int("concurrent")),
		WithMinWordLength(c.Int("min")),
		WithMaxSolutionLength(c.Int("max"))), nil
}

func trie(c *cli.Context) (int, *Trie[any], error) {
	defer func(t time.Time) {
		log.Info().Dur("elapsed", time.Since(t)).Msg("build trie")
	}(time.Now())
	dictionary, err := wordlists(c, "qordle")
	if err != nil {
		return 0, nil, err
	}
	trie := &Trie[any]{}
	for i := range dictionary {
		trie.Add(dictionary[i], nil)
	}
	return len(dictionary), trie, nil
}

func CommandLetterBox() *cli.Command {
	return &cli.Command{
		Name:     "letterbox",
		Category: "letterbox",
		Usage:    "Solve the NYT letterbox",
		Flags: append(
			[]cli.Flag{
				&cli.IntFlag{
					Name:  "min",
					Usage: "minimum word size",
					Value: 3,
				},
				&cli.IntFlag{
					Name:  "max",
					Usage: "maximum solution length",
					Value: 2,
				},
				&cli.IntFlag{
					Name:  "concurrent",
					Usage: "number of cpus to use for concurrent solving",
					Value: sys.NumCPU(),
				},
			},
			wordlistFlags()...,
		),
		Action: func(c *cli.Context) error {
			defer func(t time.Time) {
				log.Info().Dur("elapsed", time.Since(t)).Msg(c.Command.Name)
			}(time.Now())
			n, t, err := trie(c)
			if err != nil {
				return err
			}
			box, err := box(c)
			if err != nil {
				return err
			}
			words := box.Words(t)
			log.Info().Int("matching", len(words)).Int("possible", n).Msg("dictionary")
			sol := map[int]int{}
			enc := Runtime(c).Encoder
			for solution := range box.Solutions(words) {
				sol[len(solution)]++
				if err = enc.Encode(solution); err != nil {
					return err
				}
			}
			log.Info().Interface("solutions", sol).Msg("solutions")
			return nil
		},
	}
}
