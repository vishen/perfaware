# JSON Parser

Simple JSON parser in Go.

```
$ cat test.json
[
	{
		"name": "some name",
		"points": [
			{"x0": 123.4, "y0": 987.6, "x1": 100000, "y1": 9999999999},
			{"x0": 0.4, "y0": 9.6, "x1": 100, "y1": 99999}
		]
	}
]

$ json_parser
{TOKEN_CHAR '['}
{TOKEN_CHAR '{'}
{TOKEN_KEY name}
{TOKEN_CHAR ':'}
{TOKEN_KEY some name}
{TOKEN_CHAR ','}
{TOKEN_KEY points}
{TOKEN_CHAR ':'}
{TOKEN_CHAR '['}
{TOKEN_CHAR '{'}
{TOKEN_KEY x0}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 123.4}
{TOKEN_KEY y0}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 987.6}
{TOKEN_KEY x1}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 100000}
{TOKEN_KEY y1}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 9999999999}
{TOKEN_CHAR ','}
{TOKEN_CHAR '{'}
{TOKEN_KEY x0}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 0.4}
{TOKEN_KEY y0}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 9.6}
{TOKEN_KEY x1}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 100}
{TOKEN_KEY y1}
{TOKEN_CHAR ':'}
{TOKEN_VALUE 99999}
{TOKEN_CHAR ']'}
{TOKEN_CHAR '}'}
{TOKEN_CHAR ']'}
```
