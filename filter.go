package qordle

import (
	"errors"
	"unicode"
)

type state struct {
	exact     rune
	forbidden map[rune]struct{}
}

type FilterFunc func(string) bool

func Filter(words Dictionary, fns ...FilterFunc) Dictionary {
	var count int
	res := make([]string, len(words))
	for _, word := range words {
		matches := true
		for i := 0; matches && i < len(fns); i++ {
			matches = fns[i](word)
		}
		if matches {
			res[count] = word
			count++
		}
	}
	return res[:count]
}

func NoOp(_ string) bool {
	return true
}

func IsLower() FilterFunc {
	return func(word string) bool {
		return unicode.IsLower(rune(word[0]))
	}
}

func Length(length int) FilterFunc {
	return func(word string) bool {
		return len(word) == length
	}
}

func filter(ms []*state, rq map[rune]int) FilterFunc {
	return func(word string) bool {
		ws, rs := []rune(word), make(map[rune]int)
		for i := range ws {
			rs[ws[i]]++
			if ms[i].exact != 0 && ms[i].exact != ws[i] {
				return false
			}
			if _, ok := ms[i].forbidden[ws[i]]; ok {
				return false
			}
		}
		for key, val := range rq {
			num, ok := rs[key]
			if !ok || num < val {
				return false
			}
		}
		return true
	}
}

func compile(marks map[rune]map[int]Mark, ix int) FilterFunc { //nolint:gocognit
	ms, rq := make([]*state, ix), make(map[rune]int)
	// prepare the table
	for i := 0; i < ix; i++ {
		ms[i] = &state{
			forbidden: make(map[rune]struct{}),
		}
	}
	// populate the table
	for letter, states := range marks {
		for index, mark := range states {
			switch mark {
			case MarkExact:
				rq[letter]++
				ms[index].exact = letter
			case MarkMisplaced:
				rq[letter]++
				ms[index].forbidden[letter] = struct{}{}
			case MarkMiss:
				/*
					if the same letter exists as a misplaced or exact mark for any
					index, then only add the miss to the current index otherwise add
					it to all indices

					go run cmd/qordle/main.go --debug validate cleat .legwl | jq
					.l.egAl => the second l should put itself only in it's slot?
					=> feedback [egwl]  [egwl] [gwle] [egwl] [egwl]
					=> regex    [^egwl] [^weg] [^egw] [^egw] [^egwl]
				*/
				all := true
				for i := 0; all && i < ix; i++ {
					switch marks[letter][i] {
					case MarkMiss:
						// ignore
					case MarkMisplaced, MarkExact:
						all = false
					}
				}
				if all {
					for i := 0; i < ix; i++ {
						ms[i].forbidden[letter] = struct{}{}
					}
				} else {
					ms[index].forbidden[letter] = struct{}{}
				}
			}
		}
	}
	// if zerolog.GlobalLevel() == zerolog.DebugLevel {
	// 	var buf []string
	// 	for i := range ms {
	// 		var bar []string
	// 		for letter := range ms[i].forbidden {
	// 			bar = append(bar, string(letter))
	// 		}
	// 		buf = append(buf, "["+strings.Join(bar, "")+"]")
	// 	}
	// 	log.Debug().Str("pattern", strings.Join(buf, " ")).Any("required", rq).Msg("compile")
	// }
	return filter(ms, rq)
}

func parse(feedback string) (FilterFunc, error) {
	ix, rs := 0, []rune(feedback)
	marks := make(map[rune]map[int]Mark)
	for i := 0; i < len(rs); i++ {
		var mark Mark
		switch {
		case unicode.IsSpace(rs[i]):
			fallthrough
		case unicode.IsNumber(rs[i]):
			fallthrough
		case unicode.IsLower(rs[i]):
			mark = MarkMiss
		case unicode.IsUpper(rs[i]):
			mark = MarkExact
		default:
			i++
			if i >= len(rs) {
				return nil, errors.New("too few characters")
			}
			mark = MarkMisplaced
		}
		lower := unicode.ToLower(rs[i])
		hit, ok := marks[lower]
		if !ok {
			hit = make(map[int]Mark)
			marks[lower] = hit
		}
		hit[ix] = mark
		ix++
	}
	return compile(marks, ix), nil
}

func Guesses(guesses ...string) (FilterFunc, error) {
	var fns []FilterFunc
	for _, guess := range guesses {
		if guess == "" {
			continue
		}
		f, err := parse(guess)
		if err != nil {
			return nil, err
		}
		fns = append(fns, f)
	}
	if len(fns) == 0 {
		return NoOp, nil
	}
	return func(word string) bool {
		for _, fn := range fns {
			if !fn(word) {
				return false
			}
		}
		return true
	}, nil
}
