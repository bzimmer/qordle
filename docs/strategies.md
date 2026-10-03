## Strategies

`qordle` supports a number of different word selection strategies as documented below.
The strategies are designed to be composable via [chaining](#chaining).

### alpha
The alpha strategy sorts the word list alphabetically

### bigram
The bigram strategy sums the bigrams for each word using the
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

### entropy
The entropy strategy scores each word by the expected information its feedback reveals,
treating every remaining word as equally likely to be the secret. A word splitting the
remaining words into many small groups of identical feedback ranks above one leaving a few
large groups. Like [elimination](#elimination) it compares every word with every other, but
encodes feedback as one of 3<sup>5</sup> patterns without allocating, so it stays fast on
filtered lists.

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

### probe
The probe option (`--probe` on the CLI, the **probe** pill in the web solver) wraps any
strategy. It considers every accepted word as the next guess, including words already ruled
out, and leads with one when its [entropy](#entropy) beats guessing a remaining word. Each
remaining word earns a bonus for the chance of being the secret, so with only a few words left
a candidate wins. When preferred words are set (see `--prefer`) it plans against those while
any remain. It generalises [speculation](#speculation), which only probes once the remaining
words differ by a single letter.

Simulated games with `slate` as the opener, the solutions ranked first, and the 63 answers
since 2022 missing from the solutions list held out as unseen answers:

| strategy | solved in six (1000 original / 63 newer) | average guesses |
|---|---|---|
| frequency, position | 99.6% / 96.8% | 3.60 / 4.48 |
| frequency, position + speculate | 99.8% / 96.8% | 3.62 / 4.51 |
| entropy | 99.8% / 95.2% | 3.54 / 4.44 |
| frequency, position + probe | 100% / 100% | 3.47 / 4.33 |
| entropy + probe | 100% / 98.4% | 3.47 / 4.43 |

## Chaining
All strategies are composable via chaining. The chaining strategy, itself a strategy, executes
all child strategies **concurrently** on the same word list and combines the results by accumulating
each word's normalised rank position across every child strategy. Because this combination is a
simple sum — a commutative operation — **the order in which strategies are specified does not
matter**. `chain{frequency,position}` and `chain{position,frequency}` produce identical output.

This also means the pill-toggle selector in the web solver works perfectly: selecting or
deselecting strategies in any order gives the same result as specifying them on the CLI in
any order.

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