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
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"devt.de/krotik/ecal/stdlib"
)

func TestStdlib(t *testing.T) {
	stdlib.AddStdlibPkg("fmt", "fmt package")
	stdlib.AddStdlibFunc("fmt", "Sprint",
		stdlib.NewECALFunctionAdapter(reflect.ValueOf(fmt.Sprint), "foo"))

	res, err := UnitTestEvalAndAST(
		`fmt.Sprint([1,2,3])`, nil,
		`
identifier: fmt
  identifier: Sprint
    funccall
      list
        number: 1
        number: 2
        number: 3
`[1:])

	if err != nil || res != "[1 2 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`fmt.Sprint(math.Pi)`, nil,
		`
identifier: fmt
  identifier: Sprint
    funccall
      identifier: math
        identifier: Pi
`[1:])

	if err != nil || res != "3.141592653589793" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	// Negative case

	res, err = UnitTestEvalAndAST(
		`a.fmtSprint([1,2,3])`, nil,
		`
identifier: a
  identifier: fmtSprint
    funccall
      list
        number: 1
        number: 2
        number: 3
`[1:])

	if err == nil ||
		err.Error() != "ECAL error in ECALTestRuntime: Unknown construct (Unknown function: fmtSprint) (Line:1 Pos:3)" {
		t.Error("Unexpected result: ", res, err)
		return
	}
}

func TestSimpleFunctions(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`len([1,2,3])`, nil,
		`
identifier: len
  funccall
    list
      number: 1
      number: 2
      number: 3
`[1:])

	if err != nil || res != 3. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`len({"a":1, 2:"b"})`, nil,
		`
identifier: len
  funccall
    map
      kvp
        string: 'a'
        number: 1
      kvp
        number: 2
        string: 'b'
`[1:])

	if err != nil || res != 2. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`del([1,2,3], 1)`, nil,
		`
identifier: del
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 1
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`del({
  "a" : 1
  "b" : 2
  "c" : 3
}, "b")`, nil,
		`
identifier: del
  funccall
    map
      kvp
        string: 'a'
        number: 1
      kvp
        string: 'b'
        number: 2
      kvp
        string: 'c'
        number: 3
    string: 'b'
`[1:])

	if err != nil || fmt.Sprint(res) != "map[a:1 c:3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`add([1,2,3], 4)`, nil,
		`
identifier: add
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 4
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 2 3 4]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`add([1,2,3], 4, 0)`, nil,
		`
identifier: add
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 4
    number: 0
`[1:])

	if err != nil || fmt.Sprint(res) != "[4 1 2 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`add([1,2,3], 4, 1)`, nil,
		`
identifier: add
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 4
    number: 1
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 4 2 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`concat([1,2,3], [4,5,6], [7,8,9])`, nil,
		`
identifier: concat
  funccall
    list
      number: 1
      number: 2
      number: 3
    list
      number: 4
      number: 5
      number: 6
    list
      number: 7
      number: 8
      number: 9
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 2 3 4 5 6 7 8 9]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`dumpenv()`, nil,
		`
identifier: dumpenv
  funccall
`[1:])

	if err != nil || fmt.Sprint(res) != `GlobalScope {
}` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`
func foo() {
	log("hello")
}
doc(foo)`, nil)

	if err != nil || fmt.Sprint(res) != `Declared function: foo (Line 2, Pos 1)` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`doc(len)`, nil)

	if err != nil || fmt.Sprint(res) != `Len returns the size of a list or map.` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	stdlib.AddStdlibPkg("fmt", "fmt package")
	stdlib.AddStdlibFunc("fmt", "Println",
		stdlib.NewECALFunctionAdapter(reflect.ValueOf(fmt.Sprint), "foo"))

	res, err = UnitTestEval(
		`doc(fmt.Println)`, nil)

	if err != nil || res != "foo" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`
/*
Foo is my custom function.
*/
func foo() {
	log("hello")
}
doc(foo)`, nil)

	if err != nil || fmt.Sprint(res) != `Foo is my custom function.` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	// Negative case

	res, err = UnitTestEvalAndAST(
		`a.len([1,2,3])`, nil,
		`
identifier: a
  identifier: len
    funccall
      list
        number: 1
        number: 2
        number: 3
`[1:])

	if err == nil ||
		err.Error() != "ECAL error in ECALTestRuntime: Unknown construct (Unknown function: len) (Line:1 Pos:3)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	_, err = UnitTestEval(`sleep(10)`, nil)

	if err != nil {
		t.Error("Unexpected result: ", err)
		return
	}
}

func TestCronTrigger(t *testing.T) {

	res, err := UnitTestEval(
		`setCronTrigger("1 * * * *", "foo", "bar")`, nil)

	if err == nil ||
		err.Error() != "ECAL error in ECALTestRuntime: Runtime error (Cron spec must have 6 entries separated by space) (Line:1 Pos:1)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`
sink test
  kindmatch [ "foo.*" ],
{
	log("test rule - Handling request: ", event)
}

log("Cron:", setCronTrigger("1 1 *%10 * * *", "cronevent", "foo.bar"))
`, nil)

	if err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	testcron.Start()
	time.Sleep(200 * time.Millisecond)

	if testlogger.String() != `
Cron:at second 1 of minute 1 of every 10th hour every day
test rule - Handling request: {
  "kind": "foo.bar",
  "name": "cronevent",
  "state": {
    "tick": 1,
    "time": "2000-01-01T00:01:01Z",
    "timestamp": "946684861000"
  }
}
test rule - Handling request: {
  "kind": "foo.bar",
  "name": "cronevent",
  "state": {
    "tick": 2,
    "time": "2000-01-01T10:01:01Z",
    "timestamp": "946720861000"
  }
}
test rule - Handling request: {
  "kind": "foo.bar",
  "name": "cronevent",
  "state": {
    "tick": 3,
    "time": "2000-01-01T20:01:01Z",
    "timestamp": "946756861000"
  }
}`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}
}

func TestPulseTrigger(t *testing.T) {

	res, err := UnitTestEval(
		`setPulseTrigger("test", "foo", "bar")`, nil)

	if err == nil ||
		err.Error() != "ECAL error in ECALTestRuntime: Runtime error (Parameter 1 should be a number) (Line:1 Pos:1)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEval(
		`
sink test
  kindmatch [ "foo.*" ],
{
	log("test rule - Handling request: ", event)
	log("Duration: ", event.state.currentMicros - event.state.lastMicros," us (microsecond)")
}

setPulseTrigger(100, "pulseevent", "foo.bar")
`, nil)

	if err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	time.Sleep(100 * time.Millisecond)
	testprocessor.Finish()

	if !strings.Contains(testlogger.String(), "Handling request") {
		t.Error("Unexpected result:", testlogger.String())
		return
	}
}

func TestDocstrings(t *testing.T) {
	for k, v := range InbuildFuncMap {
		if res, _ := v.DocString(); res == "" {
			t.Error("Docstring missing for ", k)
			return
		}
	}
}

func TestErrorConditions(t *testing.T) {

	ib := &inbuildBaseFunc{}

	if _, err := ib.AssertNumParam(1, "bob"); err == nil || err.Error() != "Parameter 1 should be a number" {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := ib.AssertMapParam(1, "bob"); err == nil || err.Error() != "Parameter 1 should be a map" {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := ib.AssertListParam(1, "bob"); err == nil || err.Error() != "Parameter 1 should be a list" {
		t.Error("Unexpected result:", err)
		return
	}

	rf := &rangeFunc{&inbuildBaseFunc{}}

	if _, err := rf.Run("", nil, nil, 0, nil); err == nil || err.Error() != "Need at least an end range as first parameter" {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := rf.Run("", nil, nil, 0, []interface{}{"bob"}); err == nil || err.Error() != "Parameter 1 should be a number" {
		t.Error("Unexpected result:", err)
		return
	}

}
