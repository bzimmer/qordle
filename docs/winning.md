## Workflow

The original purpose for `qordle` was to solve the daily
[wordle](https://www.nytimes.com/games/wordle/index.html) puzzle faster than my
family. The `qordle` [suggest](commands.md#suggest) command will quickly perform
a ranking of the best next guesses. I wrote a simple shell script to facilitate
using two different [strategy](strategies.md) combinations and `jq` to show only
the top ten words.

``` zsh title="qordle-suggest"
#!/usr/bin/env zsh

set -eo pipefail

strategies=(
    # '-s freq'
    # '-s pos'
    # '-s bigram'
    '-s freq -s pos'
    # '-s freq -s bigram'
    '-s freq -s pos -s bigram'
)
for strategy in ${strategies}
do
    eval "qordle suggest $strategy -S "$@"" |
        jq --arg strategy $strategy '{strategy:$strategy, words:.[:10]}'
done
```

An example session of a few guesses might follow the below pattern.

``` zsh title="qordle suggest session"
$ qordle-suggest t.a.res g.r.ain m.o.l.a.r
{
  "strategy": "-s freq -s pos",
  "words": [
    "flora"
  ]
}
{
  "strategy": "-s freq -s pos -s bigram",
  "words": [
    "flora"
  ]
}
```

To see how `qordle` might have played a word using different strategies, use the
[play](commands.md#play) command. The following is a short script I use to
abbreviate the results.

``` zsh title="qordle-play"
#!/usr/bin/env zsh

set -eo pipefail

qordle play -s frequency -s position -S "$@" |
    jq -s 'map({secret:.secret, words:(.rounds|last|.words)})'
```

If no starting word is provided, `qordle` will automatically find the optimal
starting word for the specified strategies. In this case it ranked *tares*
best.

``` zsh title="play with freq and pos strategies"
$ qordle-play under
[
  {
    "secret": "under",
    "words": [
      "tares",
      "cider",
      "older",
      "under"
    ]
  }
]
```

[Chaining](strategies.md#chaining) the [bigram](strategies.md#bigram) strategy would
have resulted in a faster path to the solution.

``` zsh title="play with freq, pos, and bigram strategies"
qordle play -s freq -s pos -s b -S under | jq -s 'map({secret:.secret, words:(.rounds|last|.words)})'
[
  {
    "secret": "under",
    "words": [
      "rates",
      "diner",
      "under"
    ]
  }
]
```
