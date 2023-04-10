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
