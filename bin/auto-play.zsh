#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)

games=$(
    cat $repo/data/solutions.txt |
    $repo/dist/qordle play -A -S --start "${1:-"brain"}" |
    jq -r -s '
        map([.secret, (.rounds|length), (.rounds|last|.success), .elapsed])
        | .[] 
        | @csv
    '
)

print -P "%F{magenta}\nFailed to find any solution!%f\n"
csvq --no-header '
    with
        games as
        (
            select
                c1 as secret, c2 as rounds, c3 as success, c4 as elapsed
            from
                stdin
        )
    select
            secret, rounds, success, elapsed from games where success is false
' <<< $games

print -P "%F{magenta}\nFailed to find the solution in six rounds!%f\n"
csvq --no-header '
    with
        games as
        (
            select
                c1 as secret, c2 as rounds, c3 as success, c4 as elapsed
            from
                stdin
        )
    select
        secret, rounds, success, elapsed from games where rounds > 6
    order by
        rounds, secret
' <<< $games
