package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{name: "unknown strategy", target: "/qordle/suggest/brain?strategy=nope", status: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			engine, err := newEngine()
			require.NoError(t, err)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, tt.target, nil))
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
	engine, err := newEngine()
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/qordle/strategies", nil))
	a.Equal(http.StatusOK, rec.Code)
	var strategies map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &strategies))
	a.Len(strategies, 5)
	for name, desc := range strategies {
		a.NotEmpty(desc, name)
	}
}
