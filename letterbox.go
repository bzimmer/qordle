package qordle

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/rs/zerolog/log"
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

func (trie *Trie) Load(in io.Reader) (int, error) {
	defer func(t time.Time) {
		log.Info().Dur("elapsed", time.Since(t)).Msg("build trie")
	}(time.Now())
	var count int
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		count++
		text := strings.ToLower(scanner.Text())
		trie.Add(text)
	}
	return count, scanner.Err()
}

func (trie *Trie) Add(word string) {
	node := trie
	for _, r := range word {
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
		if !node.Prefix() {
			return nil
		}
		child := node.children[r]
		if child == nil {
			return nil
		}
		node = child
	}
	return node
}

func (trie *Trie) Prefix() bool {
	return len(trie.children) > 0
}

func (trie *Trie) Word() bool {
	return trie.word
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

func (box *Box) solutions(graph Graph, bm *roaring.Bitmap, solution []string, first rune) [][]string {
	if len(solution) > box.max {
		// log.Debug().Strs("solution", solution).Str("first", string(first)).Msg("exceed")
		return nil
	}
	if bm.GetCardinality() == letters {
		return [][]string{solution}
	}
	// if log.Logger.GetLevel() == zerolog.DebugLevel {
	// 	_, ok := graph[first]
	// 	if !ok {
	// 		log.Debug().Strs("solution", solution).Str("first", string(first)).Msg("graph")
	// 	}
	// }
	var solutions [][]string
	for _, next := range graph[first] {
		union := roaring.Or(bm, bitmap(next))
		if union.GetCardinality() == bm.GetCardinality() {
			// log.Debug().Strs("solution", solution).Str("next", next).Str("first", string(first)).Msg("cardinality")
			continue
		}
		last := rune(next[len(next)-1])
		soln := append(slices.Clone(solution), next)
		solutions = append(slices.Clone(solutions), box.solutions(graph, union, soln, last)...)
	}
	return solutions
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
	solutions := make(chan []string)
	for i := 0; i < box.con; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range wc {
				bm := bitmap(word)
				last := rune(word[len(word)-1])
				for _, sol := range box.solutions(graph, bm, []string{word}, last) {
					solutions <- sol
				}
			}
		}()
	}
	go func() {
		defer close(solutions)
		wg.Wait()
	}()
	return solutions
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
			var count int
			trie := NewTrie()
			for i := 0; i < c.NArg(); i++ {
				if err := func() error {
					fp, err := os.Open(c.Args().Get(i))
					if err != nil {
						return err
					}
					defer fp.Close()
					n, err := trie.Load(fp)
					if err != nil {
						return err
					}
					count += n
					return nil
				}(); err != nil {
					return err
				}
			}
			box := NewBox(
				WithSides(c.String("box")),
				WithConcurrent(c.Int("concurrent")),
				WithMinWordLength(c.Int("min")),
				WithMaxSolutionLength(c.Int("max")))
			words := box.Words(trie)
			log.Info().Int("matching", len(words)).Int("possible", count).Msg("dictonary")

			start := time.Now()
			solutions := map[int]int{}
			enc := json.NewEncoder(c.App.Writer)
			for solution := range box.Solutions(words) {
				solutions[len(solution)]++
				if err := enc.Encode(solution); err != nil {
					return err
				}
			}
			log.Info().Interface("solutions", solutions).Dur("elapsed", time.Since(start)).Msg("find solutions")

			return nil
		},
	}
}
