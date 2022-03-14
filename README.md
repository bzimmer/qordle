# qordle

![build](https://github.com/bzimmer/qordle/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/qordle/branch/main/graph/badge.svg?token=8L0KXGSM9N)](https://codecov.io/gh/bzimmer/qordle)

Simple wordle solution suggester.

## Usage

`qordle` uses the hits, misses, and a pattern (if known) to suggest words matching the solution.

## Input

* Input correctly placed letters as an uppercase
* Input incorrectly placed letters as a lowercase letter preceeded by a `~` (tilde)
* Input missesas as lowercase letters

## Example

![Screenshot](screenshot.png)

By default the letter frequency strategy orders the possible words.

```sh
~ > qordle b~rAin stARt peARl
["chard","hoard","dwarf","wharf","award","guard","charm","ovary","quark"]
```

Specify the `-A` flag to order the words alphabetically.

```sh
~ > qordle -A b~rAin stARt peARl
["award","chard","charm","dwarf","guard","hoard","ovary","quark","wharf"]
```
