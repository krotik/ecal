/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package scope

import (
	"encoding/json"
	"fmt"
	"testing"

	"devt.de/krotik/ecal/parser"
)

func TestVarScopeSetMap(t *testing.T) {

	parentVS := NewScope("global")
	childVs := NewScopeWithParent("c1", parentVS)

	// Test map

	parentVS.SetValue("xx", map[interface{}]interface{}{
		"foo": map[interface{}]interface{}{},
	})

	childVs.SetValue("xx.foo.bar", map[interface{}]interface{}{})

	childVs.SetValue("xx.foo.bar.99", "tester")

	if childVs.Parent().String() != `
global {
    xx (map[interface {}]interface {}) : {"foo":{"bar":{"99":"tester"}}}
}`[1:] {
		t.Error("Unexpected result:", parentVS.String())
		return
	}

	childVs.SetValue("xx.foo.bar.99", []interface{}{1, 2})

	if parentVS.String() != `
global {
    xx (map[interface {}]interface {}) : {"foo":{"bar":{"99":[1,2]}}}
}`[1:] {
		t.Error("Unexpected result:", parentVS.String())
		return
	}

	childVs.SetValue("xx.foo.bar", map[interface{}]interface{}{
		float64(22): "foo",
		"33":        "bar",
	})

	if res, _, _ := childVs.GetValue("xx.foo.bar.33"); res != "bar" {
		t.Error("Unexpected result:", res)
		return
	}

	if res, _, _ := childVs.GetValue("xx.foo.bar.22"); res != "foo" {
		t.Error("Unexpected result:", res)
		return
	}

	// Test errors

	err := parentVS.SetValue("xx.foo.a.b", 5)

	if err == nil || err.Error() != "Container field xx.foo.a does not exist" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	// Test list (test also overwriting of variables)

	parentVS.SetValue("xx", []interface{}{1, 2, []interface{}{3, 4}})

	if parentVS.String() != `
global {
    xx ([]interface {}) : [1,2,[3,4]]
}`[1:] {
		t.Error("Unexpected result:", parentVS.String())
		return
	}

	parentVS.SetValue("xx.2", []interface{}{3, 4, 5})

	if parentVS.String() != `
global {
    xx ([]interface {}) : [1,2,[3,4,5]]
}`[1:] {
		t.Error("Unexpected result:", parentVS.String())
		return
	}

	parentVS.SetValue("xx.-1", []interface{}{3, 4, 6})

	if parentVS.String() != `
global {
    xx ([]interface {}) : [1,2,[3,4,6]]
}`[1:] {
		t.Error("Unexpected result:", parentVS.String())
		return
	}

	parentVS.SetValue("xx.-1.-1", 7)

	if parentVS.String() != `
global {
    xx ([]interface {}) : [1,2,[3,4,7]]
}`[1:] {
		t.Error("Unexpected result:", parentVS.String())
		return
	}

	testVarScopeSetMapErrors(t, parentVS)
}

func testVarScopeSetMapErrors(t *testing.T, parentVS parser.Scope) {

	err := parentVS.SetValue("xx.a", []interface{}{3, 4, 5})

	if err.Error() != "List xx needs a number index not: a" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.2.b", []interface{}{3, 4, 5})

	if err.Error() != "List xx.2 needs a number index not: b" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.2.b.1", []interface{}{3, 4, 5})

	if err.Error() != "List xx.2 needs a number index not: b" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.2.1.b.1", []interface{}{3, 4, 5})

	if err.Error() != "Variable xx.2.1 is not a container" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.2.1.2", []interface{}{3, 4, 5})

	if err.Error() != "Variable xx.2.1 is not a container" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.5", []interface{}{3, 4, 5})

	if err.Error() != "Out of bounds access to list xx with index: 5" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.5.1", []interface{}{3, 4, 5})

	if err.Error() != "Out of bounds access to list xx with index: 5" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("xx.2.5.1", []interface{}{3, 4, 5})

	if err.Error() != "Out of bounds access to list xx.2 with index: 5" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}

	err = parentVS.SetValue("yy.2.5.1", []interface{}{3, 4, 5})

	if err.Error() != "Variable yy is not a container" {
		t.Error("Unexpected result:", parentVS.String(), err)
		return
	}
}

func TestVarScopeGet(t *testing.T) {

	parentVs := NewScope("")
	childVs := parentVs.NewChild("")

	parentVs.SetValue("xx", map[interface{}]interface{}{
		"foo": map[interface{}]interface{}{
			"bar": 99,
		},
	})

	parentVs.SetValue("test", []interface{}{1, 2, []interface{}{3, 4}})

	if res := fmt.Sprint(childVs.GetValue("xx")); res != "map[foo:map[bar:99]] true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("xx.foo")); res != "map[bar:99] true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("xx.foo.bar")); res != "99 true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test")); res != "[1 2 [3 4]] true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.2")); res != "[3 4] true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.2.1")); res != "4 true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.-1.1")); res != "4 true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.-1.-1")); res != "4 true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.-1.-2")); res != "3 true <nil>" {
		t.Error("Unexpected result:", res)
		return
	}

	// Test error cases

	if res := fmt.Sprint(childVs.GetValue("test.a")); res != "<nil> false List test needs a number index not: a" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.5")); res != "<nil> false Out of bounds access to list test with index: 5" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.2.1.1")); res != "<nil> false Variable test.2.1 is not a container" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test.1.1.1")); res != "<nil> false Variable test.1 is not a container" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(childVs.GetValue("test2.1.1.1")); res != "<nil> false <nil>" {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestVarScopeDump(t *testing.T) {

	// Build a small tree of VS

	globalVS1 := NewScope("global")
	globalVS2 := NewScopeWithParent("global2", globalVS1)
	globalVS3 := NewScopeWithParent("global3", globalVS1)
	sinkVs1 := globalVS2.NewChild("sink: 1")
	globalVS2.NewChild("sink: 1") // This should have no effect
	globalVS2.NewChild("sink: 2") // Reference is provided in the next call
	sinkVs2 := globalVS1.NewChild("sink: 2")
	for1Vs := sinkVs1.NewChild("block: for1")
	for2Vs := sinkVs1.NewChild("block: for2")
	for21Vs := for2Vs.NewChild("block: for2-1")
	for211Vs := for21Vs.NewChild("block: for2-1-1")

	// Populate tree

	globalVS1.SetValue("0", 0)
	globalVS2.SetValue("a", 1)
	globalVS3.SetValue("a", 2)
	sinkVs1.SetValue("b", 2)
	for1Vs.SetValue("c", 3)
	for2Vs.SetValue("d", 4)
	for21Vs.SetValue("e", 5)
	for211Vs.SetValue("f", 6)
	for211Vs.SetValue("x", ToObject(for211Vs))
	sinkVs2.SetValue("g", 2)

	sinkVs1.SetLocalValue("a", 5)

	// Dump the sinkVs1 scope

	if res := sinkVs1.String(); res != `global {
    0 (int) : 0
    global2 {
        a (int) : 1
        sink: 1 {
            a (int) : 5
            b (int) : 2
            block: for1 {
                c (int) : 3
            }
            block: for2 {
                d (int) : 4
                block: for2-1 {
                    e (int) : 5
                    block: for2-1-1 {
                        f (int) : 6
                        x (map[interface {}]interface {}) : {"f":6}
                    }
                }
            }
        }
    }
}` {
		t.Error("Unexpected result:", res)
		return
	}

	bytes, _ := json.Marshal(sinkVs1.ToJSONObject())
	if res := string(bytes); res != `{"a":5,"b":2}` {
		t.Error("Unexpected result:", res)
		return
	}

	bytes, _ = json.Marshal(for211Vs.ToJSONObject())
	if res := string(bytes); res != `{"f":6,"x":{"f":6}}` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := globalVS1.Name(); res != "global" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := sinkVs2.String(); res != `global {
    0 (int) : 0
    sink: 2 {
        g (int) : 2
    }
}` {
		t.Error("Unexpected result:", res)
		return
	}

	// Dumping the global scope results in the same output as it does not
	// track child scopes

	if res := globalVS1.String(); res != `global {
    0 (int) : 0
    sink: 2 {
        g (int) : 2
    }
}` {
		t.Error("Unexpected result:", res)
		return
	}

	sinkVs1.Clear()

	if res := sinkVs1.String(); res != `global {
    0 (int) : 0
    global2 {
        a (int) : 1
        sink: 1 {
        }
    }
}` {
		t.Error("Unexpected result:", res)
		return
	}
}
