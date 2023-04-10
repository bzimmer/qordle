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
