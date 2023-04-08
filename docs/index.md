# Welcome to qordle

![build](https://github.com/bzimmer/qordle/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/qordle/branch/main/graph/badge.svg?token=8L0KXGSM9N)](https://codecov.io/gh/bzimmer/qordle)

## Commands

* `qordle play` - Automatically play a wordle game using a specified strategy
* `qordle suggest` - Suggest the next word using the specified strategy and exiting guesses

## Strategies

`qordle` supports a number of different strategies and those strategies can be chain
to build new strategies.

```txt
Winners out of 2000 randomly chosen words

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
