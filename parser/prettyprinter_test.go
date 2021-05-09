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
	if err == nil || err.Error() != "Nil pointer in AST" {
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

func TestSpecialCasePrinting1(t *testing.T) {
	input := `a := {"a":1,"b":1,"c":1,"d"  :  1,  "e":1,"f":1,"g":1,"h":1,}`

	if err := UnitTestPrettyPrinting(input, "",
		`a := {
    "a" : 1,
    "b" : 1,
    "c" : 1,
    "d" : 1,
    "e" : 1,
    "f" : 1,
    "g" : 1,
    "h" : 1
}`); err != nil {
		t.Error(err)
		return
	}

	input = `a := {"a":1,"b":1,"c":1,"d"  :  {"a":1,"b":{"a":1,"b":1,"c":1,"d":1},"c":1,"d"  :  1,  "e":1,"f":{"a":1,"b":1},"g":1,"h":1,},  "e":1,"f":1,"g":1,"h":1,}`

	if err := UnitTestPrettyPrinting(input, "",
		`a := {
    "a" : 1,
    "b" : 1,
    "c" : 1,
    "d" : {
        "a" : 1,
        "b" : {
            "a" : 1,
            "b" : 1,
            "c" : 1,
            "d" : 1
        },
        "c" : 1,
        "d" : 1,
        "e" : 1,
        "f" : {"a" : 1, "b" : 1},
        "g" : 1,
        "h" : 1
    },
    "e" : 1,
    "f" : 1,
    "g" : 1,
    "h" : 1
}`); err != nil {
		t.Error(err)
		return
	}

	input = `a := [1,2,3,[1,2,[1,2],3,[1,2,3,4],[1,2,3,4,5],4,5],4,5]`

	if err := UnitTestPrettyPrinting(input, "",
		`a := [
    1,
    2,
    3,
    [
        1,
        2,
        [1, 2],
        3,
        [1, 2, 3, 4],
        [
            1,
            2,
            3,
            4,
            5
        ],
        4,
        5
    ],
    4,
    5
]`); err != nil {
		t.Error(err)
		return
	}

	input = `a := [1,2,3,[1,2,{"a":1,"b":1,"c":1,"d":1},3,[1,2,3,4],4,5],4,5]`

	if err := UnitTestPrettyPrinting(input, "",
		`a := [
    1,
    2,
    3,
    [
        1,
        2,
        {
            "a" : 1,
            "b" : 1,
            "c" : 1,
            "d" : 1
        },
        3,
        [1, 2, 3, 4],
        4,
        5
    ],
    4,
    5
]`); err != nil {
		t.Error(err)
		return
	}

}

func TestSpecialCasePrinting2(t *testing.T) {
	input := `
a := 1
a := 2
sink RegisterNewPlayer
  kindmatch   ["foo",2]
  statematch {"a":1,"b":1,"c":1,"d":1}
scopematch []
suppresses ["abs"]
priority 0
{
log("1223")
log("1223")
func foo (z=[1,2,3,4,5]) {
a := 1
b := 2
}
log("1223")
try {
x := [1,2,3,4]
    raise("test 12", null, [1,2,3])
} except e {
p := 1
}
}
`

	if err := UnitTestPrettyPrinting(input, "",
		`a := 1
a := 2
sink RegisterNewPlayer
    kindmatch ["foo", 2]
    statematch {
        "a" : 1,
        "b" : 1,
        "c" : 1,
        "d" : 1
    }
    scopematch []
    suppresses ["abs"]
    priority 0
{
    log("1223")
    log("1223")
    func foo(z=[
        1,
        2,
        3,
        4,
        5
    ]) {
        a := 1
        b := 2
    }
    log("1223")
    try {
        x := [1, 2, 3, 4]
        raise("test 12", null, [1, 2, 3])
    } except e {
        p := 1
    }
}`); err != nil {
		t.Error(err)
	}

	input = `
	/*

	Some initial comment
	
	bla
	*/
a := 1
func aaa() {
mutex myresource {
  globalResource := "new value"
}
func myfunc(a, b, c=1) {
  a := 1 + 1 # Test
}
x := [ 1,2,3,4,5]
a:=1;b:=1
/*Foo*/
Foo := {
  "super" : [ Bar ]
  
  /* 
   * Object IDs
   */
  "id" : 0 # aaaa
  "idx" : 0

/*Constructor*/
  "init" : func(id) 
{ 
    super[0]()
    this.id := id
  }

  /* 
  Return the object ID
  */
  "getId" : func() {
      return this.idx
  }

  /* 
    Set the object ID
  */
  "setId" : func(id) {
      this.idx := id
  }
}
for a in range(2, 10, 2) {
	a := 1
}
for a > 0 {
  a := 1
}
if a == 1 {
    a := a + 1
} elif a == 2 {
    a := a + 2
} else {
    a := 99
}
try {
    raise("MyError", "My error message", [1,2,3])
} except "MyError" as e {
    log(e)
}
}
b:=1
`

	if err := UnitTestPrettyPrinting(input, "",
		`/*

 Some initial comment

 bla

*/
a := 1
func aaa() {
    mutex myresource {
        globalResource := "new value"
    }

    func myfunc(a, b, c=1) {
        a := 1 + 1 # Test
    }
    x := [
        1,
        2,
        3,
        4,
        5
    ]
    a := 1
    b := 1

    /* Foo */
    Foo := {
        "super" : [Bar],

        /*
         * Object IDs
         */
        "id" : 0 # aaaa,
        "idx" : 0,

        /* Constructor */
        "init" : func (id) {
            super[0]()
            this.id := id
        },

        /*
         Return the object ID
         */
        "getId" : func () {
            return this.idx
        },

        /*
         Set the object ID
         */
        "setId" : func (id) {
            this.idx := id
        }
    }
    for a in range(2, 10, 2) {
        a := 1
    }
    for a > 0 {
        a := 1
    }
    if a == 1 {
        a := a + 1
    } elif a == 2 {
        a := a + 2
    } else {
        a := 99
    }
    try {
        raise("MyError", "My error message", [1, 2, 3])
    } except "MyError" as e {
        log(e)
    }
}
b := 1`); err != nil {
		t.Error(err)
		return
	}
}

func TestSpacing(t *testing.T) {
	input := `
	
	import "./templates.ecal" as templates
	
	
a := 1
a := 2


/*
  SomeSink
*/
sink SomeSink
  kindmatch   ["foo",2]
  statematch {"a":1,"b":1,"c":1,"d":1}
scopematch []
suppresses ["abs"]
priority 0
{
log("1223")
log("1223")

t = r"Foo bar
    {{1+2}}   
    aaa"


func foo (z=[1,2,3,4,5]) {
a := 1
b := 2

c := 1
d := 1
}
log("1223")
try {
x := [1,2,3,4]
    raise("test 12", null, [1,2,3])
} except e {
p := 1
}
}
`

	if err := UnitTestPrettyPrinting(input, "",
		`import "./templates.ecal" as templates

a := 1
a := 2

/*
 SomeSink
*/
sink SomeSink
    kindmatch ["foo", 2]
    statematch {
        "a" : 1,
        "b" : 1,
        "c" : 1,
        "d" : 1
    }
    scopematch []
    suppresses ["abs"]
    priority 0
{
    log("1223")
    log("1223")

    t="Foo bar\n    {{1+2}}   \n    aaa"

    func foo(z=[
        1,
        2,
        3,
        4,
        5
    ]) {
        a := 1
        b := 2

        c := 1
        d := 1
    }
    log("1223")
    try {
        x := [1, 2, 3, 4]
        raise("test 12", null, [1, 2, 3])
    } except e {
        p := 1
    }
}`); err != nil {
		t.Error(err)
		return
	}

	input = `

	for [a,b,c] in foo {
a := 1
a := 2

}
`
	if err := UnitTestPrettyPrinting(input, "",
		`for [a, b, c] in foo {
    a := 1
    a := 2
}`); err != nil {
		t.Error(err)
		return
	}
}
