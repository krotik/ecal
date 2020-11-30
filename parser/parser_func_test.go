/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package parser

import (
	"fmt"
	"testing"
)

func TestImportParsing(t *testing.T) {

	input := `import "foo/bar.ecal" as fooBar
	let i := let fooBar`
	expectedOutput := `
statements
  import
    string: 'foo/bar.ecal'
    identifier: fooBar
  :=
    let
      identifier: i
    let
      identifier: fooBar
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestSinkParsing(t *testing.T) {

	input := `
	sink fooBar
    kindmatch [ "priority", "t.do.bla" ],
	scopematch [ "data.read", "data.write" ],
	statematch { "priority:" : 5, test: 1, "bla 1": null },
	priority 0,
	suppresses [ "test1", test2 ]
	{
		print("test1");
		print("test2")
	}
`
	expectedOutput := `
sink
  identifier: fooBar
  kindmatch
    list
      string: 'priority'
      string: 't.do.bla'
  scopematch
    list
      string: 'data.read'
      string: 'data.write'
  statematch
    map
      kvp
        string: 'priority:'
        number: 5
      kvp
        identifier: test
        number: 1
      kvp
        string: 'bla 1'
        null
  priority
    number: 0
  suppresses
    list
      string: 'test1'
      identifier: test2
  statements
    identifier: print
      funccall
        string: 'test1'
    identifier: print
      funccall
        string: 'test2'
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
	sink mySink
    kindmatch [ "priority", t.do.bla ]
	{
	}
`
	expectedOutput = `
sink
  identifier: mySink
  kindmatch
    list
      string: 'priority'
      identifier: t
        identifier: do
          identifier: bla
  statements
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
	sink fooBar
    ==
	kindmatch [ "priority", "t.do.bla" ]
	{
	}
`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:3 Pos:5)" {
		t.Error(err)
		return
	}
}

func TestFuncParsing(t *testing.T) {

	input := `import "foo/bar.ecal" as foobar

func myfunc(a, b, c=1) {
  foo := a and b and c
  return foo
}
`
	expectedOutput := `
statements
  import
    string: 'foo/bar.ecal'
    identifier: foobar
  function
    identifier: myfunc
    params
      identifier: a
      identifier: b
      preset
        identifier: c
        number: 1
    statements
      :=
        identifier: foo
        and
          and
            identifier: a
            identifier: b
          identifier: c
      return
        identifier: foo
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
func myfunc() {
  a := 1
  return
  b := 2
  return
}
`
	expectedOutput = `
function
  identifier: myfunc
  params
  statements
    :=
      identifier: a
      number: 1
    return
    :=
      identifier: b
      number: 2
    return
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
func() {
  a := 1
  return
  b := 2
  return
}
`
	expectedOutput = `
function
  params
  statements
    :=
      identifier: a
      number: 1
    return
    :=
      identifier: b
      number: 2
    return
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestFunctionCalling(t *testing.T) {

	input := `import "foo/bar.ecal" as foobar
	foobar.test()`
	expectedOutput := `
statements
  import
    string: 'foo/bar.ecal'
    identifier: foobar
  identifier: foobar
    identifier: test
      funccall
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `a := 1
a().foo := x2.foo()
a.b.c().foo := a()
	`
	expectedOutput = `
statements
  :=
    identifier: a
    number: 1
  :=
    identifier: a
      funccall
      identifier: foo
    identifier: x2
      identifier: foo
        funccall
  :=
    identifier: a
      identifier: b
        identifier: c
          funccall
          identifier: foo
    identifier: a
      funccall
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `a(1+2).foo := x2.foo(foo)
a.b.c(x()).foo := a(1,a(),3, x, y) + 1
	`
	expectedOutput = `
statements
  :=
    identifier: a
      funccall
        plus
          number: 1
          number: 2
      identifier: foo
    identifier: x2
      identifier: foo
        funccall
          identifier: foo
  :=
    identifier: a
      identifier: b
        identifier: c
          funccall
            identifier: x
              funccall
          identifier: foo
    plus
      identifier: a
        funccall
          number: 1
          identifier: a
            funccall
          number: 3
          identifier: x
          identifier: y
      number: 1
`[1:]

	if res, err := UnitTestParseWithPPResult("mytest", input, ""); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}
