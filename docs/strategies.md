## Strategies

`qordle` supports a number of different word selection strategies as documented below.
The strategies are designed to be composable via [chaining](#chaining).

### alpha
The alpha strategy sorts the word list alphabetically

### bigram
The bigram strategy sums the bigrams for each wordin using the
[bigram frequency table](https://github.com/bzimmer/qordle/blob/main/tables.go)
and sorts the word list highest to lowest

### elimination

{==

Note: This strategy is far slower than the rest so best used when the word list has
been filtered.

==}

The elimination strategy uses each word in the word list as a secret and scores all the
remaining words against it. The accumulated sum for the letter & position from
[position frequency table](https://github.com/bzimmer/qordle/blob/main/tables.go) is used
for sorting.

* the table value for a *Misplaced* position
* two times the table value for an *Exact* position

### frequency
The frequency strategy iterates the word list accumulating the letter frequency for all
remaining words in the list. Each word is then scored by summing its letter frequencies.

### position
The position strategy, similar to the [frequency](#frequency) strategy, iterates the word
list accumulating the position frequency for each letter. Each word is then scored by
summing its letter position.

### speculation
The speculation strategy is used for solving "guessing games", those situation when all
remaining words differ by only a single letter. The strategy iterates the word list
accumulating the differing letter and then generates a word list from those words composed
of the unknown letters.

## Chaining
All strategies are composable via chaining. The chaining, itself a strategy, executes
all child strategies and sorts by accumulating the word rank in resulting word list.

## Performance

The following table shows the number of winning rounds from 2000 randomly chosen words
using different strategies.

|                         strategy                         | winners | total |  pct  |
|----------------------------------------------------------|--------:|-------|-------|
| speculate{chain{frequency,position}}                     |    1870 |  2000 | 93.5  |
| speculate{chain{frequency,elimination}}                  |    1860 |  2000 | 93.0  |
| speculate{chain{frequency,position,bigram,elimination}}  |    1858 |  2000 | 92.9  |
| speculate{chain{frequency,position,bigram}}              |    1858 |  2000 | 92.9  |
| speculate{chain{frequency,elimination,bigram}}           |    1851 |  2000 | 92.5  |
| speculate{chain{frequency,bigram}}                       |    1846 |  2000 | 92.3  |
| speculate{elimination}                                   |    1839 |  2000 | 92.0  |
| speculate{frequency}                                     |    1834 |  2000 | 91.7  |
| speculate{position}                                      |    1778 |  2000 | 88.9  |
| speculate{bigram}                                        |    1597 |  2000 | 79.8  |