# Commands

## Input Patterns

Many of the commands take an input pattern as feedback from guessing a word.

* **Miss**      &rarr; a lower case letter (eg "local")
* **Misplaced** &rarr; a lower case letter proceeded by a '.' (eg "l.ocal")
* **Exact**     &rarr; an upper case letter (eg "LOcal")


## Global Flags
|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|debug|||enable debug log level|
|monochrome|||disable color output|
|help|h||show help|

## Commands
* [bee](#bee)
* [bee today](#bee-today)
* [help](#help)
* [letterboxed](#letterboxed)
* [order](#order)
* [play](#play)
* [ranks](#ranks)
* [score](#score)
* [strategies](#strategies)
* [suggest](#suggest)
* [validate](#validate)
* [version](#version)
* [wordlists](#wordlists)

### *bee*

**Description**

Download the NYT spelling bee




### *bee today*

**Description**





**Syntax**

```sh
$ qordle bee today [flags]
```



### *help*

**Description**

Shows a list of commands or help for one command



**Syntax**

```sh
$ qordle help [flags] [command]
```



### *letterboxed*

**Description**

Solve the NYT Letter Boxed puzzle



**Syntax**

```sh
$ qordle letterboxed [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|min|||minimum word size|
|max|||maximum solution length|
|concurrent|||number of cpus to use for concurrent solving|
|wordlist|w||use the specified embedded word list|


### *order*

**Description**

Order the arguments per the strategy



**Syntax**

```sh
$ qordle order [flags] word [, word, ...]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|strategy|s||use the specified strategy|
|speculate|S||speculate if necessary|


### *play*

**Description**

Play wordle automatically



**Syntax**

```sh
$ qordle play [flags]
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|start|t|||
|progress|B||display a progress bar|
|rounds|r||max rounds not to exceed `rounds` * len(secret)|
|wordlist|w||use the specified embedded word list|
|strategy|s||use the specified strategy|
|speculate|S||speculate if necessary|


### *ranks*

**Description**

Detailed rank information from letter frequency tables



**Syntax**

```sh
$ qordle ranks [flags] <word> ...
```


**Example**

Sum all the percentages for letters in position 2

``` shell
$ qordle ranks | jq '.positions | flatten | map(."2") | add'
1
```

Compute the score for the words

``` shell
$ qordle ranks brown | jq .words
{
    "brown": {
        "bigrams": {
        "ranks": {
            "br": 0.0022,
            "ow": 0.0014,
            "ro": 0.0112,
            "wn": 0.0002
        },
        "total": 0.0750
        },
        "frequencies": {
        "ranks": {
            "0": 0.0183,
            "1": 0.0704,
            "2": 0.072,
            "3": 0.0065,
            "4": 0.0718
        },
        "total": 0.2390
        },
        "positions": {
        "ranks": {
            "0": 0.07,
            "1": 0.0685,
            "2": 0.0637,
            "3": 0.0133,
            "4": 0.0581
        },
        "total": 0.2736
        }
    }
}
```


### *score*

**Description**

Score the guesses against the secret



**Syntax**

```sh
$ qordle score [flags] <secret> <guess> [, <guess>]
```



### *strategies*

**Description**

List all available strategies



**Syntax**

```sh
$ qordle strategies [flags]
```



### *suggest*

**Description**

Suggest the next word to guess incorporating the already scored patterns



**Syntax**

```sh
$ qordle suggest [flags] <pattern>...
```


**Flags**

|Name|Aliases|EnvVars|Description|
|-|-|-|-|
|length|||word length|
|wordlist|w||use the specified embedded word list|
|strategy|s||use the specified strategy|
|speculate|S||speculate if necessary|

**Example**

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


### *validate*

**Description**

Validate the word against the pattern



**Syntax**

```sh
$ qordle validate [flags] <guess> <secret>...
```


**Example**

This command is useful for understanding why a guess was rejected against the secret. The
`--debug` flag is your friend as it will show the first reason for a word to be rejected.

```shell
$ qordle --debug validate brown local
2023-04-10T08:01:42+02:00 DBG compile pattern=[^loca][^loca][^loca][^loca][^loca] required={}
2023-04-10T08:01:42+02:00 DBG filter found=o i=2 reason=invalid word=brown
{
	"guess": "brown",
	"ok": false,
	"secrets": [
	  "local"
	]
}
```


### *version*

**Description**

Show the version information of the binary



**Syntax**

```sh
$ qordle version [flags]
```



### *wordlists*

**Description**

List all available wordlists



**Syntax**

```sh
$ qordle wordlists [flags]
```


**Example**

``` shell title="List all available wordlists"
$ qordle wordlists | jq
[
  "possible",
  "qordle",
  "solutions"
]
```

