# gravedigger
finds unused gocode

# Example:

Running this in the git directory (inside the standard `GOPATH`)

```
> gravedigger test
---------step 0-------------(find all code in directories)

test/main.go
test/sub/sub.go
test/sub/sub/sub.go

---------step 1-------------(mark all declarations)

* b = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:8:5
* foo = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:11:6
* AB = /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go:7:2
* a = /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go:6:2
* c = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:9:5
* bar = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:18:6
* A = /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub.go:3:5
* C = /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go:3:5
* X = /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go:5:6

---------step 2-------------(mark all used declarations)

A
used in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go 8 13
defd in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub.go 3 5
C
used in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go 9 13
defd in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go 3 5
X
used in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go 12 11
defd in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go 5 6
AB
used in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go 13 7
defd in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go 7 2
foo
used in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go 19 2
defd in:  /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go 11 6

---------step 3-------------(list all unused declarations)

* b = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:8:5
* c = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:9:5
* bar = /home/t/go/src/github.com/tscholl2/gravedigger/test/main.go:18:6
* a = /home/t/go/src/github.com/tscholl2/gravedigger/test/sub/sub/sub.go:6:2
```
