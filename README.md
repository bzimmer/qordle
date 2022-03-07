# qordle
Simple wordle solution suggester.

## Usage

`qordle` uses the hits, misses, and a pattern (if known) to suggest words matching the solution.

It does not:

* score or rank the possibilities
* eliminate possibilities based on prior guesses using letters in incorrect positions

It does:

* greatly eliminate the universe of possibilities

## Example

![Screenshot](screenshot.png)

```sh
~ > qordle --hits ar --misses binstpel --pattern ..ar. | jq
[
  "acara",
  "afara",
  "arara",
  "award",
  "chard",
  "chark",
  "charm",
  "charr",
  "chary",
  "dwarf",
  "guara",
  "guard",
  "hoard",
  "hoary",
  "orary",
  "ovary",
  "quark",
  "uzara",
  "wharf"
]
```

```sh
~ > qordle --hits ar --misses binstpel --pattern ".[acdfghjklmoquvwxyz]ar." | jq
[
  "acara",
  "afara",
  "award",
  "chard",
  "chark",
  "charm",
  "charr",
  "chary",
  "dwarf",
  "guara",
  "guard",
  "hoard",
  "hoary",
  "ovary",
  "quark",
  "uzara",
  "wharf"
]
```
