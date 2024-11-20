## Determinism

This is a fork of the March 30 2018 commit [#96](https://github.com/google/starlark-go/pull/96) of [Starlark-Go](https://github.com/google/starlark-go). It is the last commit which contains a tree-walking interpreter of [Starlark](https://github.com/google/starlark-go/blob/master/doc/spec.md). Afterwards [Alan Donovan](https://www.youtube.com/watch?v=9P_YKVhncWI) et al. moved on to a bytecode compiler.

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

Minor code fixes: Added `go.mod`, fixed the paths for the tests not passing due to $GOPATH in a few places, took care of [#32479](https://github.com/golang/go/issues/32479), replaced "skylark" with "starlark".

## To Do

I would further simplify this code by removing init(), the special "_test.go" files with TestXYX and ExampleXYZ functions that run when invoking "go test -v" (a pain to debug as they are not true "*.go" files), the package "internal", text/code chunking...

## A Note About Go Modules

Any Go code with "mod.go" is a published library (!) as the Go module names must be real github links to the code.

Starlark-Go goes even further with the so called "vanity import paths in Go". Their module name is "go.starlark.net" which is a domain name for the web server which serves a special HTML with meta tags to Go tools such as `go get` and `go install`. These tools get redirected to their github repo. In order to have a nice module name in Go, one needs to buy a domain. 

## Further Remarks

- 10 KLOC of Go with some real beef: Immutable modules, caching, parallelism.

- eval_test.go hints on how to implement module loading with cycle detection.

- hashtable.go does a hash/map/dict in less than 350 lines of code.

- The books by [Thorsten Ball](https://thorstenball.com/books/) serve as a great introduction to compilers with a lot of following on github. Starlark-Go is the next level.

- Go is highly praised for its "go build" and minimal polymorphism (at least prior to Go v1.18). "go build" is a joy compared to multiple TypeScript runtimes, Make, CMake, or, god forbid, [Bazel](https://www.reddit.com/r/devops/comments/1c2g3s4/bazel_is_ruining_my_life/).

- See also [Clace](https://clace.io/) where Starlark serves as a configuration and html-templating language.

- The Nobel Prize in Physics 2024 goes to [Alan Donovan](https://www.youtube.com/watch?v=9P_YKVhncWI) et al...
