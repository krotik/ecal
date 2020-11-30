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

func TestSimpleExpressionParsing(t *testing.T) {

	// Test error output

	input := `"bl\*a"conversion`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Lexical error (invalid syntax while parsing string) (Line:1 Pos:1)" {
		t.Error(err)
		return
	}

	// Test incomplete expression

	input = `a *`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Unexpected end" {
		t.Error(err)
		return
	}

	input = `not ==`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:1 Pos:5)" {
		t.Error(err)
		return
	}

	input = `(==)`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:1 Pos:2)" {
		t.Error(err)
		return
	}

	input = "5 ( 5"
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term can only start an expression (() (Line:1 Pos:3)" {
		t.Error(err)
		return
	}

	input = "5 + \""
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Lexical error (Unexpected end while reading string value (unclosed quotes)) (Line:1 Pos:5)" {
		t.Error(err)
		return
	}

	// Test prefix operator

	input = ` + a - -5`
	expectedOutput := `
minus
  plus
    identifier: a
  minus
    number: 5
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

}

func TestArithmeticParsing(t *testing.T) {
	input := "a + b * 5 /2"
	expectedOutput := `
plus
  identifier: a
  div
    times
      identifier: b
      number: 5
    number: 2
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	// Test brackets

	input = "a + 1 * (5 + 6)"
	expectedOutput = `
plus
  identifier: a
  times
    number: 1
    plus
      number: 5
      number: 6
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	// Test needless brackets

	input = "(a + 1) * (5 / (6 - 2))"
	expectedOutput = `
times
  plus
    identifier: a
    number: 1
  div
    number: 5
    minus
      number: 6
      number: 2
`[1:]

	// Pretty printer should get rid of the needless brackets

	res, err := UnitTestParseWithPPResult("mytest", input, "(a + 1) * 5 / (6 - 2)")
	if err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestLogicParsing(t *testing.T) {
	input := "not (a + 1) * 5 and tRue == false or not 1 - 5 != test"
	expectedOutput := `
or
  and
    not
      times
        plus
          identifier: a
          number: 1
        number: 5
    ==
      true
      false
  not
    !=
      minus
        number: 1
        number: 5
      identifier: test
`[1:]

	res, err := UnitTestParseWithPPResult("mytest", input, "not (a + 1) * 5 and true == false or not 1 - 5 != test")

	if err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = "a > b or a <= p or b hasSuffix 'test' or c hasPrefix 'test' and x < 4 or x >= 10"
	expectedOutput = `
or
  or
    or
      or
        >
          identifier: a
          identifier: b
        <=
          identifier: a
          identifier: p
      hassuffix
        identifier: b
        string: 'test'
    and
      hasprefix
        identifier: c
        string: 'test'
      <
        identifier: x
        number: 4
  >=
    identifier: x
    number: 10
`[1:]

	res, err = UnitTestParseWithPPResult("mytest", input, `a > b or a <= p or b hassuffix "test" or c hasprefix "test" and x < 4 or x >= 10`)

	if err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = "(a in null or c notin d) and false like 9 or x // 6 > 2 % 1"
	expectedOutput = `
or
  and
    or
      in
        identifier: a
        null
      notin
        identifier: c
        identifier: d
    like
      false
      number: 9
  >
    divint
      identifier: x
      number: 6
    modint
      number: 2
      number: 1
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestCompositionStructureParsing(t *testing.T) {

	// Assignment of map

	input := `x := { z : "foo", y : "bar", z : "zzz" }`
	expectedOutput := `
:=
  identifier: x
  map
    kvp
      identifier: z
      string: 'foo'
    kvp
      identifier: y
      string: 'bar'
    kvp
      identifier: z
      string: 'zzz'
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `x := { ==`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Term cannot start an expression (==) (Line:1 Pos:8)" {
		t.Error(err)
		return
	}

	// Statement separator

	input = `print(123); x := { z : "foo", y : "bar", z : "zzz" }; foo := y == 1`
	expectedOutput = `
statements
  identifier: print
    funccall
      number: 123
  :=
    identifier: x
    map
      kvp
        identifier: z
        string: 'foo'
      kvp
        identifier: y
        string: 'bar'
      kvp
        identifier: z
        string: 'zzz'
  :=
    identifier: foo
    ==
      identifier: y
      number: 1
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `print(123); x := { z : "foo", y : "bar", z : "zzz" }; foo £ y == 1`
	if _, err := UnitTestParse("mytest", input); err.Error() !=
		"Parse error in mytest: Lexical error (Cannot parse identifier '£'. Identifies may only contain [a-zA-Z] and [a-zA-Z0-9] from the second character) (Line:1 Pos:59)" {
		t.Error(err)
		return
	}

	input = `x := [1,2]
[a,b] := x`
	expectedOutput = `
statements
  :=
    identifier: x
    list
      number: 1
      number: 2
  :=
    list
      identifier: a
      identifier: b
    identifier: x
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `x := [1,2];[a,b] := x`
	expectedOutput = `
statements
  :=
    identifier: x
    list
      number: 1
      number: 2
  :=
    list
      identifier: a
      identifier: b
    identifier: x
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}
