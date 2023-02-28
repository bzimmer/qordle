#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)
regression="${repo}/dist/regression.csv"

print -P "%F{magenta}\nFailed to find any solution!%f\n" 1>&2
csvq -c '
    select
        secret, strategy, success
    from
        stdin
    where
        success is false
    order by
        strategy, secret
' < ${regression}

print -P "%F{magenta}\nFailed to find the solution in six rounds!%f\n" 1>&2
csvq -c '
    select
        secret, strategy, rounds, success, elapsed
    from
        stdin
    where
        rounds > 6
    order by
        rounds, secret, strategy
' < ${regression}

# print -P "%F{magenta}\nSummary statistics%f\n" 1>&2
# csvq -c '
#     select
#         strategy,
#         avg(rounds) as avg_rounds,
#         stdev(rounds) as stdev_rounds,
#         max(rounds) as max_rounds,
#         min(rounds) as min_rounds,
#         avg(elapsed) as avg_elapsed
#     from
#         stdin
#     group by
#         strategy
#     order by
#         strategy
# ' < ${regression}

print -P "%F{magenta}\nHistograms%f\n" 1>&2
csvq -c '
    select
        strategy, rounds, count(rounds) as bins
    from
        stdin
    group by
        strategy, rounds
    order by
        len(strategy), strategy, rounds, bins
' < ${regression}
