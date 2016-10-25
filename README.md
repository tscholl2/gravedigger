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

main.go
test/main.go
test/sub/sub.go
test/sub/sub/sub.go

---------step 1-------------(mark all declarations)

* unmarkGuru = main.go:206:6
* def = main.go:240:9
* c = test/main.go:9:5
* bar = test/main.go:18:6
* A = test/sub/sub.go:3:5
* dir = main.go:23:2
* declaration = main.go:27:6
* mark = main.go:119:6
* foo = test/main.go:11:6
* unmarkFast = main.go:172:6
* X = test/sub/sub/sub.go:5:6
* useGuru = main.go:24:2
* parse = main.go:99:6
* b = test/main.go:8:5
* C = test/sub/sub/sub.go:3:5

---------step 2-------------(mark all used declarations)

useGuru
used in:  main.go:33:16
defd in:  main.go:24:2
dir
used in:  main.go:47:3
defd in:  main.go:23:2
declaration
used in:  main.go:57:34
defd in:  main.go:27:6
parse
used in:  main.go:61:2
defd in:  main.go:99:6
mark
used in:  main.go:73:2
defd in:  main.go:119:6
unmarkGuru
used in:  main.go:84:3
defd in:  main.go:206:6
unmarkFast
used in:  main.go:86:3
defd in:  main.go:172:6
def
used in:  main.go:241:36
defd in:  main.go:240:9
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
* bar = test/main.go:18:6
* c = test/main.go:9:5
```
