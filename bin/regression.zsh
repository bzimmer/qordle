#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)

strategies=('-s freq' '-s pos' '-s bigram' '-s freq -s pos' '-s freq -s bigram')
for strategy in ${strategies}
do
    cmd="$repo/dist/qordle play -B -S $strategy"
    print -P "%F{magenta}\nRunning {$strategy} strategy...%f\n" 1>&2
    cat $repo/data/possible.txt |
    eval ${cmd} |
    gojq -r -s '
        map([.secret, .strategy, (.rounds|length), (.rounds|last|.success), (.rounds|last|.dictionary), .elapsed])
        | .[]
        | @csv
    ' >> $repo/dist/regression.csv
done
