#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)

strategies=(frequency position chain bigram)
for strategy in ${strategies}
do
    print -P "%F{magenta}\nRunning {$strategy} strategy...%f\n" 1>&2
    cat $repo/data/possible.txt |
    $repo/dist/qordle play -B -S -s $strategy |
    jq -r -s '
        map([.secret, .strategy, (.rounds|length), (.rounds|last|.success), (.rounds|last|.dictionary), .elapsed])
        | .[]
        | @csv
    ' >> $repo/dist/regression.csv
done
