# gravedigger
finds unused gocode

# Example:

Running this in the git directory (inside the standard `GOPATH`)

```
> gravedigger test
---------step 0-------------(find all code in directory)

github.com/tscholl2/gravedigger/test/sub
github.com/tscholl2/gravedigger/test/sub/sub
github.com/tscholl2/gravedigger/test

---------step 1-------------(mark all declarations)

github.com/tscholl2/gravedigger/test:
	- b
	- c
	- foo
	- bar
github.com/tscholl2/gravedigger/test/sub:
	- A
github.com/tscholl2/gravedigger/test/sub/sub:
	- C

---------step 2-------------(mark all used declarations)

github.com/tscholl2/gravedigger/test/sub:A --- used
github.com/tscholl2/gravedigger/test/sub/sub:C --- used
github.com/tscholl2/gravedigger/test:foo --- used

---------step 3-------------(list all unused declarations)

github.com/tscholl2/gravedigger/test:
	- b ---> main.go:test/main.go:8:5
	- c ---> main.go:test/main.go:9:5
	- bar ---> main.go:test/main.go:14:6
```
