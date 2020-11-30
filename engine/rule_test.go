/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package engine

import (
	"fmt"
	"regexp"
	"sort"
	"testing"
)

func TestRuleIndexSimple(t *testing.T) {
	ruleindexidcounter = 0
	defer func() {
		ruleindexidcounter = 0
	}()

	// Store a simple rule

	rule := &Rule{
		"TestRule", // Name
		"",         // Description
		[]string{"core.main.tester", "core.tmp.*"}, // Kind match
		[]string{"data.read", "data.test"},         // Match on event cascade scope
		nil,
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	}

	index := NewRuleIndex()

	index.AddRule(rule)

	// Check error cases

	err := index.AddRule(&Rule{
		"TestRuleError",              // Name
		"",                           // Description
		[]string{"core.main.tester"}, // Kind match
		nil,
		nil,
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	})
	if err.Error() != "Cannot add rule without a scope match: TestRuleError" {
		t.Error("Unexpected result:", err)
		return
	}

	err = index.AddRule(&Rule{
		"TestRuleError2",                   // Name
		"",                                 // Description
		nil,                                // Kind match
		[]string{"data.read", "data.test"}, // Match on event cascade scope
		nil,
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	})
	if err.Error() != "Cannot add rule without a kind match: TestRuleError2" {
		t.Error("Unexpected result:", err)
		return
	}

	// Check index layout

	if res := index.String(); res != `
core - RuleIndexKind (0)
  main - RuleIndexKind (1)
    tester - RuleIndexKind (2)
      RuleIndexAll (3)
        Rule:TestRule [] (Priority:0 Kind:[core.main.tester core.tmp.*] Scope:[data.read data.test] StateMatch:null Suppress:[TestRule66])
  tmp - RuleIndexKind (1)
    * - RuleIndexKind (4)
      RuleIndexAll (5)
        Rule:TestRule [] (Priority:0 Kind:[core.main.tester core.tmp.*] Scope:[data.read data.test] StateMatch:null Suppress:[TestRule66])
`[1:] {
		t.Error("Unexpected index layout:", res)
		return
	}

	// Check trigger queries

	if !index.IsTriggering(&Event{
		"bla",
		[]string{"core", "tmp", "bla"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	if index.IsTriggering(&Event{
		"bla",
		[]string{"core", "tmp"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	if index.IsTriggering(&Event{
		"bla",
		[]string{"core", "tmpp", "bla"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	if !index.IsTriggering(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	if index.IsTriggering(&Event{
		"bla",
		[]string{"core", "main", "tester", "bla"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	if index.IsTriggering(&Event{
		"bla",
		[]string{"core", "main", "teste"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	if index.IsTriggering(&Event{
		"bla",
		[]string{"core", "main"},
		nil,
	}) {
		t.Error("Unexpected result")
		return
	}

	// Event matching

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		nil,
	}); printRules(res) != "[TestRule]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp", "x"},
		nil,
	}); printRules(res) != "[TestRule]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp"},
		nil,
	}); printRules(res) != "[]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp", "x", "y"},
		nil,
	}); printRules(res) != "[]" {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestRuleIndexStateMatch(t *testing.T) {
	ruleindexidcounter = 0
	defer func() {
		ruleindexidcounter = 0
	}()

	rule1 := &Rule{
		"TestRule1", // Name
		"",          // Description
		[]string{"core.main.tester", "core.tmp.*"}, // Kind match
		[]string{"data.read", "data.test"},         // Match on event cascade scope
		map[string]interface{}{ // Match on event state
			"name": nil,
			"test": "val1",
		},
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	}

	rule2 := &Rule{
		"TestRule2",                  // Name
		"",                           // Description
		[]string{"core.main.tester"}, // Kind match
		[]string{"data.read"},        // Match on event cascade scope
		map[string]interface{}{ // Match on event state
			"name":  nil,
			"test":  "val2",
			"test2": 42,
		},
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	}

	rule3 := &Rule{
		"TestRule3",                  // Name
		"",                           // Description
		[]string{"core.main.tester"}, // Kind match
		[]string{"data.read"},        // Match on event cascade scope
		map[string]interface{}{ // Match on event state
			"name":  nil,
			"test":  "val2",
			"test2": 42,
			"test3": 15,
		},
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	}

	index := NewRuleIndex()

	index.AddRule(rule1)
	index.AddRule(rule2)
	index.AddRule(rule3)

	if err := index.AddRule(rule3); err.Error() != "Cannot add rule TestRule3 twice" {
		t.Error("Unexpected result:", err)
		return
	}

	if len(index.Rules()) != 3 {
		t.Error("Unexpected number of rules:", len(index.Rules()))
	}

	// Check index layout

	if res := index.String(); res != `
core - RuleIndexKind (0)
  main - RuleIndexKind (1)
    tester - RuleIndexKind (2)
      RuleIndexState (3) [TestRule1 TestRule2 TestRule3 ]
        name - 00000007 *:00000007 [] []
        test - 00000007 *:00000000 [val1:00000001 val2:00000006 ] []
        test2 - 00000006 *:00000000 [42:00000006 ] []
        test3 - 00000004 *:00000000 [15:00000004 ] []
  tmp - RuleIndexKind (1)
    * - RuleIndexKind (4)
      RuleIndexState (5) [TestRule1 ]
        name - 00000001 *:00000001 [] []
        test - 00000001 *:00000000 [val1:00000001 ] []
`[1:] {
		t.Error("Unexpected index layout:", res)
		return
	}

	// Make sure events without state do not match

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp", "x"},
		nil,
	}); printRules(res) != "[]" {
		t.Error("Unexpected result:", res)
		return
	}

	// Single rule match

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp", "x"},
		map[interface{}]interface{}{ // Match on event state
			"name": nil,
			"test": "val1",
		},
	}); printRules(res) != "[TestRule1]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		map[interface{}]interface{}{ // Match on event state
			"name": nil,
			"test": "val1",
		},
	}); printRules(res) != "[TestRule1]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		map[interface{}]interface{}{ // Match on event state
			"name":  "foobar",
			"test":  "val2",
			"test2": 42,
		},
	}); printRules(res) != "[TestRule2]" {
		t.Error("Unexpected result:", res)
		return
	}

	// Test multiple rule match

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		map[interface{}]interface{}{ // Match on event state
			"name":  nil,
			"test":  "val2",
			"test2": 42,
			"test3": 15,
		},
	}); printRules(res) != "[TestRule2 TestRule3]" {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestRuleIndexStateRegexMatch(t *testing.T) {
	ruleindexidcounter = 0
	defer func() {
		ruleindexidcounter = 0
	}()

	rule1 := &Rule{
		"TestRule1", // Name
		"",          // Description
		[]string{"core.main.tester", "core.tmp.*"}, // Kind match
		[]string{"data.read", "data.test"},         // Match on event cascade scope
		map[string]interface{}{ // Match on event state
			"name": nil,
			"test": regexp.MustCompile("val.*"),
		},
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	}

	rule2 := &Rule{
		"TestRule2",                  // Name
		"",                           // Description
		[]string{"core.main.tester"}, // Kind match
		[]string{"data.read"},        // Match on event cascade scope
		map[string]interface{}{ // Match on event state
			"name": nil,
			"test": regexp.MustCompile("va..*"),
		},
		0,                      // Priority of the rule
		[]string{"TestRule66"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return nil
		},
	}

	index := NewRuleIndex()

	index.AddRule(rule1)
	index.AddRule(rule2)

	// Check index layout

	if res := index.String(); res != `
core - RuleIndexKind (0)
  main - RuleIndexKind (1)
    tester - RuleIndexKind (2)
      RuleIndexState (3) [TestRule1 TestRule2 ]
        name - 00000003 *:00000003 [] []
        test - 00000003 *:00000003 [] [00000001:val.* 00000002:va..* ]
  tmp - RuleIndexKind (1)
    * - RuleIndexKind (4)
      RuleIndexState (5) [TestRule1 ]
        name - 00000001 *:00000001 [] []
        test - 00000001 *:00000001 [] [00000001:val.* ]
`[1:] {
		t.Error("Unexpected index layout:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp", "x"},
		map[interface{}]interface{}{ // Match on event state
			"name": "boo",
			"test": "val1",
		},
	}); printRules(res) != "[TestRule1]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "tmp", "x"},
		map[interface{}]interface{}{ // Match on event state
			"name": "boo",
			"test": "val",
		},
	}); printRules(res) != "[TestRule1]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		map[interface{}]interface{}{ // Match on event state
			"name": "boo",
			"test": "var",
		},
	}); printRules(res) != "[TestRule2]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester"},
		map[interface{}]interface{}{ // Match on event state
			"name": "boo",
			"test": "val",
		},
	}); printRules(res) != "[TestRule1 TestRule2]" {
		t.Error("Unexpected result:", res)
		return
	}

	// Test error cases

	if res := index.IsTriggering(&Event{
		"bla",
		[]string{"core", "main", "tester", "a"},
		map[interface{}]interface{}{ // Match on event state
			"name": "boo",
			"test": "val",
		},
	}); res {
		t.Error("Unexpected result:", res)
		return
	}

	if res := index.Match(&Event{
		"bla",
		[]string{"core", "main", "tester", "a"},
		map[interface{}]interface{}{ // Match on event state
			"name": "boo",
			"test": "val",
		},
	}); printRules(res) != "[]" {
		t.Error("Unexpected result:", res)
		return
	}
}

func printRules(rules []*Rule) string {
	var ret []string

	for _, r := range rules {
		ret = append(ret, r.Name)
	}

	sort.Strings(ret)

	return fmt.Sprint(ret)
}
