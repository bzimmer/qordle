#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)
games=$(cat ${repo}/dist/regression.csv)

print -P "%F{magenta}\nFailed to find any solution!%f\n" 1>&2
csvq -c --no-header '
    with
        games as
        (
            select
                c1 as secret, c2 as strategy, c3 as rounds, c4 as success, c5 as dictionary, c6 as elapsed
            from
                stdin
        )
    select
        secret, strategy, rounds, success, elapsed from games where success is false
' <<< $games

print -P "%F{magenta}\nFailed to find the solution in six rounds!%f\n" 1>&2
csvq -c --no-header '
    with
        games as
        (
            select
                c1 as secret, c2 as strategy, c3 as rounds, c4 as success, c5 as dictionary, c6 as elapsed
            from
                stdin
        )
    select
        secret, strategy, rounds, success, elapsed from games where rounds > 6
    order by
        rounds, secret, strategy
' <<< $games

print -P "%F{magenta}\nSummary statistics%f\n" 1>&2
csvq -c --no-header '
    with
        games as
        (
            select
                c1 as secret, c2 as strategy, c3 as rounds, c4 as success, c5 as dictionary, c6 as elapsed
            from
                stdin
        )
    select
        strategy, avg(rounds) as avg_rounds, stdev(rounds) as stdev_rounds, max(rounds) as max_rounds, min(rounds) as min_rounds, avg(elapsed) as avg_elapsed
    from
        games
    group by
        strategy
    order by
        strategy
' <<< $games
