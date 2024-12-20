# Tests of resolver errors.
#
# The initial environment contains the predeclared names "M"
# (module-specific) and "U" (universal). This distinction
# should be unobservable to the Starlark program.

# use of declared global
x = 1
_ = x

---
# premature use of global
_ = x ### "undefined: x"
x = 1

---
# use of undefined global
_ = x ### "undefined: x"

---
# redeclaration of global
x = 1
x = 2 ### "cannot reassign global x declared at .*resolve.star:22:1"

---
# Redeclaration of predeclared names is allowed.
#
# This rule permits tool maintainers to add members to the predeclared
# environment without breaking existing programs.

# module-specific predeclared name
M = 1 # ok
M = 2 ### "cannot reassign global M declared at .*/resolve.star"

# universal predeclared name
U = 1 # ok
U = 1 ### "cannot reassign global U declared at .*/resolve.star"

---
# reference to predeclared name
M()

---
# locals may be referenced before they are defined

def f():
   M(x) # dynamic error
   x = 1

---
# Various forms of assignment:

def f(x): # parameter
    M(x)
    M(y) ### "undefined: y"

(a, b) = 1, 2
M(a)
M(b)
M(c) ### "undefined: c"

[p, q] = 1, 2
M(p)
M(q)
M(r) ### "undefined: r"

---
# a comprehension introduces a separate lexical block

_ = [x for x in "abc"]
M(x) ### "undefined: x"

---
# Functions may have forward refs.   (option:lambda option:nesteddef)
def f():
   g()
   h() ### "undefined: h"
   def inner():
     i()
     i = lambda: 0


def g():
  f()

---
# It's permitted to rebind a global using a += assignment.

x = [1]
x.extend([2]) # ok
x += [3] # ok (a list mutation, not a global rebinding)

def f():
   x += [4] # x is local to f

y = 1
y += 2 # ok (even though it is in fact a global rebinding)

z += 3 # ok (but fails dynamically because z is undefined)

---
def f(a):
  if 1==1:
    b = 1
  c = 1
  M(a) # ok: param
  M(b) # ok: maybe bound local
  M(c) # ok: bound local
  M(d) # NB: we don't do a use-before-def check on local vars!
  M(e) # ok: global
  M(f) # ok: global
  d = 1

e = 1

---
# This program should resolve successfully but fail dynamically.
# However, the Java implementation currently reports the dynamic
# error at the x=2 statement.
x = 1

def f():
  M(x) # dynamic error: reference to undefined local
  x = 2

f()

---
load("module", "name") # ok

def f():
  load("foo", "bar") ### "load statement within a function"

load("foo",
     "",     ### "load: empty identifier"
     "_a",   ### "load: names with leading underscores are not exported: _a"
     b="",   ### "load: empty identifier"
     c="_d", ### "load: names with leading underscores are not exported: _d"
     _e="f") # ok

---
# return, if statements and for loops at top-level are forbidden

for x in "abc": ### "for loop not within a function"
  pass

if x: ### "if statement not within a function"
  pass

return ### "return statement not within a function"

---
# The parser allows any expression on the LHS of an assignment.

1 = 2 ### "can't assign to literal"
1+2 = 3 ### "can't assign to binaryexpr"
f() = 4 ### "can't assign to callexpr"

[a, b] = [1, 2]
[a, b] += [3, 4] ### "can't use list expression in augmented assignment"
(a, b) += [3, 4] ### "can't use tuple expression in augmented assignment"
[] = [] ### "can't assign to \\[\\]"
() = () ### "can't assign to ()"

---
# break and continue statements must appear within a loop

break ### "break not in a loop"

continue ### "continue not in a loop"

pass

---
# Positional arguments (and required parameters)
# must appear before named arguments (and optional parameters).

M(x=1, 2) ### `positional argument may not follow named`

def f(x=1, y): pass ### `required parameter may not follow optional`
---
# No parameters may follow **kwargs

def f(**kwargs, x): ### `parameter may not follow \*\*kwargs`
  pass

def g(**kwargs, *args): ### `\*args may not follow \*\*kwargs`
  pass

def h(**kwargs1, **kwargs2): ### `multiple \*\*kwargs not allowed`
  pass

---
# Only **kwargs may follow *args

def f(*args, x): ### `parameter may not follow \*args`
  pass

def g(*args1, *args2): ### `multiple \*args not allowed`
  pass

def h(*args, **kwargs): # ok
  pass

---
# No arguments may follow **kwargs
def f(*args, **kwargs):
  pass

f(**{}, 1) ### `argument may not follow \*\*kwargs`
f(**{}, x=1) ### `argument may not follow \*\*kwargs`
f(**{}, *[]) ### `\*args may not follow \*\*kwargs`
f(**{}, **{}) ### `multiple \*\*kwargs not allowed`

---
# Only keyword arguments may follow *args
def f(*args, **kwargs):
  pass

f(*[], 1) ### `argument may not follow \*args`
f(*[], a=1) # ok
f(*[], *[]) ### `multiple \*args not allowed`
f(*[], **{}) # ok

---
# Parameter names must be unique.

def f(a, b, a): pass ### "duplicate parameter: a"
def g(args, b, *args): pass ### "duplicate parameter: args"
def h(kwargs, a, **kwargs): pass ### "duplicate parameter: kwargs"
def i(*x, **x): pass ### "duplicate parameter: x"

---
# No floating point
a = float("3.141") ### `dialect does not support floating point`
b = 1 / 2          ### `dialect does not support floating point \(use //\)`
c = 3.141          ### `dialect does not support floating point`
---
# Floating point support (option:float)
a = float("3.141")
b = 1 / 2
c = 3.141
