package qordle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli/v2"
)

const (
	spellingBeeURL      = "https://www.nytimes.com/puzzles/spelling-bee"
	spellingBeeGameData = "window.gameData\\s*=\\s*(.*)"
)

type Bee struct {
	// Expiration     int      `json:"expiration"`
	// DisplayWeekday string   `json:"displayWeekday"`
	// DisplayDate    string   `json:"displayDate"`
	// PrintDate      string   `json:"printDate"`
	CenterLetter string   `json:"centerLetter"`
	OuterLetters []string `json:"outerLetters"`
	// ValidLetters []string `json:"validLetters"`
	Pangrams []string `json:"pangrams"`
	Answers  []string `json:"answers"`
	// ID             int      `json:"id"`
	// FreeExpiration int      `json:"freeExpiration"`
	// Editor         string   `json:"editor"`
}

func today(c *cli.Context) error {
	req, err := http.NewRequestWithContext(c.Context, http.MethodGet, spellingBeeURL, http.NoBody)
	if err != nil {
		return err
	}
	res, err := Runtime(c).Grab.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	enc := Runtime(c).Encoder
	re := regexp.MustCompile(spellingBeeGameData)
	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		res := re.FindAllStringSubmatch(s.Text(), -1)
		for i := range res {
			var bees struct {
				Today *Bee `json:"today"`
				// Yesterday *Bee `json:"yesterday"`
			}
			dec := json.NewDecoder(strings.NewReader(res[i][1]))
			err = dec.Decode(&bees)
			if err != nil {
				err = fmt.Errorf("failed to decode %w", err)
				return false
			}
			err = enc.Encode(bees.Today)
			if err != nil {
				err = fmt.Errorf("failed to encode %w", err)
				return false
			}
		}
		return true
	})

	return err
}

func CommandBee() *cli.Command {
	return &cli.Command{
		Name:  "bee",
		Usage: "manage nyt spelling bee",
		Subcommands: []*cli.Command{
			{
				Name:   "today",
				Action: today,
			},
		},
	}
}
