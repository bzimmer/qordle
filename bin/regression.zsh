#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)
strategies=(
    '-s freq'
    '-s pos'
    '-s bigram'
    '-s freq -s pos'
    '-s freq -s bigram'
    '-s freq -s pos -s bigram'
)
regression=${repo}/dist/regression.csv

jq -rn '["secret", "strategy", "rounds", "success", "dictionary", "elapsed"] | @csv' > $regression
for strategy in ${strategies}
do
    print -P "%F{magenta}\nRunning {$strategy} strategy...%f\n" 1>&2
    cat $repo/data/possible.txt |
    eval "$repo/dist/qordle play -B -S $strategy" |
    gojq -r -s '
        map([.secret, .strategy, (.rounds|length), (.rounds|last|.success), (.rounds|last|.dictionary), .elapsed])
        | .[]
        | @csv
    ' >> $regression
done
