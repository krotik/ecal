/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package interpreter

import (
	"testing"

	"devt.de/krotik/ecal/scope"
)

func TestGuardStatements(t *testing.T) {

	// Test normal if

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
a := 1
if a == 1 {
	b := 1
    a := a + 1	
}
`, vs,
		`
statements
  :=
    identifier: a
    number: 1
  if
    guard
      ==
        identifier: a
        number: 1
    statements
      :=
        identifier: b
        number: 1
      :=
        identifier: a
        plus
          identifier: a
          number: 1
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    a (float64) : 2
    block: if (Line:3 Pos:1) {
        b (float64) : 1
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	// Test elif

	vs = scope.NewScope(scope.GlobalScope)

	_, err = UnitTestEvalAndAST(
		`
	   a := 2
	   if a == 1 {
	       a := a + 1
	   } elif a == 2 {
	   	a := a + 2
	   }
	   `, vs, `
statements
  :=
    identifier: a
    number: 2
  if
    guard
      ==
        identifier: a
        number: 1
    statements
      :=
        identifier: a
        plus
          identifier: a
          number: 1
    guard
      ==
        identifier: a
        number: 2
    statements
      :=
        identifier: a
        plus
          identifier: a
          number: 2
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    a (float64) : 4
    block: if (Line:3 Pos:5) {
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	// Test else

	vs = scope.NewScope(scope.GlobalScope)

	_, err = UnitTestEvalAndAST(
		`
	   a := 3
	   if a == 1 {
	       a := a + 1
	   } elif a == 2 {
	   	a := a + 2
	   } else {
	       a := 99
	   }
	   `, vs, `
statements
  :=
    identifier: a
    number: 3
  if
    guard
      ==
        identifier: a
        number: 1
    statements
      :=
        identifier: a
        plus
          identifier: a
          number: 1
    guard
      ==
        identifier: a
        number: 2
    statements
      :=
        identifier: a
        plus
          identifier: a
          number: 2
    guard
      true
    statements
      :=
        identifier: a
        number: 99
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    a (float64) : 99
    block: if (Line:3 Pos:5) {
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}
}

func TestLoopStatements(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)
	buf := addLogFunction(vs)

	_, err := UnitTestEvalAndAST(
		`
a := 10

for a > 0 {

	testlog("Info: ", "-> ", a)
	a := a - 1
}
`, vs,
		`
statements
  :=
    identifier: a
    number: 10
  loop
    guard
      >
        identifier: a
        number: 0
    statements
      identifier: testlog
        funccall
          string: 'Info: '
          string: '-> '
          identifier: a
      :=
        identifier: a
        minus
          identifier: a
          number: 1
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    a (float64) : 0
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:4 Pos:1) {
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info: -> 10
Info: -> 9
Info: -> 8
Info: -> 7
Info: -> 6
Info: -> 5
Info: -> 4
Info: -> 3
Info: -> 2
Info: -> 1`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)
	buf = addLogFunction(vs)

	_, err = UnitTestEvalAndAST(
		`
	   for a in range(2, 10, 1) {
           testlog("Info", "->", a)
	   }
	   `, vs,
		`
loop
  in
    identifier: a
    identifier: range
      funccall
        number: 2
        number: 10
        number: 1
  statements
    identifier: testlog
      funccall
        string: 'Info'
        string: '->'
        identifier: a
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:2 Pos:5) {
        a (float64) : 10
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info->2
Info->3
Info->4
Info->5
Info->6
Info->7
Info->8
Info->9
Info->10`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)
	buf = addLogFunction(vs)

	_, err = UnitTestEvalAndAST(
		`
for a in range(10, 3, -3) {
  testlog("Info", "->", a)
}
	   `, vs,
		`
loop
  in
    identifier: a
    identifier: range
      funccall
        number: 10
        number: 3
        minus
          number: 3
  statements
    identifier: testlog
      funccall
        string: 'Info'
        string: '->'
        identifier: a
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:2 Pos:1) {
        a (float64) : 4
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info->10
Info->7
Info->4`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}
}

func TestLoopStatements2(t *testing.T) {

	// Test nested loops

	vs := scope.NewScope(scope.GlobalScope)
	buf := addLogFunction(vs)

	_, err := UnitTestEvalAndAST(
		`
for a in range(10, 3, -3) {
  for b in range(1, 3, 1) {
    testlog("Info", "->", a, b)
  }
}
	   `, vs,
		`
loop
  in
    identifier: a
    identifier: range
      funccall
        number: 10
        number: 3
        minus
          number: 3
  statements
    loop
      in
        identifier: b
        identifier: range
          funccall
            number: 1
            number: 3
            number: 1
      statements
        identifier: testlog
          funccall
            string: 'Info'
            string: '->'
            identifier: a
            identifier: b
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:2 Pos:1) {
        a (float64) : 4
        block: loop (Line:3 Pos:3) {
            b (float64) : 3
        }
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info->10 1
Info->10 2
Info->10 3
Info->7 1
Info->7 2
Info->7 3
Info->4 1
Info->4 2
Info->4 3`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	// Break statement

	vs = scope.NewScope(scope.GlobalScope)
	buf = addLogFunction(vs)

	_, err = UnitTestEvalAndAST(
		`
for a in range(1, 10, 1) {
  testlog("Info", "->", a)
  if a == 3 {
    break
  }
}`, vs,
		`
loop
  in
    identifier: a
    identifier: range
      funccall
        number: 1
        number: 10
        number: 1
  statements
    identifier: testlog
      funccall
        string: 'Info'
        string: '->'
        identifier: a
    if
      guard
        ==
          identifier: a
          number: 3
      statements
        break
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:2 Pos:1) {
        a (float64) : 3
        block: if (Line:4 Pos:3) {
        }
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info->1
Info->2
Info->3`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	// Continue statement

	vs = scope.NewScope(scope.GlobalScope)
	buf = addLogFunction(vs)

	_, err = UnitTestEvalAndAST(
		`
for a in range(1, 10, 1) {
  if a > 3 and a < 6  {
    continue
  }
  testlog("Info", "->", a)
}`, vs,
		`
loop
  in
    identifier: a
    identifier: range
      funccall
        number: 1
        number: 10
        number: 1
  statements
    if
      guard
        and
          >
            identifier: a
            number: 3
          <
            identifier: a
            number: 6
      statements
        continue
    identifier: testlog
      funccall
        string: 'Info'
        string: '->'
        identifier: a
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:2 Pos:1) {
        a (float64) : 10
        block: if (Line:3 Pos:3) {
        }
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info->1
Info->2
Info->3
Info->6
Info->7
Info->8
Info->9
Info->10`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

}

func TestLoopStatements3(t *testing.T) {

	// Loop over lists

	vs := scope.NewScope(scope.GlobalScope)
	buf := addLogFunction(vs)

	_, err := UnitTestEvalAndAST(
		`
for a in [1,2] {
  for b in [1,2,3,"Hans", 4] {
    testlog("Info", "->", a, "-", b)
  }
}
	   `, vs,
		`
loop
  in
    identifier: a
    list
      number: 1
      number: 2
  statements
    loop
      in
        identifier: b
        list
          number: 1
          number: 2
          number: 3
          string: 'Hans'
          number: 4
      statements
        identifier: testlog
          funccall
            string: 'Info'
            string: '->'
            identifier: a
            string: '-'
            identifier: b
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:2 Pos:1) {
        a (float64) : 2
        block: loop (Line:3 Pos:3) {
            b (float64) : 4
        }
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info->1-1
Info->1-2
Info->1-3
Info->1-Hans
Info->1-4
Info->2-1
Info->2-2
Info->2-3
Info->2-Hans
Info->2-4`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)
	buf = addLogFunction(vs)

	_, err = UnitTestEvalAndAST(
		`
l := [1,2,3,4]
for a in range(0, 3, 1) {
  testlog("Info", "-a>", a, "-", l[a])
}
for a in range(0, 3, 1) {
  testlog("Info", "-b>", a, "-", l[-a])
}
testlog("Info", "xxx>", l[-1])
	   `, vs,
		`
statements
  :=
    identifier: l
    list
      number: 1
      number: 2
      number: 3
      number: 4
  loop
    in
      identifier: a
      identifier: range
        funccall
          number: 0
          number: 3
          number: 1
    statements
      identifier: testlog
        funccall
          string: 'Info'
          string: '-a>'
          identifier: a
          string: '-'
          identifier: l
            compaccess
              identifier: a
  loop
    in
      identifier: a
      identifier: range
        funccall
          number: 0
          number: 3
          number: 1
    statements
      identifier: testlog
        funccall
          string: 'Info'
          string: '-b>'
          identifier: a
          string: '-'
          identifier: l
            compaccess
              minus
                identifier: a
  identifier: testlog
    funccall
      string: 'Info'
      string: 'xxx>'
      identifier: l
        compaccess
          minus
            number: 1
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `
GlobalScope {
    l ([]interface {}) : [1,2,3,4]
    testlog (*interpreter.TestLogger) : TestLogger
    block: loop (Line:3 Pos:1) {
        a (float64) : 3
    }
    block: loop (Line:6 Pos:1) {
        a (float64) : 3
    }
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := buf.String(); res != `
Info-a>0-1
Info-a>1-2
Info-a>2-3
Info-a>3-4
Info-b>0-1
Info-b>1-4
Info-b>2-3
Info-b>3-2
Infoxxx>4`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	// Loop over a map

	vs = scope.NewScope(scope.GlobalScope)
	buf = addLogFunction(vs)

	_, err = UnitTestEvalAndAST(
		`
x := { "c": 0, "a":2, "b":4}
for [a, b] in x {
  testlog("Info", "->", a, "-", b)
}
	   `, vs,
		`
statements
  :=
    identifier: x
    map
      kvp
        string: 'c'
        number: 0
      kvp
        string: 'a'
        number: 2
      kvp
        string: 'b'
        number: 4
  loop
    in
      list
        identifier: a
        identifier: b
      identifier: x
    statements
      identifier: testlog
        funccall
          string: 'Info'
          string: '->'
          identifier: a
          string: '-'
          identifier: b
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if res := buf.String(); res != `
Info->a-2
Info->b-4
Info->c-0`[1:] {
		t.Error("Unexpected result: ", res)
		return
	}

	_, err = UnitTestEval(
		`
x := { "c": 0, "a":2, "b":4}
for [1, b] in x {
  testlog("Info", "->", a, "-", b)
}
	   `, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid construct (Must have a list of simple variables on the left side of the In expression) (Line:3 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}
}

func TestLoopStatements4(t *testing.T) {
	vs := scope.NewScope(scope.GlobalScope)

	// Test continue

	_, err := UnitTestEval(`
for [a] in [1,2,3] {
  continue
  [a, b] := "Hans"
}
	   `[1:], vs)

	if err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(`
a := 1
for a < 10 {
  a := a + 1
  continue
  [a,b] := "Hans"
}
	   `[1:], vs)

	if err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	// Test single value

	_, err = UnitTestEval(`
for a in 1 {
  continue
  [a,b] := "Hans"
}
	   `[1:], vs)

	if err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(`
for a[t] in 1 {
  continue
  [a,b] := "Hans"
}
	   `[1:], vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid construct (Must have a simple variable on the left side of the In expression) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(`
for [a, b] in [[1,2],[3,4],3] {
}
	   `[1:], vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Runtime error (Result for loop variable is not a list (value is 3)) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(`
for [a, b] in [[1,2],[3,4],[5,6,7]] {
}
	   `[1:], vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Runtime error (Assigned number of variables is different to number of values (2 variables vs 3 values)) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}
}

func TestTryStatements(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
try {
	debug("Raising custom error")
    raise("test 12", null, [1,2,3])
} except "test 12" as e {
	error("Something happened: ", e)
} finally {
	log("Cleanup")
}
`, vs,
		`
try
  statements
    identifier: debug
      funccall
        string: 'Raising custom error'
    identifier: raise
      funccall
        string: 'test 12'
        null
        list
          number: 1
          number: 2
          number: 3
  except
    string: 'test 12'
    as
      identifier: e
    statements
      identifier: error
        funccall
          string: 'Something happened: '
          identifier: e
  finally
    statements
      identifier: log
        funccall
          string: 'Cleanup'
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
debug: Raising custom error
error: Something happened: {
  "data": [
    1,
    2,
    3
  ],
  "detail": "",
  "error": "ECAL error in ECALTestRuntime (ECALEvalTest): test 12 () (Line:4 Pos:5)",
  "line": 4,
  "pos": 5,
  "source": "ECALTestRuntime (ECALEvalTest)",
  "trace": [
    "raise(\"test 12\", null, [1, 2, 3]) (ECALEvalTest:4)"
  ],
  "type": "test 12"
}
Cleanup`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	_, err = UnitTestEval(
		`
try {
	debug("Raising custom error")
    raise("test 13", null, [1,2,3])
} except "test 12" as e {
	error("Something happened: ", e)
} except e {
	error("Something else happened: ", e)

	try {
		x := 1 + a
	} except e {
		log("Runtime error: ", e)
	}

} finally {
	log("Cleanup")
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
debug: Raising custom error
error: Something else happened: {
  "data": [
    1,
    2,
    3
  ],
  "detail": "",
  "error": "ECAL error in ECALTestRuntime (ECALEvalTest): test 13 () (Line:4 Pos:5)",
  "line": 4,
  "pos": 5,
  "source": "ECALTestRuntime (ECALEvalTest)",
  "trace": [
    "raise(\"test 13\", null, [1, 2, 3]) (ECALEvalTest:4)"
  ],
  "type": "test 13"
}
Runtime error: {
  "detail": "a=NULL",
  "error": "ECAL error in ECALTestRuntime (ECALEvalTest): Operand is not a number (a=NULL) (Line:11 Pos:12)",
  "line": 11,
  "pos": 12,
  "source": "ECALTestRuntime (ECALEvalTest)",
  "trace": [],
  "type": "Operand is not a number"
}
Cleanup`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	_, err = UnitTestEval(
		`
try {
	x := 1 + "a"
} except {
	error("This did not work")
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
error: This did not work`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	_, err = UnitTestEval(
		`
try {
	try {
		x := 1 + "a"
	} except e {
		raise("usererror", "This did not work", e)
	}
} except e {
	error(e)
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
error: {
  "data": {
    "detail": "a",
    "error": "ECAL error in ECALTestRuntime (ECALEvalTest): Operand is not a number (a) (Line:4 Pos:12)",
    "line": 4,
    "pos": 12,
    "source": "ECALTestRuntime (ECALEvalTest)",
    "trace": [],
    "type": "Operand is not a number"
  },
  "detail": "This did not work",
  "error": "ECAL error in ECALTestRuntime (ECALEvalTest): usererror (This did not work) (Line:6 Pos:3)",
  "line": 6,
  "pos": 3,
  "source": "ECALTestRuntime (ECALEvalTest)",
  "trace": [
    "raise(\"usererror\", \"This did not work\", e) (ECALEvalTest:6)"
  ],
  "type": "usererror"
}`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	_, err = UnitTestEval(
		`
try {
  x := 1
} except e {
  raise("usererror", "This did not work", e)
} otherwise {
  log("all good")
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
all good`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}
}

func TestMutexStatements(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
a := 2
mutex foo {
	a := 1
    raise("test 12", null, [1,2,3])
}
`, vs,
		`
statements
  :=
    identifier: a
    number: 2
  mutex
    identifier: foo
    statements
      :=
        identifier: a
        number: 1
      identifier: raise
        funccall
          string: 'test 12'
          null
          list
            number: 1
            number: 2
            number: 3
`[1:])

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): test 12 () (Line:5 Pos:5)" {
		t.Error(err)
		return
	}

	if vs.String() != `GlobalScope {
    a (float64) : 1
    block: mutex (Line:3 Pos:1) {
    }
}` {
		t.Error("Unexpected variable scope:", vs)
	}

	// Can take mutex twice

	_, err = UnitTestEvalAndAST(
		`
a := 2
mutex foo {
	a := 1
	mutex foo {
		a := 3
	}
}
`, vs,
		`
statements
  :=
    identifier: a
    number: 2
  mutex
    identifier: foo
    statements
      :=
        identifier: a
        number: 1
      mutex
        identifier: foo
        statements
          :=
            identifier: a
            number: 3
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	if vs.String() != `GlobalScope {
    a (float64) : 3
    block: mutex (Line:3 Pos:1) {
        block: mutex (Line:5 Pos:2) {
        }
    }
}` {
		t.Error("Unexpected variable scope:", vs)
	}
}
