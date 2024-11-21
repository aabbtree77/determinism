## A Tree-Walking Interpreter

This is a fork of the March 30 2018 commit [#96](https://github.com/google/starlark-go/pull/96) of [Starlark-Go](https://github.com/google/starlark-go): 

```
git checkout -b tree-walker 0d5491befad9f6af126fdcb886dc58a8f059bea7
```

It is the latest commit which contains a tree-walking interpreter of [Starlark](https://github.com/google/starlark-go/blob/master/doc/spec.md). Afterwards [Alan Donovan](https://www.youtube.com/watch?v=9P_YKVhncWI) et al. moved on to a bytecode compiler-interpreter.

```
git clone https://github.com/aabbtree77/determinism.git
cd determinism
go build ./cmd/starlark
```

```
./starlark coins.star
By name:	dime, nickel, penny, quarter
By value:	penny, nickel, dime, quarter
```

```
go test -v
```

```
cd syntax
go test -v
```

**The idea here is to make the tree walker more visible for learning and exploration.**

Minor code restoration: Added `go.mod`, fixed the paths for the tests not passing due to $GOPATH in a few places, took care of [#32479](https://github.com/golang/go/issues/32479), replaced "skylark" with "starlark" in code, but not in the docs as they contain links.

## Remarks

- The books by [Thorsten Ball](https://thorstenball.com/books/) are great, but Starlark-Go is the next level, yet under 10 KLOC of Go.

- example_test.go states a demo of the concurrent cache with cycle detection for Starlark module loading in Go. More of that in Sect. 9.7 of the book by Alan A. A. Donovan and Brian W. Kernighan [gopl.io](https://www.gopl.io/).

- hashtable.go does a hash map from scratch in less than 350 lines of Go.

- For some real uses of [Starlark-Go](https://github.com/google/starlark-go) it could be worth checking out [Clace](https://clace.io/).

