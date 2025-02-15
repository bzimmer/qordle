#!/usr/bin/env zsh

set -eo pipefail

repo=$(git rev-parse --show-toplevel)
regression="${repo}/dist/regression.csv"

print -P "%F{magenta}\nFailed to find any solution!%f\n" 1>&2
duckdb -s '
    select
        secret, strategy, success
    from
        read_csv("/dev/stdin")
    where
        success is false
    order by
        strategy, secret;
' < ${regression}

print -P "%F{magenta}\nFailed to find the solution in six rounds!%f\n" 1>&2
duckdb -s '
    select
        secret, strategy, rounds, success, elapsed
    from
        read_csv("/dev/stdin")
    where
        rounds > 6
    order by
        rounds, secret, strategy;
' < ${regression}

print -P "%F{magenta}\nHistograms%f\n" 1>&2
duckdb -s '
    select
        strategy, rounds, count(rounds) as bins
    from
        read_csv("/dev/stdin")
    group by
        strategy, rounds
    order by
        len(strategy), strategy, rounds, bins;
' < ${regression}

print -P "%F{magenta}\nWinners%f\n" 1>&2
duckdb -s '
    create table
        totals
    as select
        strategy, count(*) as total from read_csv("/dev/stdin") group by strategy;
    create table
        stdin
    as select * from read_csv("/dev/stdin");
    select
        stdin.strategy, count(stdin.rounds) as winners, totals.total, 100 * (winners / totals.total) as pct
    from
        stdin, totals
    where
        stdin.rounds <= 6
        and stdin.strategy = totals.strategy
    group by
        stdin.strategy, totals.total
    order by
        winners desc, stdin.strategy;
' < ${regression}
