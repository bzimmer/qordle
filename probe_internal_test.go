package qordle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeedbackMatchesCheck(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	words, err := Read("solutions")
	require.NoError(t, err)
	// include repeated letters on both sides, the case most likely to diverge
	pairs := [][2]string{{"peeve", "eerie"}, {"abbey", "babes"}, {"llama", "algal"}, {"speed", "erase"}}
	for i := 0; i < len(words); i += 7 {
		for j := 0; j < len(words); j += 53 {
			pairs = append(pairs, [2]string{words[i], words[j]})
		}
	}
	for _, p := range pairs {
		marks, cerr := Check(p[0], p[1])
		require.NoError(t, cerr)
		code, place := 0, 1
		for _, m := range marks[0] {
			code += int(m) * place
			place *= 3
		}
		a.Equal(code, feedback(p[0], p[1]), "secret %s guess %s", p[0], p[1])
	}
	a.Greater(len(pairs), 10000)
}
