package qordle

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type Graph map[rune]map[rune]map[string]string

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

func (box *Box) hash(v string) string {
	s := []rune(v)
	sort.SliceStable(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	if len(s) < 2 {
		return string(s)
	}
	e := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			continue
		}
		s[e] = s[i]
		e++
	}
	return string(s[:e])
}

func (box *Box) Solutions(words []string) any {
	var unique int
	graph := make(map[rune]map[rune]map[string]string)
	for _, word := range words {
		first, last, m := rune(word[0]), rune(word[len(word)-1]), box.hash(word)
		if _, ok := graph[first]; !ok {
			graph[first] = map[rune]map[string]string{last: {}}
		}
		if _, ok := graph[first][last]; !ok {
			graph[first][last] = map[string]string{}
		}
		if _, ok := graph[first][last][m]; !ok {
			unique++
			graph[first][last][m] = word
		}
	}
	log.Info().Int("unique", unique).Int("words", len(words)).Msg("solutions")
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
				Value: "afh-mie-pow-bwu",
			},
			&cli.IntFlag{
				Name:  "min",
				Usage: "minimum word size",
				Value: 3,
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
					scanner := bufio.NewScanner(fp)
					for scanner.Scan() {
						count++
						trie.Add(scanner.Text())
					}
					return scanner.Err()
				}(); err != nil {
					return err
				}
			}
			log.Info().Int("words", count).Msg("possible")
			box := NewBox(c.String("box"), c.Int("min"))
			words := box.Words(trie)
			enc := json.NewEncoder(c.App.Writer)
			return enc.Encode(box.Solutions(words))
		},
	}
}
