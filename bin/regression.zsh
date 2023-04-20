#!/usr/bin/env zsh

set -eo pipefail

N="${1:-2000}"
repo=$(git rev-parse --show-toplevel)
strategies=(
    '-s bigram'
    '-s el --start tares'
    '-s freq -s bigram'
    '-s freq -s el --start tares'
    '-S -s freq -s el --start tares'
    '-s freq -s el -s bigram --start tares'
    '-s freq -s pos -s bigram -s el --start tares'
    '-s freq -s pos -s bigram'
    '-s freq -s pos'
    '-S -s freq -s pos'
    '-s freq'
    '-S -s freq'
    '-s pos'
    '-S -s pos'
)
words=$(cat ${repo}/data/possible.txt | sort -R | tail -n ${N})

progress=""
if [[ -t 1 ]] ; then progress="-B" ; fi

regression=${repo}/dist/regression.csv
jq -rn '["secret", "strategy", "rounds", "success", "dictionary", "elapsed"] | @csv' > ${regression}
for strategy in ${strategies}
do
    print -P "%F{magenta}\nRunning {$strategy} strategy...%f\n" 1>&2
    eval "${repo}/dist/qordle play ${progress} ${strategy}" <<< ${words} |
    jq -r -s '
        map([.secret, .strategy, (.rounds|length), (.rounds|last|.success), (.rounds|last|.dictionary), .elapsed])
        | .[]
        | @csv
    ' >> ${regression}
done
