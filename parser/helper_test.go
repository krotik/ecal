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

func TestASTNode(t *testing.T) {

	n, err := ParseWithRuntime("test1", "- 1", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	n2, err := ParseWithRuntime("test2", "-2", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	if ok, msg := n.Equals(n2, false); ok || msg != `Path to difference: minus > number

Token is different:
Pos is different 2 vs 1
Val is different 1 vs 2
Lpos is different 3 vs 2
{
  "ID": 6,
  "Pos": 2,
  "Val": "1",
  "Identifier": false,
  "AllowEscapes": false,
  "Lsource": "test1",
  "Lline": 1,
  "Lpos": 3
}
vs
{
  "ID": 6,
  "Pos": 1,
  "Val": "2",
  "Identifier": false,
  "AllowEscapes": false,
  "Lsource": "test2",
  "Lline": 1,
  "Lpos": 2
}

AST Nodes:
number: 1
vs
number: 2
` {
		t.Error("Unexpected result: ", msg)
		return
	}

	n, err = ParseWithRuntime("test1", "-1", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	n2, err = ParseWithRuntime("test2", "-a", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	if ok, msg := n.Equals(n2, true); ok || msg != `Path to difference: minus > number

Name is different number vs identifier
Token is different:
ID is different 6 vs 7
Val is different 1 vs a
Identifier is different false vs true
{
  "ID": 6,
  "Pos": 1,
  "Val": "1",
  "Identifier": false,
  "AllowEscapes": false,
  "Lsource": "test1",
  "Lline": 1,
  "Lpos": 2
}
vs
{
  "ID": 7,
  "Pos": 1,
  "Val": "a",
  "Identifier": true,
  "AllowEscapes": false,
  "Lsource": "test2",
  "Lline": 1,
  "Lpos": 2
}

AST Nodes:
number: 1
vs
identifier: a
` {
		t.Error("Unexpected result: ", msg)
		return
	}

	n, err = ParseWithRuntime("", "- 1", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	n2, err = ParseWithRuntime("", "a - b", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	if ok, msg := n.Equals(n2, false); ok || msg != `Path to difference: minus

Number of children is different 1 vs 2

AST Nodes:
minus
  number: 1
vs
minus
  identifier: a
  identifier: b
` {
		t.Error("Unexpected result: ", msg)
		return
	}

	n, err = ParseWithRuntime("", "-1 #test", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	n2, err = ParseWithRuntime("", "-1", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	if ok, msg := n.Equals(n2, false); ok || msg != `Path to difference: minus > number

Number of meta data entries is different 1 vs 0

AST Nodes:
number: 1 # test
vs
number: 1
` {
		t.Error("Unexpected result: ", msg)
		return
	}

	n, err = ParseWithRuntime("", "-1 #test", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	n2, err = ParseWithRuntime("", "-1 #wurst", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	if ok, msg := n.Equals(n2, false); ok || msg != `Path to difference: minus > number

Meta data value is different test vs wurst

AST Nodes:
number: 1 # test
vs
number: 1 # wurst
` {
		t.Error("Unexpected result: ", msg)
		return
	}

	n, err = ParseWithRuntime("", "1 #test", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	n2, err = ParseWithRuntime("", "/*test*/ 1", &DummyRuntimeProvider{})
	if err != nil {
		t.Error("Cannot parse test AST:", err)
		return
	}

	if ok, msg := n.Equals(n2, false); ok || msg != `Path to difference: number

Token is different:
Pos is different 0 vs 9
Lpos is different 1 vs 10
{
  "ID": 6,
  "Pos": 0,
  "Val": "1",
  "Identifier": false,
  "AllowEscapes": false,
  "Lsource": "",
  "Lline": 1,
  "Lpos": 1
}
vs
{
  "ID": 6,
  "Pos": 9,
  "Val": "1",
  "Identifier": false,
  "AllowEscapes": false,
  "Lsource": "",
  "Lline": 1,
  "Lpos": 10
}
Meta data type is different MetaDataPostComment vs MetaDataPreComment

AST Nodes:
number: 1 # test
vs
number: 1 # test
` {
		t.Error("Unexpected result: ", msg)
		return
	}

	// Test building an AST from an invalid

	if _, err := ASTFromJSONObject(map[string]interface{}{
		"value": "foo",
	}); err == nil || err.Error() != "Found json ast node without a name: map[value:foo]" {
		t.Error("Unexpected result: ", err)
		return
	}

	if _, err := ASTFromJSONObject(map[string]interface{}{
		"name": "foo",
		"children": []map[string]interface{}{
			{
				"value": "bar",
			},
		},
	}); err == nil || err.Error() != "Found json ast node without a name: map[value:bar]" {
		t.Error("Unexpected result: ", err)
		return
	}

	// Test population of missing information

	if ast, err := ASTFromJSONObject(map[string]interface{}{
		"name": "foo",
	}); err != nil || ast.String() != "foo\n" || ast.Token.String() != `v:""` {
		t.Error("Unexpected result: ", ast.Token.String(), ast.String(), err)
		return
	}

	if ast, err := ASTFromJSONObject(map[string]interface{}{
		"name": "foo",
		"children": []map[string]interface{}{
			{
				"name": "bar",
			},
		},
	}); err != nil || ast.String() != "foo\n  bar\n" || ast.Token.String() != `v:""` {
		t.Error("Unexpected result: ", ast.Token.String(), ast.String(), err)
		return
	}
}

func TestLABuffer(t *testing.T) {

	buf := NewLABuffer(Lex("test", "1 2 3 4 5 6 7 8 9"), 3)

	if token, ok := buf.Next(); token.Val != "1" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "2" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Check Peek

	if token, ok := buf.Peek(0); token.Val != "3" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(1); token.Val != "4" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(2); token.Val != "5" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(3); token.ID != TokenEOF || ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Continue

	if token, ok := buf.Next(); token.Val != "3" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "4" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "5" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "6" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "7" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "8" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Check Peek

	if token, ok := buf.Peek(0); token.Val != "9" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(1); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(2); token.ID != TokenEOF || ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Continue

	if token, ok := buf.Next(); token.Val != "9" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Check Peek

	if token, ok := buf.Peek(0); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(1); token.ID != TokenEOF || ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Continue

	if token, ok := buf.Next(); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// New Buffer

	buf = NewLABuffer(Lex("test", "1 2 3"), 3)

	if token, ok := buf.Next(); token.Val != "1" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "2" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// Check Peek

	if token, ok := buf.Peek(0); token.Val != "3" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(1); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(2); token.ID != TokenEOF || ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.Val != "3" || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	// New Buffer - test edge case

	buf = NewLABuffer(Lex("test", ""), 0)

	if token, ok := buf.Peek(0); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.ID != TokenEOF || !ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Peek(0); token.ID != TokenEOF || ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}

	if token, ok := buf.Next(); token.ID != TokenEOF || ok {
		t.Error("Unexpected result: ", token, ok)
		return
	}
}
