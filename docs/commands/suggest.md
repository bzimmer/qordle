This command uses the specified strategies and any information about prior
guesses to sort the remaining words.

```shell
$ qordle suggest -s freq -s pos -S .rais.e to.nER | jq
[
  "neper",
  "nuder",
  "never",
  "ender",
  "under",
  "newer"
]
```

```shell
$ qordle suggest -s bigram .rais.e to.nER | jq
[
  "neper",
  "nuder",
  "never",
  "ender",
  "under",
  "newer"
]
```
