package qordle

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const letters = 12
const concurrency = 5

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
	len   int // minimum word chain length to be a solution
	sides []string
}

func NewBox(box string, min, length int) *Box {
	sides := strings.Split(box, "-")
	return &Box{sides: sides, min: min, len: length}
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

func (box *Box) solutions(graph Graph, bm *roaring.Bitmap, words []string, first rune) [][]string {
	if len(words) == box.len {
		return nil
	}
	var solutions [][]string
	for _, next := range graph[first] {
		u := roaring.Or(bm, bitmap(next))
		if u.GetCardinality() == letters {
			return append(solutions, append(words, next))
		}
		last := rune(next[len(next)-1])
		solutions = append(solutions, box.solutions(graph, u, append(words, next), last)...)
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
	for i := 0; i < concurrency; i++ {
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
				Name:  "len",
				Usage: "maximum solution length",
				Value: 2,
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
			box := NewBox(c.String("box"), c.Int("min"), c.Int("len"))
			words := box.Words(trie)
			log.Info().Int("words", len(words)).Int("dictonary", count).Msg("possible")

			m := map[int]int{}
			enc := json.NewEncoder(c.App.Writer)
			for solution := range box.Solutions(words) {
				m[len(solution)]++
				if err := enc.Encode(solution); err != nil {
					return err
				}
			}
			log.Info().Interface("combinations", m).Msg("solutions")

			return nil
		},
	}
}
