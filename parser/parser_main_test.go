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

func TestStatementParsing(t *testing.T) {

	// Comment parsing without statements

	input := `a := 1
	b := 2; c:= 3`
	expectedOutput := `
statements
  :=
    identifier: a
    number: 1
  :=
    identifier: b
    number: 2
  :=
    identifier: c
    number: 3
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestIdentifierParsing(t *testing.T) {

	input := `a := 1
	a.foo := 2
	a.b.c.foo := a.b
	`
	expectedOutput := `
statements
  :=
    identifier: a
    number: 1
  :=
    identifier: a
      identifier: foo
    number: 2
  :=
    identifier: a
      identifier: b
        identifier: c
          identifier: foo
    identifier: a
      identifier: b
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `a := b[1 + 1]
	a[4].foo["aaa"] := c[i]
	`
	expectedOutput = `
statements
  :=
    identifier: a
    identifier: b
      compaccess
        plus
          number: 1
          number: 1
  :=
    identifier: a
      compaccess
        number: 4
      identifier: foo
        compaccess
          string: 'aaa'
    identifier: c
      compaccess
        identifier: i
`[1:]
	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestCommentParsing(t *testing.T) {

	// Comment parsing without statements

	input := `/* This is  a comment*/ a := 1 + 1 # foo bar`
	expectedOutput := `
:=
  identifier: a #  This is  a comment
  plus
    number: 1
    number: 1 #  foo bar
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `/* foo */ 1 # foo bar`
	expectedOutput = `
number: 1 #  foo   foo bar
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `# 123 
/* foo */ 1 # foo bar`
	expectedOutput = `
number: 1 #  foo   foo bar
`[1:]

	if res, err := UnitTestParse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestErrorConditions(t *testing.T) {

	input := ``
	if ast, err := Parse("test", input); err == nil || err.Error() != "Parse error in test: Unexpected end" {
		t.Errorf("Unexpected result: %v\nAST:\n%v", err, ast)
		return
	}

	input = `a := 1 a`
	if ast, err := Parse("test", input); err == nil || err.Error() != `Parse error in test: Unexpected end (extra token id:7 ("a")) (Line:1 Pos:8)` {
		t.Errorf("Unexpected result: %v\nAST:\n%v", err, ast)
		return
	}

	tokenStringEntry := astNodeMap[TokenSTRING]
	delete(astNodeMap, TokenSTRING)
	defer func() {
		astNodeMap[TokenSTRING] = tokenStringEntry
	}()

	input = `"foo"`
	if ast, err := Parse("test", input); err == nil || err.Error() != `Parse error in test: Unknown term (id:5 (v:"foo")) (Line:1 Pos:1)` {
		t.Errorf("Unexpected result: %v\nAST:\n%v", err, ast)
		return
	}

	// Test parser functions

	input = `a := 1 + a`

	p := &parser{"test", nil, NewLABuffer(Lex("test", input), 3), nil}
	node, _ := p.next()
	p.node = node

	if err := skipToken(p, TokenAND); err == nil || err.Error() != "Parse error in test: Unexpected term (a) (Line:1 Pos:1)" {
		t.Errorf("Unexpected result: %v", err)
		return
	}

	if err := acceptChild(p, node, TokenAND); err == nil || err.Error() != "Parse error in test: Unexpected term (a) (Line:1 Pos:1)" {
		t.Errorf("Unexpected result: %v", err)
		return
	}
}
