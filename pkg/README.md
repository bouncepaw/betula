# pkg
`pkg` holds packages that can be thought of as ‘libraries’.
They don't have any dependencies on Betula's domain,
and could generally be used in other projects easily.

Note that sometimes logic regarding some format is split into multiple packages.
In `pkg`, you would have an implementation of it,
and in a service you would have some logic on top of it.

Sometimes we would vendor packages here,
like we did with Ted Unangst's `rss` and `httpsig` after his disappearance.

Some packages here start with `bx`. That means ‘Betula's extended’.
