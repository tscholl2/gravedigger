package main

import (
	"github.com/tscholl2/gravedigger/test/sub"
	SUB "github.com/tscholl2/gravedigger/test/sub/sub"
)

var b = sub.A
var c = SUB.C

func foo() {
	x := SUB.X{}
	if x.AB == 3 {
		panic("no three")
	}
}

func bar() {
	foo()
}
