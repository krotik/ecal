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

func TestSimpleAssignments(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	res, err := UnitTestEvalAndAST(
		`let a := 42`, vs,
		`
:=
  let
    identifier: a
  number: 42
`[1:])

	if vsRes := vs.String(); vsRes != `GlobalScope {
    a (float64) : 42
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`a := "test"`, vs,
		`
:=
  identifier: a
  string: 'test'
`[1:])

	if vsRes := vs.String(); vsRes != `GlobalScope {
    a (string) : test
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`a := [1,2,3]`, vs,
		`
:=
  identifier: a
  list
    number: 1
    number: 2
    number: 3
`[1:])

	if vsRes := vs.String(); vsRes != `GlobalScope {
    a ([]interface {}) : [1,2,3]
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`a := {
			1 : "foo",
			"2" : "bar",
			"foobar" : 3,
			x : 4,
		}`, vs,
		`
:=
  identifier: a
  map
    kvp
      number: 1
      string: 'foo'
    kvp
      string: '2'
      string: 'bar'
    kvp
      string: 'foobar'
      number: 3
    kvp
      identifier: x
      number: 4
`[1:])

	if vsRes := vs.String(); vsRes != `GlobalScope {
    a (map[interface {}]interface {}) : {"1":"foo","2":"bar","foobar":3,"null":4}
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	_, err = UnitTestEval(
		`1 := [1, 2]`, vs)

	if err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Cannot access variable (Must have a variable or list of variables on the left side of the assignment) (Line:1 Pos:3)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`[1] := [1, 2]`, vs)

	if err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Cannot access variable (Must have a list of variables on the left side of the assignment) (Line:1 Pos:5)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`[a, b] := [1, 2, 3]`, vs)

	if err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid state (Assigned number of variables is different to number of values (2 variables vs 3 values)) (Line:1 Pos:8)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`[a, b] := 1`, vs)

	if err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid state (Result is not a list (value is 1)) (Line:1 Pos:8)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`[a, b.c, c] := [1, 2, 3]`, vs)

	if err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Cannot access variable (Variable b is not a container) (Line:1 Pos:13)" {
		t.Error("Unexpected result:", err)
		return
	}
}

func TestComplexAssignments(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	res, err := UnitTestEvalAndAST(
		`let [a, b] := ["test", [1,2,3]]`, vs,
		`
:=
  let
    list
      identifier: a
      identifier: b
  list
    string: 'test'
    list
      number: 1
      number: 2
      number: 3
`[1:])

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (string) : test
    b ([]interface {}) : [1,2,3]
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEvalAndAST(`
a := {
	"b" : [ 0, {
		"c" : [ 0, 1, 2, {
			"test" : 1
		}]
	}]
}
a.b[1].c["3"]["test"] := 3
`, vs,
		`
statements
  :=
    identifier: a
    map
      kvp
        string: 'b'
        list
          number: 0
          map
            kvp
              string: 'c'
              list
                number: 0
                number: 1
                number: 2
                map
                  kvp
                    string: 'test'
                    number: 1
  :=
    identifier: a
      identifier: b
        compaccess
          number: 1
        identifier: c
          compaccess
            string: '3'
          compaccess
            string: 'test'
    number: 3
`[1:])

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (map[interface {}]interface {}) : {"b":[0,{"c":[0,1,2,{"test":3}]}]}
    b ([]interface {}) : [1,2,3]
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEvalAndAST(`
z := a.b[1].c["3"]["test"]
[x, y] := a.b
`, vs,
		`
statements
  :=
    identifier: z
    identifier: a
      identifier: b
        compaccess
          number: 1
        identifier: c
          compaccess
            string: '3'
          compaccess
            string: 'test'
  :=
    list
      identifier: x
      identifier: y
    identifier: a
      identifier: b
`[1:])

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (map[interface {}]interface {}) : {"b":[0,{"c":[0,1,2,{"test":3}]}]}
    b ([]interface {}) : [1,2,3]
    x (float64) : 0
    y (map[interface {}]interface {}) : {"c":[0,1,2,{"test":3}]}
    z (float64) : 3
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

}

func TestScopedDeclaration(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	res, err := UnitTestEval(`
a := 5
func foo() {
	let a := 2
}
foo()`, vs)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (float64) : 5
    foo (*interpreter.function) : ecal.function: foo (Line 3, Pos 1)
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEval(`
a := 5
func foo() {
	let [a, b, c]
	a := 2
}
foo()`, vs)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (float64) : 5
    foo (*interpreter.function) : ecal.function: foo (Line 3, Pos 1)
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	res, err = UnitTestEval(`let [1]`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid construct (Let can only declare variables within a list) (Line:1 Pos:1)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(`let 1`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid construct (Let must declare a variable or list of variables) (Line:1 Pos:1)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(`let a.b`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime (ECALEvalTest): Invalid construct (Let can only declare simple variables) (Line:1 Pos:1)" {
		t.Error("Unexpected result: ", res, err)
		return
	}
}
