# Tests of Starlark 'function'

# TODO(adonovan):
# - add some introspection functions for looking at function values
#   and test that functions have correct position, free vars, names of locals, etc.
# - move the hard-coded tests of parameter passing from eval_test.go to here.

load("assert.star", "assert", "freeze")

# Test lexical scope and closures:
def outer(x):
   def inner(y):
     return x + x + y # multiple occurrences of x should create only 1 freevar
   return inner

z = outer(3)
assert.eq(z(5), 11)
assert.eq(z(7), 13)
z2 = outer(4)
assert.eq(z2(5), 13)
assert.eq(z2(7), 15)
assert.eq(z(5), 11)
assert.eq(z(7), 13)

# Function name
assert.eq(str(outer), '<function outer>')
assert.eq(str(z), '<function inner>')
assert.eq(str(str), '<built-in function str>')
assert.eq(str("".startswith), '<built-in method startswith of string value>')

# Stateful closure
def squares():
    x = [0]
    def f():
      x[0] += 1
      return x[0] * x[0]
    return f

sq = squares()
assert.eq(sq(), 1)
assert.eq(sq(), 4)
assert.eq(sq(), 9)
assert.eq(sq(), 16)

# Freezing a closure
sq2 = freeze(sq)
assert.fails(sq2, "frozen list")

# recursion detection, simple
def fib(x):
  if x < 2:
    return x
  return fib(x-2) + fib(x-1)
assert.fails(lambda: fib(10), "function fib called recursively")

# recursion detection, advanced
#
# A simplistic recursion check that looks for repeated calls to the
# same function value will not detect recursion using the Y
# combinator, which creates a new closure at each step of the
# recursion.  To truly prohibit recursion, the dynamic check must look
# for repeated calls of the same syntactic function body.
Y = lambda f: (lambda x: x(x))(lambda y: f(lambda *args: y(y)(*args)))
fibgen = lambda fib: lambda x: (x if x<2 else fib(x-1)+fib(x-2))
fib2 = Y(fibgen)
assert.fails(lambda: [fib2(x) for x in range(10)], "function lambda called recursively")

# call of function not through its name
# (regression test for parsing suffixes of primary expressions)
hf = hasfields()
hf.x = [len]
assert.eq(hf.x[0]("abc"), 3)
def f():
   return lambda: 1
assert.eq(f()(), 1)
assert.eq(["abc"][0][0].upper(), "A")

# functions may be recursively defined,
# so long as they don't dynamically recur.
calls = []
def yin(x):
  calls.append("yin")
  if x:
    yang(False)

def yang(x):
  calls.append("yang")
  if x:
    yin(False)

yin(True)
assert.eq(calls, ["yin", "yang"])

calls.clear()
yang(True)
assert.eq(calls, ["yang", "yin"])


# hash(builtin_function_or_method) should be deterministic.
closures = set(["".count for _ in range(10)])
assert.eq(len(closures), 10)
hashes = set([hash("".count) for _ in range(10)])
assert.eq(len(hashes), 1)

---
# Default values of function parameters are mutable.
load("assert.star", "assert", "freeze")

def f(x=[0]):
  return x

assert.eq(f(), [0])

f().append(1)
assert.eq(f(), [0, 1])

# Freezing a function value freezes its parameter defaults.
freeze(f)
assert.fails(lambda: f().append(2), "cannot append to frozen list")

---
# This is a well known corner case of parsing in Python.
load("assert.star", "assert")

f = lambda x: 1 if x else 0
assert.eq(f(True), 1)
assert.eq(f(False), 0)

x = True
f2 = (lambda x: 1) if x else 0
assert.eq(f2(123), 1)

tf = lambda: True, lambda: False
assert.true(tf[0]())
assert.true(not tf[1]())

---
# Missing parameters are correctly reported
# in functions of more than 64 parameters.
# (This tests a corner case of the implementation:
# we avoid a map allocation for <64 parameters)

load("assert.star", "assert")

def f(a, b, c, d, e, f, g, h,
      i, j, k, l, m, n, o, p,
      q, r, s, t, u, v, w, x,
      y, z, A, B, C, D, E, F,
      G, H, I, J, K, L, M, N,
      O, P, Q, R, S, T, U, V,
      W, X, Y, Z, aa, bb, cc, dd,
      ee, ff, gg, hh, ii, jj, kk, ll,
      mm):
  pass

assert.fails(lambda: f(
    1, 2, 3, 4, 5, 6, 7, 8,
    9, 10, 11, 12, 13, 14, 15, 16,
    17, 18, 19, 20, 21, 22, 23, 24,
    25, 26, 27, 28, 29, 30, 31, 32,
    33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48,
    49, 50, 51, 52, 53, 54, 55, 56,
    57, 58, 59, 60, 61, 62, 63, 64), "takes exactly 65 arguments .64 given.")

assert.fails(lambda: f(
    1, 2, 3, 4, 5, 6, 7, 8,
    9, 10, 11, 12, 13, 14, 15, 16,
    17, 18, 19, 20, 21, 22, 23, 24,
    25, 26, 27, 28, 29, 30, 31, 32,
    33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48,
    49, 50, 51, 52, 53, 54, 55, 56,
    57, 58, 59, 60, 61, 62, 63, 64, 65,
    mm = 100), 'multiple values for keyword argument "mm"')
