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

func TestNextItem(t *testing.T) {

	l := &lexer{"Test", "1234", 0, 0, 0, 0, 0, make(chan LexToken)}

	r := l.next(1)

	if r != '1' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(0); r != '1' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(0); r != '2' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(1); r != '3' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(2); r != '4' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(0); r != '3' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(0); r != '4' {
		t.Errorf("Unexpected token: %q", r)
		return
	}

	if r := l.next(0); r != RuneEOF {
		t.Errorf("Unexpected token: %q", r)
		return
	}
}

func TestEquals(t *testing.T) {
	l := LexToList("mytest", "not\n test")

	if mt := l[0].Type(); mt != "MetaDataGeneral" {
		t.Error("Unexpected meta type:", mt)
		return
	}

	if ok, msg := l[0].Equals(l[1], false); ok || msg != `ID is different 54 vs 7
Pos is different 0 vs 5
Val is different not vs test
Identifier is different false vs true
Lline is different 1 vs 2
Lpos is different 1 vs 2
{
  "ID": 54,
  "Pos": 0,
  "Val": "not",
  "Identifier": false,
  "AllowEscapes": false,
  "Lsource": "mytest",
  "Lline": 1,
  "Lpos": 1
}
vs
{
  "ID": 7,
  "Pos": 5,
  "Val": "test",
  "Identifier": true,
  "AllowEscapes": false,
  "Lsource": "mytest",
  "Lline": 2,
  "Lpos": 2
}` {
		t.Error("Unexpected result:", msg)
		return
	}
}

func TestBasicTokenLexing(t *testing.T) {

	// Test empty string parsing

	if res := fmt.Sprint(LexToList("mytest", "    \t   ")); res != "[EOF]" {
		t.Error("Unexpected lexer result:\n  ", res)
		return
	}

	// Test arithmetics

	input := `name := a + 1 and (ver+x!=1) * 5 > name2`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`["name" := "a" + v:"1" <AND> ( "ver" + "x" != v:"1" ) * v:"5" > "name2" EOF]` {
		t.Error("Unexpected lexer result:\n  ", res)
		return
	}

	input = `test := not a * 1.3 or (12 / aa) * 5 DiV 3 % 1 > trUe`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`["test" := <NOT> "a" * v:"1.3" <OR> ( v:"12" / "aa" ) * v:"5" "DiV" v:"3" % v:"1" > <TRUE> EOF]` {
		t.Error("Unexpected lexer result:\n  ", res)
		return
	}

	input = `-1.234560e+02+5+2.123 // 1`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`[- v:"1.234560e+02" + v:"5" + v:"2.123" // v:"1" EOF]` {
		t.Error("Unexpected lexer result:\n  ", res)
		return
	}

	// Test invalid identifier

	input = `5test`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`[v:"5" "test" EOF]` {
		t.Error("Unexpected lexer result:\n  ", res)
		return
	}

	input = `@test`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`[Error: Cannot parse identifier '@test'. Identifies may only contain [a-zA-Z] and [a-zA-Z0-9] from the second character (Line 1, Pos 1) EOF]` {
		t.Error("Unexpected lexer result:\n  ", res)
		return
	}
}

func TestAssignmentLexing(t *testing.T) {

	input := `name := a + 1`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`["name" := "a" + v:"1" EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `name := a.a + a.b`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`["name" := "a" . "a" + "a" . "b" EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `name:=a[1] + b["d"] + c[a]`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`["name" := "a" [ v:"1" ] + "b" [ v:"d" ] + "c" [ "a" ] EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}
}

func TestBlockLexing(t *testing.T) {

	input := `
if a == 1 {
    print("xxx")
} elif b > 2 {
    print("yyy")
} else {
    print("zzz")
}
`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`[<IF> "a" == v:"1" { "print" ( v:"xxx" ) } <ELIF> "b" > v:"2" { "print" ( v:"yyy" ) } <ELSE> { "print" ( v:"zzz" ) } EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `
for a, b in enum(blist) {
    do(a)
}
`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`[<FOR> "a" , "b" <IN> "enum" ( "blist" ) { "do" ( "a" ) } EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `
for true {
	x := "1"
	break; continue
}
`
	if res := LexToList("mytest", input); fmt.Sprint(res) !=
		`[<FOR> <TRUE> { "x" := v:"1" <BREAK> ; <CONTINUE> } EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}
}

func TestStringLexing(t *testing.T) {

	// Test unclosed quotes

	input := `name "test  bla`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `["name" Error: Unexpected end while reading string value (unclosed quotes) (Line 1, Pos 6) EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `name "test"  'bla'`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `["name" v:"test" v:"bla" EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `name "te
	st"  'bla'`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `["name" Error: invalid syntax while parsing string (Line 1, Pos 6)]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `name r"te
	st"  'bla'`
	res := LexToList("mytest", input)
	if fmt.Sprint(res) != `["name" v:"te\n\tst" v:"bla" EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	if res[1].AllowEscapes {
		t.Error("String value should not allow escapes")
		return
	}

	// Parsing with escape sequences

	input = `"test\n\ttest"  '\nfoo\u0028bar' "test{foo}.5w3f"`
	res = LexToList("mytest", input)
	if fmt.Sprint(res) != `[v:"test\n\ttest" v:"\nfoo(bar" v:"test{foo}.5w3f" EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	if !res[0].AllowEscapes {
		t.Error("String value should allow escapes")
		return
	}
}

func TestCommentLexing(t *testing.T) {

	input := `name /* foo
		bar
    x*/ 'b/* - */la' /*test*/`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `["name" /*  foo
		bar
    x */ v:"b/* - */la" /* test */ EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `name /* foo
		bar`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `["name" Error: Unexpected end while reading comment (Line 1, Pos 8) EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `foo
   1+ 2 # Some comment
bar`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `["foo" v:"1" + v:"2" #  Some comment
 "bar" EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `1+ 2 # Some comment`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `[v:"1" + v:"2" #  Some comment EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}

	input = `
/*
Conway's Game of Life

A zero-player game that evolves based on its initial state.

https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life
*/

1+ 2 # Some comment`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `[/* 
Conway's Game of Life

A zero-player game that evolves based on its initial state.

https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life
 */ v:"1" + v:"2" #  Some comment EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}
}

func TestSinkLexing(t *testing.T) {

	input := `sink "mysink"
r"
A comment describing the sink.
"
kindmatch [ foo.bar.* ],
scopematch [ "data.read", "data.write" ],
statematch { a : 1, b : NULL },
priority 0,
suppresses [ "myothersink" ]
{
  a := 1
}`
	if res := LexToList("mytest", input); fmt.Sprint(res) != `[<SINK> v:"mysink" v:"\nA comment"... <KINDMATCH> `+
		`[ "foo" . "bar" . * ] , <SCOPEMATCH> [ v:"data.read" , v:"data.write" ] , <STATEMATCH> `+
		`{ "a" : v:"1" , "b" : <NULL> } , <PRIORITY> v:"0" , <SUPPRESSES> [ v:"myothersink" ] `+
		`{ "a" := v:"1" } EOF]` {
		t.Error("Unexpected lexer result:", res)
		return
	}
}
