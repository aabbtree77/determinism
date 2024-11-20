## A Tree-Walking Interpreter

This is a fork of the March 30 2018 commit [#96](https://github.com/google/starlark-go/pull/96) of [Starlark-Go](https://github.com/google/starlark-go): 

```
git checkout -b tree-walker 0d5491befad9f6af126fdcb886dc58a8f059bea7
```

It is the last commit which contains a tree-walking interpreter of [Starlark](https://github.com/google/starlark-go/blob/master/doc/spec.md). Afterwards [Alan Donovan](https://www.youtube.com/watch?v=9P_YKVhncWI) et al. moved on to a bytecode compiler.

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

Minor code restoration: Added `go.mod`, fixed the paths for the tests not passing due to $GOPATH in a few places, took care of [#32479](https://github.com/golang/go/issues/32479), replaced "skylark" with "starlark" in code, but not in the docs as they contain links. The code may also contain links...

## Remarks

- 10 KLOC of Go with some beef towards immutable modules, caching, parallelism.

- eval_test.go gives hints about module loading with cycle detection.

- hashtable.go does a hash/map/dict in less than 350 lines of code.

- The books by [Thorsten Ball](https://thorstenball.com/books/) are great, but Starlark-Go is the next level.

- The commit history reveals the edge cases emerging in a simple language with a spec.

- In [Clace](https://clace.io/) Starlark-Go serves as configuration and the HTML generation. It is unclear to me if this is not yet another broken web framework or cloud service. Go might not be a good choice to do web apps, but this is another story.

- The Nobel Prize in Physics 2024 goes to [Alan Donovan](https://www.youtube.com/watch?v=9P_YKVhncWI) et al...
