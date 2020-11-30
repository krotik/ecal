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

	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

func TestGeneralCases(t *testing.T) {

	rt := NewECALRuntimeProvider("a", nil, nil)
	id1 := rt.NewThreadID()
	id2 := rt.NewThreadID()

	if id1 == id2 {
		t.Error("Thread ids should not be the same:", id1)
		return
	}

	n, _ := parser.Parse("a", "a")
	inv := &invalidRuntime{newBaseRuntime(rt, n)}

	if err := inv.Validate().Error(); err != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := inv.Eval(nil, nil, 0); err.Error() != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	n, _ = parser.Parse("a", "a")
	inv = &invalidRuntime{newBaseRuntime(NewECALRuntimeProvider("a", nil, nil), n)}
	n.Runtime = inv

	n2, _ := parser.Parse("a", "a")
	inv2 := &invalidRuntime{newBaseRuntime(NewECALRuntimeProvider("a", nil, nil), n2)}
	n2.Runtime = inv2
	n.Children = append(n.Children, n2)

	if err := inv.Validate().Error(); err != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	n, _ = parser.Parse("a", "a")
	void := &voidRuntime{newBaseRuntime(NewECALRuntimeProvider("a", nil, nil), n)}
	n.Runtime = void
	void.Validate()

	if res, err := void.Eval(nil, nil, 0); err != nil || res != nil {
		t.Error("Unexpected result:", res, err)
		return
	}
}

func TestImporting(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)
	il := &util.MemoryImportLocator{Files: make(map[string]string)}

	il.Files["foo/bar"] = `
b := 123
`

	res, err := UnitTestEvalAndASTAndImport(
		`
	   import "foo/bar" as foobar
	   a := foobar.b`, vs,
		`
statements
  import
    string: 'foo/bar'
    identifier: foobar
  :=
    identifier: a
    identifier: foobar
      identifier: b
`[1:], il)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (float64) : 123
    foobar (map[interface {}]interface {}) : {"b":123}
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	n, _ := parser.Parse("a", "a")
	imp := &importRuntime{newBaseRuntime(NewECALRuntimeProvider("a", nil, nil), n)}
	n.Runtime = imp
	imp.erp = NewECALRuntimeProvider("ECALTestRuntime", nil, nil)
	imp.erp.ImportLocator = nil
	imp.Validate()

	if res, err := imp.Eval(nil, nil, 0); err == nil || err.Error() != "ECAL error in ECALTestRuntime: Runtime error (No import locator was specified) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", res, err)
		return
	}
}

func TestLogging(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
log("Hello")
debug("foo")
error("bar")
`, vs,
		`
statements
  identifier: log
    funccall
      string: 'Hello'
  identifier: debug
    funccall
      string: 'foo'
  identifier: error
    funccall
      string: 'bar'
`[1:])

	if err != nil {
		t.Error("Unexpected result: ", err)
		return
	}

	if testlogger.String() != `Hello
debug: foo
error: bar` {
		t.Error("Unexpected result: ", testlogger.String())
		return
	}
}

func TestOperatorRuntimeErrors(t *testing.T) {

	n, _ := parser.Parse("a", "a")
	op := &operatorRuntime{newBaseRuntime(NewECALRuntimeProvider("a", nil, nil), n)}

	if res := op.errorDetailString(n.Token, "foo"); res != "a=foo" {
		t.Error("Unexpected result:", res)
	}

	n.Token.Identifier = false

	if res := op.errorDetailString(n.Token, "foo"); res != "a" {
		t.Error("Unexpected result:", res)
	}

	n.Token.Identifier = true

	res, err := UnitTestEval(
		`+ "a"`, nil)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Operand is not a number (a) (Line:1 Pos:3)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`x := "a"; + x`, nil)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Operand is not a number (x=a) (Line:1 Pos:13)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`not "a"`, nil)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Operand is not a boolean (a) (Line:1 Pos:5)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`a:= 5; a or 6`, nil)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Operand is not a boolean (a=5) (Line:1 Pos:8)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`true or 2`, nil)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Operand is not a boolean (2) (Line:1 Pos:1)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`a := "foo"; x in a`, nil)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Operand is not a list (a=foo) (Line:1 Pos:13)" {
		t.Error("Unexpected result: ", res, err)
		return
	}
}
