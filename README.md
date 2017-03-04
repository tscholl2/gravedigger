# gravedigger

finds unused gocode

inspired by [deadcode](https://github.com/remyoudompheng/go-misc/blob/master/deadcode/deadcode.go)

but not as good as [unused](https://godoc.org/honnef.co/go/unused)

This is a little different in that it only works on whole projects, and includes exported functions.
It is for projects rather than libraries. If something is exported by a subdirectory and not used
in the current directory (or other subdirectories) than it will be listed.

It should not have many false positives, but some exported functions which are used by other
libraries (e.g. `UnmarshalText`) may get listed.

It probably won't have too many false negatives unless you name all your variables the same thing.

# Docs

```
> gravedigger --help
gravedigger [directory]: looks for unused code in a directory. This differs
from other packages in that it takes a directory and lists things that are unused
any where in that directory, including exported things in subpackages/subdirectories.
Example: 'gravedigger'
Example: 'gravedigger .'
Example: 'gravedigger test/'
```

# Example

Running this in on the `test/` directory gives
```
> gravedigger test/
test/sub/sub/sub.go
	- a:6:2
test/main.go
	- b:8:5
	- c:9:5
	- bar:18:6
```
