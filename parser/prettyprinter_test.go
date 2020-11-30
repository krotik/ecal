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
	"testing"
)

func TestErrorHandling(t *testing.T) {

	input := "c:= a + b"

	astres, err := ParseWithRuntime("mytest", input, &DummyRuntimeProvider{})
	if err != nil {
		t.Errorf("Unexpected parser output:\n%vError: %v", astres, err)
		return
	}

	// Make ast invalid

	astres.Children[1].Children[1] = nil

	ppres, err := PrettyPrint(astres)
	if err == nil || err.Error() != "Nil pointer in AST at level: 2" {
		t.Errorf("Unexpected result: %v error: %v", ppres, err)
		return
	}
}

func TestArithmeticExpressionPrinting(t *testing.T) {

	input := "a + b * 5 /2-1"
	expectedOutput := `
minus
  plus
    identifier: a
    div
      times
        identifier: b
        number: 5
      number: 2
  number: 1
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		"a + b * 5 / 2 - 1"); err != nil {
		t.Error(err)
		return
	}

	input = `-a + "\"'b"`
	expectedOutput = `
plus
  minus
    identifier: a
  string: '"'b'
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		`-a + "\"'b"`); err != nil {
		t.Error(err)
		return
	}

	input = `a // 5 % (50 + 1)`
	expectedOutput = `
modint
  divint
    identifier: a
    number: 5
  plus
    number: 50
    number: 1
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		`a // 5 % (50 + 1)`); err != nil {
		t.Error(err)
		return
	}

	input = "(a + 1) * 5 / (6 - 2)"
	expectedOutput = `
div
  times
    plus
      identifier: a
      number: 1
    number: 5
  minus
    number: 6
    number: 2
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		"(a + 1) * 5 / (6 - 2)"); err != nil {
		t.Error(err)
		return
	}

	input = "a + (1 * 5) / 6 - 2"
	expectedOutput = `
minus
  plus
    identifier: a
    div
      times
        number: 1
        number: 5
      number: 6
  number: 2
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		"a + 1 * 5 / 6 - 2"); err != nil {
		t.Error(err)
		return
	}
}

func TestLogicalExpressionPrinting(t *testing.T) {
	input := "not (a + 1) * 5 and tRue or not 1 - 5 != '!test'"
	expectedOutput := `
or
  and
    not
      times
        plus
          identifier: a
          number: 1
        number: 5
    true
  not
    !=
      minus
        number: 1
        number: 5
      string: '!test'
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		"not (a + 1) * 5 and true or not 1 - 5 != \"!test\""); err != nil {
		t.Error(err)
		return
	}

	input = "not x < null and a > b or 1 <= c and 2 >= false or c == true"
	expectedOutput = `
or
  or
    and
      not
        <
          identifier: x
          null
      >
        identifier: a
        identifier: b
    and
      <=
        number: 1
        identifier: c
      >=
        number: 2
        false
  ==
    identifier: c
    true
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		"not x < null and a > b or 1 <= c and 2 >= false or c == true"); err != nil {
		t.Error(err)
		return
	}

	input = "a hasPrefix 'a' and b hassuffix 'c' or d like '^.*' and 3 notin x"
	expectedOutput = `
or
  and
    hasprefix
      identifier: a
      string: 'a'
    hassuffix
      identifier: b
      string: 'c'
  and
    like
      identifier: d
      string: '^.*'
    notin
      number: 3
      identifier: x
`[1:]

	if err := UnitTestPrettyPrinting(input, expectedOutput,
		`a hasprefix "a" and b hassuffix "c" or d like "^.*" and 3 notin x`); err != nil {
		t.Error(err)
		return
	}
}
