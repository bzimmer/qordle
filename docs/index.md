# Home

![build](https://github.com/bzimmer/qordle/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/qordle/branch/main/graph/badge.svg?token=8L0KXGSM9N)](https://codecov.io/gh/bzimmer/qordle)

## Introduction

Beat your friends & family at [wordle](https://www.nytimes.com/games/wordle/index.html)!

## Installation

```shell
$ brew install bzimmer/tap/qordle
```

## Examples

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
