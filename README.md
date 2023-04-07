# qordle

![build](https://github.com/bzimmer/qordle/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/qordle/branch/main/graph/badge.svg?token=8L0KXGSM9N)](https://codecov.io/gh/bzimmer/qordle)

Simple wordle solution suggester

## Install

```sh
$ brew tap bzimmer/tap
$ brew install qordle
```

## Usage

`qordle` uses the hits, misses, and a pattern (if known) to suggest words matching the solution.

## Input

* Input correctly placed letters as an uppercase
* Input incorrectly placed letters as a lowercase letter preceded by any symbol (`.`, `@`)
* Input misses as lowercase letters

## Example

![Screenshot](screenshot.png)

```sh
$ qordle suggest -s position -w solutions b.rAin stARt peARl
["chard","award","guard","charm","hoard","wharf","dwarf","quark","ovary"]
```

```sh
$ qordle play --start brain table | jq
{
  "secret": "table",
  "strategy": "frequency",
  "dictionary": 12947,
  "rounds": [
    {
      "dictionary": 118,
      "scores": [
        "~br~ain"
      ],
      "words": [
        "brain"
      ],
      "success": false
    },
    {
      "dictionary": 5,
      "scores": [
        "~br~ain",
        "mAB~es"
      ],
      "words": [
        "brain",
        "mabes"
      ],
      "success": false
    },
    {
      "dictionary": 4,
      "scores": [
        "~br~ain",
        "mAB~es",
        "cABLE"
      ],
      "words": [
        "brain",
        "mabes",
        "cable"
      ],
      "success": false
    },
    {
      "dictionary": 3,
      "scores": [
        "~br~ain",
        "mAB~es",
        "cABLE",
        "fABLE"
      ],
      "words": [
        "brain",
        "mabes",
        "cable",
        "fable"
      ],
      "success": false
    },
    {
      "dictionary": 2,
      "scores": [
        "~br~ain",
        "mAB~es",
        "cABLE",
        "fABLE",
        "gABLE"
      ],
      "words": [
        "brain",
        "mabes",
        "cable",
        "fable",
        "gable"
      ],
      "success": false
    },
    {
      "dictionary": 1,
      "scores": [
        "~br~ain",
        "mAB~es",
        "cABLE",
        "fABLE",
        "gABLE",
        "hABLE"
      ],
      "words": [
        "brain",
        "mabes",
        "cable",
        "fable",
        "gable",
        "hable"
      ],
      "success": false
    },
    {
      "dictionary": 1,
      "scores": [
        "~br~ain",
        "mAB~es",
        "cABLE",
        "fABLE",
        "gABLE",
        "hABLE",
        "TABLE"
      ],
      "words": [
        "brain",
        "mabes",
        "cable",
        "fable",
        "gable",
        "hable",
        "table"
      ],
      "success": true
    }
  ],
  "elapsed": 2
}
```

## Strategies

`qordle` supports a number of different strategies and those strategies can be chain
to build new strategies.

```txt
Winners

+----------------------------------------------+---------+-------+
|                   strategy                   | winners |  pct  |
+----------------------------------------------+---------+-------+
| speculate{chain{frequency,position}}         |    1930 | 96.5  |
| speculate{chain{frequency,position,bigram}}  |    1915 | 95.8  |
| speculate{chain{frequency,bigram}}           |    1903 | 95.2  |
| speculate{chain{frequency,elimination}}      |    1901 | 95.0  |
| speculate{frequency}                         |    1875 | 93.8  |
| speculate{position}                          |    1858 | 92.9  |
| speculate{elimination}                       |    1851 | 92.5  |
| speculate{bigram}                            |    1753 | 87.6  |
+----------------------------------------------+---------+-------+
```
