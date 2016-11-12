# gravedigger
finds unused gocode

inspired by [deadcode](https://github.com/remyoudompheng/go-misc/blob/master/deadcode/deadcode.go)

but not as good as [unused](https://godoc.org/honnef.co/go/unused)

This is a little different in that it only works on whole projects, and includes exported functions.
It is for projects rather than libraries. If something is exported by a subdirectory and not used
in the current directory (or other subdirectories) than it will be listed.

# Docs:

```
> gravedigger --help
gravedigger: looks for unused code in a directory. This differs
from other packages in that it takes a directory and lists things that are unused
any where in that directory, including exported things in subpackages/subdirectories.
Example: 'gravedigger'
Example: 'gravedigger .'
Example: 'gravedigger test/'
Options: 

  -guru
    	use guru for name referencing (SLOW)
```

# Example:

Running this in the git directory (inside the standard `GOPATH`)

```
> gravedigger test/
---------step 0-------------(find all code in directories)

test/main.go
test/sub/sub.go
test/sub/sub/sub.go

---------step 1-------------(mark all declarations)

* c = test/main.go:9:5
* foo = test/main.go:11:6
* bar = test/main.go:18:6
* A = test/sub/sub.go:3:5
* C = test/sub/sub/sub.go:3:5
* X = test/sub/sub/sub.go:5:6
* b = test/main.go:8:5

---------step 2-------------(mark all used declarations)

A
used in:  test/main.go:8:13
defd in:  test/sub/sub.go:3:5
C
used in:  test/main.go:9:13
defd in:  test/sub/sub/sub.go:3:5
X
used in:  test/main.go:12:11
defd in:  test/sub/sub/sub.go:5:6
foo
used in:  test/main.go:19:2
defd in:  test/main.go:11:6

---------step 3-------------(list all unused declarations)

* b = test/main.go:8:5
* c = test/main.go:9:5
* bar = test/main.go:18:6
```
