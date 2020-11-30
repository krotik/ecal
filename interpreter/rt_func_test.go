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

	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/ecal/scope"
)

func TestFunctions(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	res, err := UnitTestEvalAndAST(`
foo := [ [ func (a, b, c=1) {
    return a + b + c
  }
]]

result1 := foo[0][0](3, 2)
`, vs, `
statements
  :=
    identifier: foo
    list
      list
        function
          params
            identifier: a
            identifier: b
            preset
              identifier: c
              number: 1
          statements
            return
              plus
                plus
                  identifier: a
                  identifier: b
                identifier: c
  :=
    identifier: result1
    identifier: foo
      compaccess
        number: 0
      compaccess
        number: 0
      funccall
        number: 3
        number: 2
`[1:])

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    foo ([]interface {}) : [["ecal.function:  (Line 2, Pos 12)"]]
    result1 (float64) : 6
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)

	res, err = UnitTestEval(`
b := "a"
foo := [{
  b : func (a, b, c=1) {
    return [1,[a + b + c]]
  }
}]

result1 := foo[0].a(3, 2)[1][0]
`, vs)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    b (string) : a
    foo ([]interface {}) : [{"a":"ecal.function:  (Line 4, Pos 7)"}]
    result1 (float64) : 6
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)

	res, err = UnitTestEval(`
b := "a"
foo := {
  b : [func (a, b, c=1) {
    return { "x" : { "y" : [a + b + c] }}
  }]
}

result1 := foo.a[0](3, 2).x.y[0]
`, vs)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    b (string) : a
    foo (map[interface {}]interface {}) : {"a":["ecal.function:  (Line 4, Pos 8)"]}
    result1 (float64) : 6
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)

	res, err = UnitTestEvalAndAST(`
foo := {
  "a" : {
	"b" : func myfunc(a, b, c=1) {
	      d := a + b + c
	      return d
	    }
	}
}
result1 := foo.a.b(3, 2)
result2 := myfunc(3, 3)
`, vs, `
statements
  :=
    identifier: foo
    map
      kvp
        string: 'a'
        map
          kvp
            string: 'b'
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
                  identifier: d
                  plus
                    plus
                      identifier: a
                      identifier: b
                    identifier: c
                return
                  identifier: d
  :=
    identifier: result1
    identifier: foo
      identifier: a
        identifier: b
          funccall
            number: 3
            number: 2
  :=
    identifier: result2
    identifier: myfunc
      funccall
        number: 3
        number: 3
`[1:])

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    foo (map[interface {}]interface {}) : {"a":{"b":"ecal.function: myfunc (Line 4, Pos 8)"}}
    myfunc (*interpreter.function) : ecal.function: myfunc (Line 4, Pos 8)
    result1 (float64) : 6
    result2 (float64) : 7
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}
}

func TestFunctionScoping(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	res, err := UnitTestEval(`
c := 1
foo := func (a, b=1) {
  return a + b + c
}

result1 := foo(3, 2)
`, vs)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    c (float64) : 1
    foo (*interpreter.function) : ecal.function:  (Line 3, Pos 8)
    result1 (float64) : 6
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	vs = scope.NewScope(scope.GlobalScope)

	res, err = UnitTestEval(`
func fib(n) {
  if (n <= 1) {
      return n
  }
  return fib(n-1) + fib(n-2)
}

result1 := fib(12)
`, vs)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    fib (*interpreter.function) : ecal.function: fib (Line 2, Pos 1)
    result1 (float64) : 144
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}
}

func TestObjectInstantiation(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	res, err := UnitTestEvalAndAST(`
Super := {
  "name" : "base"

  "init" : func() {
    this.name := "baseclass"
  }
}

Bar := {
  "super" : [ Super ]

  "test" : ""

  "init" : func(test) {
    this.test := test
    x := super[0]()
  }
}

Bar2 := {
  "getTest" : func() {
      return this.test
  }
}

Foo := {
  "super" : [ Bar, Bar2 ]

  # Object ID
  #
  "id" : 0

  "idx" : 0

  # Constructor
  #
  "init" : func(id, test) {
    this.id := id
    x := super[0](test)
  }

  # Return the object ID
  #
  "getId" : func() {
      return this.idx
  }

  # Set the object ID
  #
  "setId" : func(id) {
      this.idx := id
  }
}

result1 := new(Foo, 123, "tester")
result1.setId(500)
result2 := result1.getId() + result1.id
`, vs, "")

	if err == nil {
		v, _, _ := vs.GetValue("result1")
		if res := stringutil.ConvertToPrettyString(v); res != `{
  "getId": "ecal.function:  (Line 45, Pos 42)",
  "getTest": "ecal.function:  (Line 22, Pos 15)",
  "id": 123,
  "idx": 500,
  "init": "ecal.function:  (Line 38, Pos 32)",
  "name": "baseclass",
  "setId": "ecal.function:  (Line 51, Pos 39)",
  "super": [
    {
      "init": "ecal.function:  (Line 15, Pos 12)",
      "super": [
        {
          "init": "ecal.function:  (Line 5, Pos 12)",
          "name": "base"
        }
      ],
      "test": ""
    },
    {
      "getTest": "ecal.function:  (Line 22, Pos 15)"
    }
  ],
  "test": "tester"
}` {
			t.Error("Unexpected result: ", res)
			return
		}
	}

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    Bar (map[interface {}]interface {}) : {"init":"ecal.function:  (Line 15, Pos 12)","super":[{"init":"ecal.function:  (Line 5, Pos 12)","name":"base"}],"test":""}
    Bar2 (map[interface {}]interface {}) : {"getTest":"ecal.function:  (Line 22, Pos 15)"}
    Foo (map[interface {}]interface {}) : {"getId":"ecal.function:  (Line 45, Pos 42)","id":0,"idx":0,"init":"ecal.function:  (Line 38, Pos 32)","setId":"ecal.function:  (Line 51, Pos 39)","super":[{"init":"ecal.function:  (Line 15, Pos 12)","super":[{"init":"ecal.function:  (Line 5, Pos 12)","name":"base"}],"test":""},{"getTest":"ecal.function:  (Line 22, Pos 15)"}]}
    Super (map[interface {}]interface {}) : {"init":"ecal.function:  (Line 5, Pos 12)","name":"base"}
    result1 (map[interface {}]interface {}) : {"getId":"ecal.function:  (Line 45, Pos 42)","getTest":"ecal.function:  (Line 22, Pos 15)","id":123,"idx":500,"init":"ecal.function:  (Line 38, Pos 32)","name":"baseclass","setId":"ecal.function:  (Line 51, Pos 39)","super":[{"init":"ecal.function:  (Line 15, Pos 12)","super":[{"init":"ecal.function:  (Line 5, Pos 12)","name":"base"}],"test":""},{"getTest":"ecal.function:  (Line 22, Pos 15)"}],"test":"tester"}
    result2 (float64) : 623
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}

	_, err = UnitTestEval(`
Bar := {
  "super" : "hans"

}


result1 := new(Bar)
`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Runtime error (Property _super must be a list of super classes) (Line:8 Pos:12)" {
		t.Error("Unexpected result:", err)
		return
	}

}
