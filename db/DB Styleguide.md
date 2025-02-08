# Betula database styleguide
## SQL formatting
There is no chosen SQL code style yet. Just keep it close to what Bouncepaw might consider sane.

## `db` package
*All* database manipulations are encapsulated in the `db` package. Basically every possible way to use it gets a separate function. We don't have special types for the databases, the state is package-wide. Since we only have one database in use at any given moment, it's alright.

**UPD.** Special types are being incorporated now.

Be careful with migrations.

The functions should probably have the SQL code as inline strings. If you are to move them to constants, probably name them `q`. If you have multiple of them in one function, call them `qFoo` and `qBar`. Bouncepaw likes this q prefix, it's cute.

If you write the tests, that would be nice.