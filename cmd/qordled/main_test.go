package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // shared read-only test defaults
var (
	defaultWordlists = []string{"solutions", "possible"}
	defaultPrefer    = []string{"solutions"}
)

func TestSuggest(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		target string
		status int
	}{
		{name: "no guesses", target: "/qordle/suggest", status: http.StatusOK},
		{name: "trailing slash", target: "/qordle/suggest/", status: http.StatusOK},
		{name: "one guess", target: "/qordle/suggest/B.rAin", status: http.StatusOK},
		{name: "two guesses", target: "/qordle/suggest/B.rAin%20BO.reD", status: http.StatusOK},
		{name: "solved", target: "/qordle/suggest/BRAIN", status: http.StatusOK},
		{name: "invalid guess", target: "/qordle/suggest/brai.", status: http.StatusBadRequest},
		{name: "get", target: "/qordle/suggest/B.rAin", status: http.StatusOK},
		{name: "strategy prefix", target: "/qordle/suggest/B.rAin?strategy=elim", status: http.StatusOK},
		{name: "unknown strategy", target: "/qordle/suggest/brain?strategy=nope", status: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			handler, err := newHandler("", defaultWordlists, defaultPrefer)
			require.NoError(t, err)
			method := http.MethodPost
			if tt.name == "get" {
				method = http.MethodGet
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(method, tt.target, nil))
			a.Equal(tt.status, rec.Code)
			if tt.status != http.StatusOK {
				return
			}
			var words []string
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &words))
			a.NotEmpty(words)
		})
	}
}

func TestStrategies(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	handler, err := newHandler("", defaultWordlists, defaultPrefer)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/qordle/strategies", nil))
	a.Equal(http.StatusOK, rec.Code)
	var strategies map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &strategies))
	a.Len(strategies, 5)
	for name, desc := range strategies {
		a.NotEmpty(desc, name)
	}
}

func TestPlay(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		target string
		status int
	}{
		{name: "solves", target: "/qordle/play/board", status: http.StatusOK},
		{name: "solves outside solutions", target: "/qordle/play/peeve", status: http.StatusOK},
		{name: "wrong length", target: "/qordle/play/boards", status: http.StatusBadRequest},
		{name: "post not allowed", target: "/qordle/play/board", status: http.StatusMethodNotAllowed},
		{name: "unknown route", target: "/qordle/nope", status: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler, err := newHandler("", defaultWordlists, defaultPrefer)
			require.NoError(t, err)
			method := http.MethodGet
			if tt.status == http.StatusMethodNotAllowed {
				method = http.MethodPost
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(method, tt.target, nil))
			assert.Equal(t, tt.status, rec.Code)
			if tt.status == http.StatusOK {
				var board struct {
					Target string `json:"target"`
					Rounds []struct {
						Success bool `json:"success"`
					} `json:"rounds"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &board))
				assert.Equal(t, strings.TrimPrefix(tt.target, "/qordle/play/"), board.Target)
				assert.True(t, board.Rounds[len(board.Rounds)-1].Success)
			}
		})
	}
}

func TestSuggestTiered(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	handler, err := newHandler("", defaultWordlists, defaultPrefer)
	require.NoError(t, err)
	// crane then slime scored against peeve, which is outside solutions.txt
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/qordle/suggest/cranE%20slimE", nil))
	a.Equal(http.StatusOK, rec.Code)
	var words []string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &words))
	a.Contains(words, "peeve")
	// solutions-list candidates come first
	a.Equal("budge", words[0])
	a.Less(slices.Index(words, "judge"), slices.Index(words, "peeve"))
}

func TestWordlistFlags(t *testing.T) {
	t.Parallel()
	// crane then slime scored against peeve, which is outside solutions.txt
	const target = "/qordle/suggest/cranE%20slimE"
	for _, tt := range []struct {
		name            string
		wordlists       []string
		prefer          []string
		peeve, ordering bool
	}{
		{name: "defaults", wordlists: defaultWordlists, prefer: defaultPrefer, peeve: true, ordering: true},
		{name: "solutions only", wordlists: []string{"solutions"}, prefer: defaultPrefer},
		{name: "flat", wordlists: defaultWordlists, prefer: []string{""}, peeve: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			handler, err := newHandler("", tt.wordlists, tt.prefer)
			require.NoError(t, err)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, target, nil))
			a.Equal(http.StatusOK, rec.Code)
			var words []string
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &words))
			a.Equal(tt.peeve, slices.Contains(words, "peeve"))
			if tt.ordering {
				a.Less(slices.Index(words, "judge"), slices.Index(words, "peeve"))
			}
		})
	}
	_, err := newHandler("", []string{"nope"}, defaultPrefer)
	require.Error(t, err)
	_, err = newHandler("", []string{""}, defaultPrefer)
	require.Error(t, err)
}
