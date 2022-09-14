package qordle

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type Graph map[rune]map[rune]map[string]string

type Set struct {
	values map[rune]struct{}
}

func (set *Set) Add(v string) {
	if set.values == nil {
		set.values = make(map[rune]struct{})
	}
	for _, v := range v {
		set.values[v] = struct{}{}
	}
}

func (set *Set) Union(s *Set) {
	for key, value := range s.values {
		set.values[key] = value
	}
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
	min   int
	sides []string
}

func NewBox(box string, min int) *Box {
	sides := strings.Split(box, "-")
	return &Box{sides: sides, min: min}
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

func hash(v string) string {
	s := []rune(v)
	sort.SliceStable(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	if len(s) < 2 {
		return string(s)
	}
	var e int = 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			continue
		}
		s[e] = s[i]
		e++
	}
	return string(s[:e])
}

// func (box *Box) solutions(graph Graph, paths []string, letters *Set, next rune) [][]string {
// 	// if len(letters.values) == 12 {
// 	// 	return [][]string{paths}
// 	// }
// 	// if len(paths) == 5 {
// 	// 	return [][]string{}
// 	// }
// 	// fmt.Fprintf(os.Stderr, "%s\n", paths)

// 	solutions := [][]string{}
// 	// for last, words := range graph[next] {
// 	// 	for _, word := range words {
// 	// 		// if !slices.Contains(paths, word) {
// 	// 		// 	log.Warn().Strs("paths", paths).Str("word", word).Interface("solutions", solutions).Msg("solutions")
// 	// 		// 	continue
// 	// 		// }
// 	// 		log.Info().Strs("paths", paths).Str("word", word).Interface("solutions", solutions).Msg("solutions")
// 	// 		set := new(Set)
// 	// 		// fmt.Fprintf(os.Stderr, "%s\n", word)
// 	// 		set.Add(word)
// 	// 		set.Union(letters)
// 	// 		solutions = append(solutions, box.solutions(graph, append(paths, word), set, last)...)
// 	// 	}
// 	// }
// 	return solutions
// }

func (box *Box) Solutions(words []string) any {
	graph := make(map[rune]map[rune]map[string]string)
	for _, word := range words {
		first, last := rune(word[0]), rune(word[len(word)-1])
		if _, ok := graph[first]; !ok {
			graph[first] = map[rune]map[string]string{last: {}}
		}
		if _, ok := graph[first][last]; !ok {
			graph[first][last] = map[string]string{}
		}
		graph[first][last][hash(word)] = word
	}

	log.Debug().Interface("graph", graph).Msg("solutions")

	solutions := [][]string{}
	// for _, lasts := range graph {
	// 	for last, words := range lasts {
	// 		solutions = append(solutions, box.solutions(graph, words, new(Set), last)...)
	// 	}
	// }
	// fmt.Fprintf(os.Stderr, "%s\n", solutions)
	return []any{solutions, graph}
}

func CommandLetterBox() *cli.Command {
	return &cli.Command{
		Name:  "letterbox",
		Usage: "play letterbox",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "box",
				Usage: "the letter box in `aaa-bbb-ccc-ddd` format",
				Value: "mar-sej-hdw-opq",
			},
			&cli.IntFlag{
				Name:  "min",
				Usage: "minimum word size",
				Value: 3,
			},
		},
		Action: func(c *cli.Context) error {
			trie := NewTrie()
			box := NewBox(c.String("box"), c.Int("min"))
			for i := 0; i < c.NArg(); i++ {
				if err := func() error {
					fp, err := os.Open(c.Args().Get(i))
					if err != nil {
						return err
					}
					defer fp.Close()
					scanner := bufio.NewScanner(fp)
					for scanner.Scan() {
						trie.Add(scanner.Text())
					}
					return scanner.Err()
				}(); err != nil {
					return err
				}
			}
			words := box.Words(trie)
			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(box.Solutions(words))
		},
	}
}
