# Home

![build](https://github.com/bzimmer/qordle/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/qordle/branch/main/graph/badge.svg?token=8L0KXGSM9N)](https://codecov.io/gh/bzimmer/qordle)

## Introduction

Beat your friends & family at [wordle](https://www.nytimes.com/games/wordle/index.html)!

## Web Solver

`qordled` is the companion web service. Open it in a browser, type your Wordle guesses, click tiles to
set their colours (gray / yellow / green), and suggestions appear automatically.

Use the **Strategy** pill toggles to choose which ranking algorithms are applied. You can select any
combination; selection order has no effect because strategies are
[chained](strategies.md#chaining) by combining their independent results.

![Web Solver Screenshot](https://github.com/user-attachments/assets/d3a1a50c-5e23-40fe-84b1-bfde6feae470)

## Installation

```shell
$ brew install bzimmer/tap/qordle
```

## CLI Examples

```shell title="Suggest words using the elimination strategy after the first guess of 'brain'"
$ qordle suggest -s el brAin
```

```shell title="Auto-play with frequency, position, and bigrams strategies for 'ledge'"
$ qordle play -s f -s p -s bi ledge | jq ".rounds | last"
{
  "dictionary": 1,
  "scores": [
    "rat.es",
    "moniE",
    "LEDGE"
  ],
  "words": [
    "rates",
    "monie",
    "ledge"
  ],
  "success": true
}
```
