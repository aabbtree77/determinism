// The repl package provides a read/eval/print loop for Starlark.
//
// It supports readline-style command editing,
// and interrupts through Control-C.
//
// If an input line can be parsed as an expression,
// the REPL parses and evaluates it and prints its result.
// Otherwise the REPL reads lines until a blank line,
// then tries again to parse the multi-line input as an
// expression. If the input still cannot be parsed as an expression,
// the REPL parses and executes it as a file (a list of statements),
// for side effects.
package repl

// TODO(adonovan):
//
// - Unparenthesized tuples are not parsed as a single expression:
//     >>> (1, 2)
//     (1, 2)
//     >>> 1, 2
//     ...
//     >>>
//   This is not necessarily a bug.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/chzyer/readline"
	"github.com/aabbtree77/determinism"
	"github.com/aabbtree77/determinism/resolve"
	"github.com/aabbtree77/determinism/syntax"
)

var interrupted = make(chan os.Signal, 1)

// REPL executes a read, eval, print loop.
//
// Before evaluating each expression, it sets the Starlark thread local
// variable named "context" to a context.Context that is cancelled by a
// SIGINT (Control-C). Client-supplied global functions may use this
// context to make long-running operations interruptable.
//
func REPL(thread *starlark.Thread, globals starlark.StringDict) {
	signal.Notify(interrupted, os.Interrupt)
	defer signal.Stop(interrupted)

	rl, err := readline.New(">>> ")
	if err != nil {
		PrintError(err)
		return
	}
	defer rl.Close()
	for {
		if err := rep(rl, thread, globals); err != nil {
			if err == readline.ErrInterrupt {
				fmt.Println(err)
				continue
			}
			break
		}
	}
	fmt.Println()
}

// rep reads, evaluates, and prints one item.
//
// It returns an error (possibly readline.ErrInterrupt)
// only if readline failed. Starlark errors are printed.
func rep(rl *readline.Instance, thread *starlark.Thread, globals starlark.StringDict) error {
	// Each item gets its own context,
	// which is cancelled by a SIGINT.
	//
	// Note: during Readline calls, Control-C causes Readline to return
	// ErrInterrupt but does not generate a SIGINT.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-interrupted:
			cancel()
		case <-ctx.Done():
		}
	}()

	thread.SetLocal("context", ctx)

	rl.SetPrompt(">>> ")
	line, err := rl.Readline()
	if err != nil {
		return err // may be ErrInterrupt
	}

	if l := strings.TrimSpace(line); l == "" || l[0] == '#' {
		return nil // blank or comment
	}

	// If the line contains a well-formed expression, evaluate it.
	if _, err := syntax.ParseExpr("<stdin>", line, 0); err == nil {
		if v, err := starlark.Eval(thread, "<stdin>", line, globals); err != nil {
			PrintError(err)
		} else if v != starlark.None {
			fmt.Println(v)
		}
		return nil
	}

	// If the input so far is a single load or assignment statement,
	// execute it without waiting for a blank line.
	if f, err := syntax.Parse("<stdin>", line, 0); err == nil && len(f.Stmts) == 1 {
		switch f.Stmts[0].(type) {
		case *syntax.AssignStmt, *syntax.LoadStmt:
			// Execute it as a file.
			if err := execFileNoFreeze(thread, line, globals); err != nil {
				PrintError(err)
			}
			return nil
		}
	}

	// Otherwise assume it is the first of several
	// comprising a file, followed by a blank line.
	var buf bytes.Buffer
	fmt.Fprintln(&buf, line)
	for {
		rl.SetPrompt("... ")
		line, err := rl.Readline()
		if err != nil {
			return err // may be ErrInterrupt
		}
		if l := strings.TrimSpace(line); l == "" {
			break // blank
		}
		fmt.Fprintln(&buf, line)
	}
	text := buf.Bytes()

	// Try parsing it once more as an expression,
	// such as a call spread over several lines:
	//   f(
	//     1,
	//     2
	//   )
	if _, err := syntax.ParseExpr("<stdin>", text, 0); err == nil {
		if v, err := starlark.Eval(thread, "<stdin>", text, globals); err != nil {
			PrintError(err)
		} else if v != starlark.None {
			fmt.Println(v)
		}
		return nil
	}

	// Execute it as a file.
	if err := execFileNoFreeze(thread, text, globals); err != nil {
		PrintError(err)
	}

	return nil
}

// execFileNoFreeze is starlark.ExecFile without globals.Freeze().
func execFileNoFreeze(thread *starlark.Thread, src interface{}, globals starlark.StringDict) error {
	// parse
	f, err := syntax.Parse("<stdin>", src, 0)
	if err != nil {
		return err
	}

	// resolve
	if err := resolve.File(f, globals.Has, starlark.Universe.Has); err != nil {
		return err
	}

	// execute
	// The global names from the previous call become the predeclared names of this call.
	globalsArray := make([]starlark.Value, len(f.Globals))
	fr := thread.Push(globals, globalsArray, len(f.Locals))
	err = fr.ExecStmts(f.Stmts)
	thread.Pop()

	// Copy globals back to the caller's map.
	// If execution failed, some globals may be undefined.
	for i, id := range f.Globals {
		if v := globalsArray[i]; v != nil {
			globals[id.Name] = v
		}
	}

	return err
}

// PrintError prints the error to stderr,
// or its backtrace if it is a Starlark evaluation error.
func PrintError(err error) {
	if evalErr, ok := err.(*starlark.EvalError); ok {
		fmt.Fprintln(os.Stderr, evalErr.Backtrace())
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}

// MakeLoad returns a simple sequential implementation of module loading
// suitable for use in the REPL.
// Each function returned by MakeLoad accesses a distinct private cache.
func MakeLoad() func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	type entry struct {
		globals starlark.StringDict
		err     error
	}

	var cache = make(map[string]*entry)

	return func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
		e, ok := cache[module]
		if e == nil {
			if ok {
				// request for package whose loading is in progress
				return nil, fmt.Errorf("cycle in load graph")
			}

			// Add a placeholder to indicate "load in progress".
			cache[module] = nil

			// Load it.
			thread := &starlark.Thread{Load: thread.Load}
			globals, err := starlark.ExecFile(thread, module, nil, nil)
			e = &entry{globals, err}

			// Update the cache.
			cache[module] = e
		}
		return e.globals, e.err
	}
}
