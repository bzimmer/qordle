package qordle

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

type Graph map[rune]map[rune][]string

type Solution struct {
	words   []string
	letters map[rune]struct{}
}

func (sol *Solution) String() string {
	return fmt.Sprintf("%s", sol.words)
}

func (sol *Solution) Add(v string) {
	sol.words = append(sol.words, v)
	if sol.letters == nil {
		sol.letters = make(map[rune]struct{})
	}
	for _, v := range v {
		sol.letters[v] = struct{}{}
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

func (box *Box) Solutions(words []string) any {
	graph := make(map[rune]map[rune][]string)
	for _, word := range words {
		first, last := rune(word[0]), rune(word[len(word)-1])
		if _, ok := graph[first]; !ok {
			graph[first] = make(map[rune][]string)
		}
		graph[first][last] = append(graph[first][last], word)
	}
	return graph
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
