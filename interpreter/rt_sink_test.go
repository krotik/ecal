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
	"testing"

	"devt.de/krotik/ecal/scope"
)

func TestEventProcessing(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
/*
My cool rule
*/
sink rule1
    kindmatch [ "core.*", "foo.*" ],
	scopematch [ "data.write" ],
	statematch { "val" : NULL },
	priority 10,
	suppresses [ "rule2" ]
	{
        log("rule1 < ", event)
	}
`, vs,
		`
sink # 
My cool rule

  identifier: rule1
  kindmatch
    list
      string: 'core.*'
      string: 'foo.*'
  scopematch
    list
      string: 'data.write'
  statematch
    map
      kvp
        string: 'val'
        null
  priority
    number: 10
  suppresses
    list
      string: 'rule2'
  statements
    identifier: log
      funccall
        string: 'rule1 < '
        identifier: event
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	// Nothing defined in the global scope

	if vs.String() != `
GlobalScope {
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := fmt.Sprint(testprocessor.Rules()["rule1"]); res !=
		`Rule:rule1 [My cool rule] (Priority:10 Kind:[core.* foo.*] Scope:[data.write] StateMatch:{"val":null} Suppress:[rule2])` {
		t.Error("Unexpected result:", res)
		return
	}

	// Test case 1 - Multiple rules, scope match, priorities and waiting for finish (no errors)

	_, err = UnitTestEval(
		`
sink rule1
    kindmatch [ "web.page.index" ],
	scopematch [ "request.read" ],
	{
        log("rule1 - Handling request: ", event.kind)
        addEvent("Rule1Event1", "not_existing", event.state)
        addEvent("Rule1Event2", "web.log", event.state)
        addEvent("Rule1Event3", "notexisting", event.state, {})
	}

sink rule2
    kindmatch [ "web.page.*" ],
    priority 1,  # Ensure this rule is always executed after rule1
	{
        log("rule2 - Tracking user:", event.state.user)
        if event.state.user == "bar" {
            raise("UserBarWasHere", "User bar was seen", [123])
        }
	}

sink rule3
    kindmatch [ "web.log" ],
	{
        log("rule3 - Logging user:", event.state.user)
        return 123
	}

res := addEventAndWait("request", "web.page.index", {
	"user" : "foo"
}, {
	"request.read" : true
})

log("ErrorResult:", res, " ", len(res) == 0)

res := addEventAndWait("request", "web.page.index", {
	"user" : "bar"
}, {
	"request.read" : false
})
log("ErrorResult:", res, " ", res == null)
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
rule1 - Handling request: web.page.index
rule2 - Tracking user:foo
rule3 - Logging user:foo
ErrorResult:[
  {
    "errors": {
      "rule3": {
        "data": 123,
        "detail": "Return value: 123",
        "error": "ECAL error in ECALTestRuntime: *** return *** (Return value: 123) (Line:26 Pos:9)",
        "type": "*** return ***"
      }
    },
    "event": {
      "kind": "web.log",
      "name": "Rule1Event2",
      "state": {
        "user": "foo"
      }
    }
  }
] false
rule2 - Tracking user:bar
ErrorResult:[
  {
    "errors": {
      "rule2": {
        "data": [
          123
        ],
        "detail": "User bar was seen",
        "error": "ECAL error in ECALTestRuntime: UserBarWasHere (User bar was seen) (Line:18 Pos:13)",
        "type": "UserBarWasHere"
      }
    },
    "event": {
      "kind": "web.page.index",
      "name": "request",
      "state": {
        "user": "bar"
      }
    }
  }
] false`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	// Test case 2 - unexpected error

	_, err = UnitTestEval(
		`
sink rule1
    kindmatch [ "test" ],
    {
        log("rule1 - ", event.kind)
        noexitingfunctioncall()
    }

err := addEventAndWait("someevent", "test", {})

if err != null {
    error(err[0].errors)
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
rule1 - test
error: {
  "rule1": {
    "data": null,
    "detail": "Unknown function: noexitingfunctioncall",
    "error": "ECAL error in ECALTestRuntime: Unknown construct (Unknown function: noexitingfunctioncall) (Line:6 Pos:9)",
    "type": "Unknown construct"
  }
}`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	// Test case 3 - rule suppression

	_, err = UnitTestEval(
		`
sink rule1
    kindmatch [ "test.event" ],
    suppresses [ "rule3" ],
	{
        log("rule1 - Handling request: ", event.kind)
	}

sink rule2
    kindmatch [ "test.*" ],
    priority 1,  # Ensure this rule is always executed after rule1
	{
        log("rule2 - Handling request: ", event.kind)
	}

sink rule3
    kindmatch [ "test.*" ],
    priority 1,  # Ensure this rule is always executed after rule1
	{
        log("rule3 - Handling request: ", event.kind)
	}

err := addEventAndWait("myevent", "test.event", {})

if len(err) > 0 {
    error(err[0].errors)
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
rule1 - Handling request: test.event
rule2 - Handling request: test.event`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}

	// Test case 4 - state match

	_, err = UnitTestEval(
		`
sink rule1
    kindmatch [ "test.event", "foo.*" ],
    statematch { "a" : null },
	{
        log("rule1 - Handling request: ", event.kind)
	}

sink rule2
    kindmatch [ "test.*" ],
    priority 1,
    statematch { "b" : 1 },
	{
        log("rule2 - Handling request: ", event.kind)
	}

sink rule3
    kindmatch [ "test.*" ],
    priority 2,
    statematch { "c" : 2 },
	{
        log("rule3 - Handling request: ", event.kind)
	}

err := addEventAndWait("myevent", "test.event", {
	"a" : "foo",
	"b" : 1,
})

if len(err) > 0 {
    error(err[0].errors)
}
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
rule1 - Handling request: test.event
rule2 - Handling request: test.event`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}
}

func TestSinkErrorConditions(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEval(
		`
sink test
    apa
    kindmatch [ "test.event", "foo.*" ],
    statematch { "a" : null },
	{
        log("rule1 - Handling request: ", event.kind)
	}
`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Invalid construct (Unknown expression in sink declaration apa) (Line:3 Pos:5)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`
sink test
    kindmatch 1,
    statematch { "a" : null },
	{
        log("rule1 - Handling request: ", event.kind)
	}
`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Invalid construct (Expected a list as value) (Line:3 Pos:5)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`
sink test
    statematch { "a" : null },
	{
        log("rule1 - Handling request: ", event.kind)
	}
`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Invalid state (Cannot add rule without a kind match: test) (Line:2 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`
sink test
    priority "Hans",
	{
        log("rule1 - Handling request: ", event.kind)
	}
`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Invalid construct (Expected a number as value) (Line:3 Pos:5)" {
		t.Error("Unexpected result:", err)
		return
	}

	_, err = UnitTestEval(
		`
sink test
    statematch "Hans",
	{
        log("rule1 - Handling request: ", event.kind)
	}
`, vs)

	if err == nil || err.Error() != "ECAL error in ECALTestRuntime: Invalid construct (Expected a map as value) (Line:3 Pos:5)" {
		t.Error("Unexpected result:", err)
		return
	}

}
