package qordle

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

const letters = 12

type Graph map[rune][]string

func bitmap(word string) *roaring.Bitmap {
	bm := roaring.New()
	for _, r := range word {
		bm.Add(uint32(r))
	}
	return bm
}

type Trie struct {
	word     bool
	children map[rune]*Trie
}

func NewTrie() *Trie {
	return new(Trie)
}

func (trie *Trie) Add(word string) {
	node := trie
	for _, r := range strings.ToLower(word) {
		child := node.children[r]
		if child == nil {
			if node.children == nil {
				node.children = make(map[rune]*Trie)
			}
			child = new(Trie)
			node.children[r] = child
		}
		node = child
	}
	node.word = true
}

func (trie *Trie) Node(word string) *Trie {
	node := trie
	for _, r := range word {
		child := node.children[r]
		if child == nil {
			return nil
		}
		node = child
	}
	return node
}

func (trie *Trie) Prefix() bool {
	return trie != nil && len(trie.children) > 0
}

func (trie *Trie) Word() bool {
	return trie != nil && trie.word
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

func (box *Box) words(trie *Trie, prefix string, side int) []string {
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

func (box *Box) Words(trie *Trie) []string {
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
	b *roaring.Bitmap
}

func (box *Box) solutions(solutions chan<- []string, graph Graph, e elem) {
	stack := []elem{e}
	for len(stack) > 0 {
		e = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(e.s) > box.max {
			continue
		}
		if e.b.GetCardinality() == letters {
			solutions <- e.s
			continue
		}
		for _, next := range graph[e.f] {
			union := roaring.Or(e.b, bitmap(next))
			if union.GetCardinality() == e.b.GetCardinality() {
				continue
			}
			last := rune(next[len(next)-1])
			soln := append(slices.Clone(e.s), next)
			stack = append(stack, elem{b: union, f: last, s: soln})
		}
	}
}

func (box *Box) Solutions(words []string) <-chan []string {
	wc := make(chan string)
	go func() {
		defer close(wc)
		for _, word := range words {
			wc <- word
		}
	}()
	var wg sync.WaitGroup
	graph := box.graph(words)
	solutions := make(chan []string, box.con)
	for i := 0; i < box.con; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range wc {
				e := elem{b: bitmap(word), s: []string{word}, f: rune(word[len(word)-1])}
				box.solutions(solutions, graph, e)
			}
		}()
	}
	go func() {
		defer close(solutions)
		wg.Wait()
	}()
	return solutions
}

func trie(c *cli.Context) (int, *Trie, error) {
	var err error
	var dict Dictionary
	switch args := c.Args().Slice(); len(args) {
	case 0:
		if dict, err = DictionaryEm("solutions"); err != nil {
			return 0, nil, err
		}
	default:
		var m Dictionary
		fs := afero.NewOsFs()
		for i := 0; i < len(args); i++ {
			if m, err = DictionaryFs(fs, args[i]); err != nil {
				return 0, nil, err
			}
			dict = append(dict, m...)
		}
	}
	trie := NewTrie()
	for i := range dict {
		trie.Add(dict[i])
	}
	return len(dict), trie, nil
}

func CommandLetterBox() *cli.Command {
	return &cli.Command{
		Name:  "letterbox",
		Usage: "play letterbox",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "box",
				Usage: "the letter box in `aaa-bbb-ccc-ddd` format",
				Value: "rme-wcl-tgk-api",
			},
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
				Value: runtime.NumCPU(),
			},
		},
		Action: func(c *cli.Context) error {
			defer func(t time.Time) {
				log.Info().Dur("elapsed", time.Since(t)).Msg(c.Command.Name)
			}(time.Now())
			n, t, err := trie(c)
			if err != nil {
				return err
			}
			box := NewBox(
				WithSides(c.String("box")),
				WithConcurrent(c.Int("concurrent")),
				WithMinWordLength(c.Int("min")),
				WithMaxSolutionLength(c.Int("max")))
			words := box.Words(t)
			log.Info().Int("matching", len(words)).Int("possible", n).Msg("dictonary")

			start := time.Now()
			solutions := map[int]int{}
			enc := json.NewEncoder(c.App.Writer)
			for solution := range box.Solutions(words) {
				solutions[len(solution)]++
				if err = enc.Encode(solution); err != nil {
					return err
				}
			}
			log.Info().Interface("solutions", solutions).Dur("elapsed", time.Since(start)).Msg("find solutions")

			return nil
		},
	}
}
