# Tests of Starlark 'bool'

load("assert.star", "assert")

# truth
assert.true(True)
assert.true(not False)

# bool conversion
assert.eq([bool(), bool(1), bool(0), bool("hello"), bool("")],
          [False, True, False, True, False])

# comparison
assert.true(None == None)
assert.true(None != False)
assert.true(None != True)
assert.eq(1==1, True)
assert.eq(1==2, False)
assert.true(False == False)
assert.true(True == True)

# ordered comparison
assert.true(False < True)
assert.true(False <= True)
assert.true(False <= False)
assert.true(True > False)
assert.true(True >= False)
assert.true(True >= True)

# conditional expression
assert.eq(1 if 3 > 2 else 0, 1)
assert.eq(1 if "foo" else 0, 1)
assert.eq(1 if "" else 0, 0)

# short-circuit evaluation of 'and' and 'or':
# 'or' yields the first true operand, or the last if all are false.
assert.eq(0 or "" or [] or 0, 0) 
assert.eq(0 or "" or [] or 123 or 1/0, 123) 
assert.fails(lambda: 0 or "" or [] or 0 or 1/0, "division by zero")
# 'and' yields the first false operand, or the last if all are true.
assert.eq(1 and "a" and [1] and 123, 123) 
assert.eq(1 and "a" and [1] and 0 and 1/0, 0) 
assert.fails(lambda: 1 and "a" and [1] and 123 and 1/0, "division by zero")
