#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)
strategies=(
    '-s bigram'
    '-s el --start tares'
    '-s freq'
    '-s pos'
    '-s freq -s el --start tares'
    '-s freq -s bigram'
    '-s freq -s pos -s bigram'
    '-s freq -s pos'
)
words=$(cat $repo/data/possible.txt | head -2000 | sort -R)

regression=${repo}/dist/regression.csv
jq -rn '["secret", "strategy", "rounds", "success", "dictionary", "elapsed"] | @csv' > $regression
for strategy in ${strategies}
do
    print -P "%F{magenta}\nRunning {$strategy} strategy...%f\n" 1>&2
    eval "$repo/dist/qordle play -B -S $strategy" <<< $words |
    jq -r -s '
        map([.secret, .strategy, (.rounds|length), (.rounds|last|.success), (.rounds|last|.dictionary), .elapsed])
        | .[]
        | @csv
    ' >> $regression
done
