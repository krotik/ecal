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

func TestAssignmentParsing(t *testing.T) {

	input := `
z := a.b[1].c["3"]["test"]
[x, y] := a.b
`
	expectedOutput := `
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
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

}

func TestTryContext(t *testing.T) {

	input := `
try {
	raise("test", [1,2,3])
} except "test", "bla" as e {
	print(1)
} except e {
	print(1)
} except {
	print(1)
} finally {
	print(2)
}
`
	expectedOutput := `
try
  statements
    identifier: raise
      funccall
        string: 'test'
        list
          number: 1
          number: 2
          number: 3
  except
    string: 'test'
    string: 'bla'
    as
      identifier: e
    statements
      identifier: print
        funccall
          number: 1
  except
    identifier: e
    statements
      identifier: print
        funccall
          number: 1
  except
    statements
      identifier: print
        funccall
          number: 1
  finally
    statements
      identifier: print
        funccall
          number: 2
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
try {
	raise("test", [1,2,3])
}
`
	expectedOutput = `
try
  statements
    identifier: raise
      funccall
        string: 'test'
        list
          number: 1
          number: 2
          number: 3
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
try {
	raise("test", [1,2,3])
} finally {
}
`
	expectedOutput = `
try
  statements
    identifier: raise
      funccall
        string: 'test'
        list
          number: 1
          number: 2
          number: 3
  finally
    statements
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestLoopParsing(t *testing.T) {

	input := `
for a != null {
	print(1);
	print(2);
	break
	continue
}
`
	expectedOutput := `
loop
  guard
    !=
      identifier: a
      null
  statements
    identifier: print
      funccall
        number: 1
    identifier: print
      funccall
        number: 2
    break
    continue
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
for a in range(1,2) {
	print(1);
	print(2)
}
`
	expectedOutput = `
loop
  in
    identifier: a
    identifier: range
      funccall
        number: 1
        number: 2
  statements
    identifier: print
      funccall
        number: 1
    identifier: print
      funccall
        number: 2
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return

	}
	input = `
for a < 1 and b > 2 {
	print(1)
	print(2)
}
`
	expectedOutput = `
loop
  guard
    and
      <
        identifier: a
        number: 1
      >
        identifier: b
        number: 2
  statements
    identifier: print
      funccall
        number: 1
    identifier: print
      funccall
        number: 2
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
for a in range(1,2,3) {
	==
}
`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:3 Pos:2)" {
		t.Error(err)
		return
	}

	input = `
for a in == {
	@print(1)
}
`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:2 Pos:10)" {
		t.Error(err)
		return
	}
}

func TestConditionalParsing(t *testing.T) {

	input := `
if a == b or c < d {
    print(1);
	foo := 1
} elif x or y {
	x := 1; y := 2; p := {
		1:2
	}
} elif true {
	x := 1; y := 2
} else {
	x := 1
}
`
	expectedOutput := `
if
  guard
    or
      ==
        identifier: a
        identifier: b
      <
        identifier: c
        identifier: d
  statements
    identifier: print
      funccall
        number: 1
    :=
      identifier: foo
      number: 1
  guard
    or
      identifier: x
      identifier: y
  statements
    :=
      identifier: x
      number: 1
    :=
      identifier: y
      number: 2
    :=
      identifier: p
      map
        kvp
          number: 1
          number: 2
  guard
    true
  statements
    :=
      identifier: x
      number: 1
    :=
      identifier: y
      number: 2
  guard
    true
  statements
    :=
      identifier: x
      number: 1
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
if a {
    print(1)
} elif b {
	print(2)
}
`
	expectedOutput = `
if
  guard
    identifier: a
  statements
    identifier: print
      funccall
        number: 1
  guard
    identifier: b
  statements
    identifier: print
      funccall
        number: 2
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
if a {
    print(1)
} else {
	print(2)
}
`
	expectedOutput = `
if
  guard
    identifier: a
  statements
    identifier: print
      funccall
        number: 1
  guard
    true
  statements
    identifier: print
      funccall
        number: 2
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	// Test error output

	input = `else { b }`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (<ELSE>) (Line:1 Pos:1)" {
		t.Error(err)
		return
	}

	input = `elif { b }`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (<ELIF>) (Line:1 Pos:1)" {
		t.Error(err)
		return
	}

	input = `if { b }`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Unexpected end (Line:1 Pos:8)" {
		t.Error(err)
		return
	}

	input = `if == { b }`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:1 Pos:4)" {
		t.Error(err)
		return
	}

	input = `if x { b } elif == { c }`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:1 Pos:17)" {
		t.Error(err)
		return
	}
}
