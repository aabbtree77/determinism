// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package starlark_test

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aabbtree77/determinism"
	"github.com/aabbtree77/determinism/internal/chunkedfile"
	"github.com/aabbtree77/determinism/resolve"
	"github.com/aabbtree77/determinism/starlarktest"
	"github.com/aabbtree77/determinism/syntax"
)

func init() {
	// The tests make extensive use of these not-yet-standard features.
	resolve.AllowLambda = true
	resolve.AllowNestedDef = true
	resolve.AllowFloat = true
	resolve.AllowSet = true
}

func TestEvalExpr(t *testing.T) {
	// This is mostly redundant with the new *.star tests.
	// TODO(adonovan): move checks into *.star files and
	// reduce this to a mere unit test of starlark.Eval.
	thread := new(starlark.Thread)
	for _, test := range []struct{ src, want string }{
		{`123`, `123`},
		{`-1`, `-1`},
		{`"a"+"b"`, `"ab"`},
		{`1+2`, `3`},

		// lists
		{`[]`, `[]`},
		{`[1]`, `[1]`},
		{`[1,]`, `[1]`},
		{`[1, 2]`, `[1, 2]`},
		{`[2 * x for x in [1, 2, 3]]`, `[2, 4, 6]`},
		{`[2 * x for x in [1, 2, 3] if x > 1]`, `[4, 6]`},
		{`[(x, y) for x in [1, 2] for y in [3, 4]]`,
			`[(1, 3), (1, 4), (2, 3), (2, 4)]`},
		{`[(x, y) for x in [1, 2] if x == 2 for y in [3, 4]]`,
			`[(2, 3), (2, 4)]`},
		// tuples
		{`()`, `()`},
		{`(1)`, `1`},
		{`(1,)`, `(1,)`},
		{`(1, 2)`, `(1, 2)`},
		{`(1, 2, 3, 4, 5)`, `(1, 2, 3, 4, 5)`},
		// dicts
		{`{}`, `{}`},
		{`{"a": 1}`, `{"a": 1}`},
		{`{"a": 1,}`, `{"a": 1}`},

		// conditional
		{`1 if 3 > 2 else 0`, `1`},
		{`1 if "foo" else 0`, `1`},
		{`1 if "" else 0`, `0`},

		// indexing
		{`["a", "b"][0]`, `"a"`},
		{`["a", "b"][1]`, `"b"`},
		{`("a", "b")[0]`, `"a"`},
		{`("a", "b")[1]`, `"b"`},
		{`"aΩb"[0]`, `"a"`},
		{`"aΩb"[1]`, `"\xce"`},
		{`"aΩb"[3]`, `"b"`},
		{`{"a": 1}["a"]`, `1`},
		{`{"a": 1}["b"]`, `key "b" not in dict`},
		{`{}[[]]`, `unhashable type: list`},
		{`{"a": 1}[[]]`, `unhashable type: list`},
		{`[x for x in range(3)]`, "[0, 1, 2]"},
	} {
		var got string
		if v, err := starlark.Eval(thread, "<expr>", test.src, nil); err != nil {
			got = err.Error()
		} else {
			got = v.String()
		}
		if got != test.want {
			t.Errorf("eval %s = %s, want %s", test.src, got, test.want)
		}
	}
}

func TestExecFile(t *testing.T) {
	testdata := starlarktest.DataFile(".", ".")
	//fmt.Printf("Inspecting testdata, %s!\n", testdata)
	thread := &starlark.Thread{Load: load}
	starlarktest.SetReporter(thread, t)
	for _, file := range []string{
		"testdata/assign.star",
		"testdata/bool.star",
		"testdata/builtins.star",
		"testdata/control.star",
		"testdata/dict.star",
		"testdata/float.star",
		"testdata/function.star",
		"testdata/int.star",
		"testdata/list.star",
		"testdata/misc.star",
		"testdata/set.star",
		"testdata/string.star",
		"testdata/tuple.star",
	} {
		filename := filepath.Join(testdata, file)
		for _, chunk := range chunkedfile.Read(filename, t) {
			predeclared := starlark.StringDict{
				"hasfields": starlark.NewBuiltin("hasfields", newHasFields),
				"fibonacci": fib{},
			}
			_, err := starlark.ExecFile(thread, filename, chunk.Source, predeclared)
			switch err := err.(type) {
			case *starlark.EvalError:
				found := false
				for _, fr := range err.Stack() {
					posn := fr.Position()
					if posn.Filename() == filename {
						chunk.GotError(int(posn.Line), err.Error())
						found = true
						break
					}
				}
				if !found {
					t.Error(err.Backtrace())
				}
			case nil:
				// success
			default:
				t.Error(err)
			}
			chunk.Done()
		}
	}
}

// A fib is an iterable value representing the infinite Fibonacci sequence.
type fib struct{}

func (t fib) Freeze()                   {}
func (t fib) String() string            { return "fib" }
func (t fib) Type() string              { return "fib" }
func (t fib) Truth() starlark.Bool       { return true }
func (t fib) Hash() (uint32, error)     { return 0, fmt.Errorf("fib is unhashable") }
func (t fib) Iterate() starlark.Iterator { return &fibIterator{0, 1} }

type fibIterator struct{ x, y int }

func (it *fibIterator) Next(p *starlark.Value) bool {
	*p = starlark.MakeInt(it.x)
	it.x, it.y = it.y, it.x+it.y
	return true
}
func (it *fibIterator) Done() {}

// load implements the 'load' operation as used in the evaluator tests.
func load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	if module == "assert.star" {
		return starlarktest.LoadAssertModule(".")
	}

	// TODO(adonovan): test load() using this execution path.
	filename := filepath.Join(filepath.Dir(thread.Caller().Position().Filename()), module)
	return starlark.ExecFile(thread, filename, nil, nil)
}

func newHasFields(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return &hasfields{attrs: make(map[string]starlark.Value)}, nil
}

// hasfields is a test-only implementation of HasAttrs.
// It permits any field to be set.
// Clients will likely want to provide their own implementation,
// so we don't have any public implementation.
type hasfields struct {
	attrs  starlark.StringDict
	frozen bool
}

var (
	_ starlark.HasAttrs  = (*hasfields)(nil)
	_ starlark.HasBinary = (*hasfields)(nil)
)

func (hf *hasfields) String() string        { return "hasfields" }
func (hf *hasfields) Type() string          { return "hasfields" }
func (hf *hasfields) Truth() starlark.Bool   { return true }
func (hf *hasfields) Hash() (uint32, error) { return 42, nil }

func (hf *hasfields) Freeze() {
	if !hf.frozen {
		hf.frozen = true
		for _, v := range hf.attrs {
			v.Freeze()
		}
	}
}

func (hf *hasfields) Attr(name string) (starlark.Value, error) { return hf.attrs[name], nil }

func (hf *hasfields) SetField(name string, val starlark.Value) error {
	if hf.frozen {
		return fmt.Errorf("cannot set field on a frozen hasfields")
	}
	hf.attrs[name] = val
	return nil
}

func (hf *hasfields) AttrNames() []string {
	names := make([]string, 0, len(hf.attrs))
	for key := range hf.attrs {
		names = append(names, key)
	}
	return names
}

func (hf *hasfields) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
	// This method exists so we can exercise 'list += x'
	// where x is not Iterable but defines list+x.
	if op == syntax.PLUS {
		if _, ok := y.(*starlark.List); ok {
			return starlark.MakeInt(42), nil // list+hasfields is 42
		}
	}
	return nil, nil
}

func TestParameterPassing(t *testing.T) {
	const filename = "parameters.go"
	const src = `
def a():
	return
def b(a, b):
	return a, b
def c(a, b=42):
	return a, b
def d(*args):
	return args
def e(**kwargs):
	return kwargs
def f(a, b=42, *args, **kwargs):
	return a, b, args, kwargs
`

	thread := new(starlark.Thread)
	globals, err := starlark.ExecFile(thread, filename, src, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct{ src, want string }{
		{`a()`, `None`},
		{`a(1)`, `function a takes no arguments (1 given)`},
		{`b()`, `function b takes exactly 2 arguments (0 given)`},
		{`b(1)`, `function b takes exactly 2 arguments (1 given)`},
		{`b(1, 2)`, `(1, 2)`},
		{`b`, `<function b>`}, // asserts that b's parameter b was treated as a local variable
		{`b(1, 2, 3)`, `function b takes exactly 2 arguments (3 given)`},
		{`b(1, b=2)`, `(1, 2)`},
		{`b(1, a=2)`, `function b got multiple values for keyword argument "a"`},
		{`b(1, x=2)`, `function b got an unexpected keyword argument "x"`},
		{`b(a=1, b=2)`, `(1, 2)`},
		{`b(b=1, a=2)`, `(2, 1)`},
		{`b(b=1, a=2, x=1)`, `function b got an unexpected keyword argument "x"`},
		{`b(x=1, b=1, a=2)`, `function b got an unexpected keyword argument "x"`},
		{`c()`, `function c takes at least 1 argument (0 given)`},
		{`c(1)`, `(1, 42)`},
		{`c(1, 2)`, `(1, 2)`},
		{`c(1, 2, 3)`, `function c takes at most 2 arguments (3 given)`},
		{`c(1, b=2)`, `(1, 2)`},
		{`c(1, a=2)`, `function c got multiple values for keyword argument "a"`},
		{`c(a=1, b=2)`, `(1, 2)`},
		{`c(b=1, a=2)`, `(2, 1)`},
		{`d()`, `()`},
		{`d(1)`, `(1,)`},
		{`d(1, 2)`, `(1, 2)`},
		{`d(1, 2, k=3)`, `function d got an unexpected keyword argument "k"`},
		{`d(args=[])`, `function d got an unexpected keyword argument "args"`},
		{`e()`, `{}`},
		{`e(1)`, `function e takes exactly 0 arguments (1 given)`},
		{`e(k=1)`, `{"k": 1}`},
		{`e(kwargs={})`, `{"kwargs": {}}`},
		{`f()`, `function f takes at least 1 argument (0 given)`},
		{`f(0)`, `(0, 42, (), {})`},
		{`f(0)`, `(0, 42, (), {})`},
		{`f(0, 1)`, `(0, 1, (), {})`},
		{`f(0, 1, 2)`, `(0, 1, (2,), {})`},
		{`f(0, 1, 2, 3)`, `(0, 1, (2, 3), {})`},
		{`f(a=0)`, `(0, 42, (), {})`},
		{`f(0, b=1)`, `(0, 1, (), {})`},
		{`f(0, a=1)`, `function f got multiple values for keyword argument "a"`},
		{`f(0, b=1, c=2)`, `(0, 1, (), {"c": 2})`},
	} {
		var got string
		if v, err := starlark.Eval(thread, "<expr>", test.src, globals); err != nil {
			got = err.Error()
		} else {
			got = v.String()
		}
		if got != test.want {
			t.Errorf("eval %s = %s, want %s", test.src, got, test.want)
		}
	}
}

// TestPrint ensures that the Starlark print function calls
// Thread.Print, if provided.
func TestPrint(t *testing.T) {
	const src = `
print("hello")
def f(): print("world")
f()
`
	buf := new(bytes.Buffer)
	print := func(thread *starlark.Thread, msg string) {
		caller := thread.Caller()
		name := "<module>"
		if caller.Function() != nil {
			name = caller.Function().Name()
		}
		fmt.Fprintf(buf, "%s: %s: %s\n", caller.Position(), name, msg)
	}
	thread := &starlark.Thread{Print: print}
	if _, err := starlark.ExecFile(thread, "foo.go", src, nil); err != nil {
		t.Fatal(err)
	}
	want := "foo.go:2:6: <module>: hello\n" +
		"foo.go:3:15: f: world\n"
	if got := buf.String(); got != want {
		t.Errorf("output was %s, want %s", got, want)
	}
}

func Benchmark(b *testing.B) {
	testdata := starlarktest.DataFile(".", ".")
	thread := new(starlark.Thread)
	for _, file := range []string{
		"testdata/benchmark.star",
		// ...
	} {
		filename := filepath.Join(testdata, file)

		// Evaluate the file once.
		globals, err := starlark.ExecFile(thread, filename, nil, nil)
		if err != nil {
			reportEvalError(b, err)
		}

		// Repeatedly call each global function named bench_* as a benchmark.
		for name, value := range globals {
			if fn, ok := value.(*starlark.Function); ok && strings.HasPrefix(name, "bench_") {
				b.Run(name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_, err := starlark.Call(thread, fn, nil, nil)
						if err != nil {
							reportEvalError(b, err)
						}
					}
				})
			}
		}
	}
}

func reportEvalError(tb testing.TB, err error) {
	if err, ok := err.(*starlark.EvalError); ok {
		tb.Fatal(err.Backtrace())
	}
	tb.Fatal(err)
}

// TestInt exercises the Int.Int64 and Int.Uint64 methods.
// If we can move their logic into math/big, delete this test.
func TestInt(t *testing.T) {
	one := starlark.MakeInt(1)

	for _, test := range []struct {
		i          starlark.Int
		wantInt64  string
		wantUint64 string
	}{
		{starlark.MakeInt64(math.MinInt64).Sub(one), "error", "error"},
		{starlark.MakeInt64(math.MinInt64), "-9223372036854775808", "error"},
		{starlark.MakeInt64(-1), "-1", "error"},
		{starlark.MakeInt64(0), "0", "0"},
		{starlark.MakeInt64(1), "1", "1"},
		{starlark.MakeInt64(math.MaxInt64), "9223372036854775807", "9223372036854775807"},
		{starlark.MakeUint64(math.MaxUint64), "error", "18446744073709551615"},
		{starlark.MakeUint64(math.MaxUint64).Add(one), "error", "error"},
	} {
		gotInt64, gotUint64 := "error", "error"
		if i, ok := test.i.Int64(); ok {
			gotInt64 = fmt.Sprint(i)
		}
		if u, ok := test.i.Uint64(); ok {
			gotUint64 = fmt.Sprint(u)
		}
		if gotInt64 != test.wantInt64 {
			t.Errorf("(%s).Int64() = %s, want %s", test.i, gotInt64, test.wantInt64)
		}
		if gotUint64 != test.wantUint64 {
			t.Errorf("(%s).Uint64() = %s, want %s", test.i, gotUint64, test.wantUint64)
		}
	}
}

func TestBacktrace(t *testing.T) {
	// This test ensures continuity of the stack of active Starlark
	// functions, including propagation through built-ins such as 'min'
	// (though min does not itself appear in the stack).
	const src = `
def f(x): return 1//x
def g(x): f(x)
def h(): return min([1, 2, 0], key=g)
def i(): return h()
i()
`
	thread := new(starlark.Thread)
	_, err := starlark.ExecFile(thread, "crash.go", src, nil)
	switch err := err.(type) {
	case *starlark.EvalError:
		got := err.Backtrace()
		const want = `Traceback (most recent call last):
  crash.go:6:2: in <toplevel>
  crash.go:5:18: in i
  crash.go:4:20: in h
  crash.go:3:12: in g
  crash.go:2:19: in f
Error: floored division by zero`
		if got != want {
			t.Errorf("error was %s, want %s", got, want)
		}
	case nil:
		t.Error("ExecFile succeeded unexpectedly")
	default:
		t.Errorf("ExecFile failed with %v, wanted *EvalError", err)
	}
}

// TestRepeatedExec parses and resolves a file syntax tree once then
// executes it repeatedly with different values of its predeclared variables.
func TestRepeatedExec(t *testing.T) {
	f, err := syntax.Parse("repeat.star", "y = 2 * x", 0)
	if err != nil {
		t.Fatal(f) // parse error
	}

	isPredeclared := func(name string) bool { return name == "x" } // x, but not y

	if err := resolve.File(f, isPredeclared, starlark.Universe.Has); err != nil {
		t.Fatal(err) // resolve error
	}

	const yIndex = 0
	if yName := f.Globals[yIndex].Name; yName != "y" {
		t.Fatalf("global[%d] = %s, want y", yIndex, yName)
	}

	thread := new(starlark.Thread)
	for _, test := range []struct {
		x, want starlark.Value
	}{
		{x: starlark.MakeInt(42), want: starlark.MakeInt(84)},
		{x: starlark.String("mur"), want: starlark.String("murmur")},
		{x: starlark.Tuple{starlark.None}, want: starlark.Tuple{starlark.None, starlark.None}},
	} {
		predeclared := starlark.StringDict{"x": test.x}
		globals := make([]starlark.Value, len(f.Globals))
		fr := thread.Push(predeclared, globals, len(f.Locals))
		if err := fr.ExecStmts(f.Stmts); err != nil {
			t.Errorf("x=%v: %v", test.x, err) // exec error
		} else if eq, err := starlark.Equal(globals[yIndex], test.want); err != nil {
			t.Errorf("x=%v: %v", test.x, err) // comparison error
		} else if !eq {
			t.Errorf("x=%v: got y=%v, want %v", test.x, globals[yIndex], test.want)
		}
		thread.Pop()
	}
}

// TestUnpackUserDefined tests that user-defined
// implementations of starlark.Value may be unpacked.
func TestUnpackUserDefined(t *testing.T) {
	// success
	want := new(hasfields)
	var x *hasfields
	if err := starlark.UnpackArgs("unpack", starlark.Tuple{want}, nil, "x", &x); err != nil {
		t.Errorf("UnpackArgs failed: %v", err)
	}
	if x != want {
		t.Errorf("for x, got %v, want %v", x, want)
	}

	// failure
	err := starlark.UnpackArgs("unpack", starlark.Tuple{starlark.MakeInt(42)}, nil, "x", &x)
	if want := "unpack: for parameter 1: got int, want hasfields"; fmt.Sprint(err) != want {
		t.Errorf("unpack args error = %q, want %q", err, want)
	}
}
