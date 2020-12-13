/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package stdlib

import (
	"fmt"
	"math"
	"plugin"
	"reflect"
	"strings"
	"testing"
)

func TestGetPkgDocString(t *testing.T) {
	AddStdlibPkg("foo", "foo doc")

	mathFuncMap["Println"] = &ECALFunctionAdapter{reflect.ValueOf(fmt.Println), "foo"}

	f, _ := GetStdlibFunc("math.Println")

	if s, _ := f.DocString(); s != "foo" {
		t.Error("Unexpected result:", s)
		return
	}

	doc, _ := GetPkgDocString("math")

	if doc == "" {
		t.Error("Unexpected result:", doc)
		return
	}

	doc, _ = GetPkgDocString("foo")

	if doc != "foo doc" {
		t.Error("Unexpected result:", doc)
		return
	}

	if err := AddStdlibPkg("foo", "foo doc"); err == nil || err.Error() != "Package foo already exists" {
		t.Error("Unexpected error:", err)
		return
	}
}

func TestSymbols(t *testing.T) {
	AddStdlibPkg("foo", "foo doc")
	AddStdlibFunc("foo", "bar", nil)

	p, c, f := GetStdlibSymbols()
	if len(p) == 0 || len(c) == 0 || len(f) == 0 {
		t.Error("Should have some entries in symbol lists:", p, c, f)
		return
	}
}

func TestSplitModuleAndName(t *testing.T) {

	if m, n := splitModuleAndName("fmt.Println"); m != "fmt" || n != "Println" {
		t.Error("Unexpected result:", m, n)
		return
	}

	if m, n := splitModuleAndName(""); m != "" || n != "" {
		t.Error("Unexpected result:", m, n)
		return
	}

	if m, n := splitModuleAndName("a"); m != "a" || n != "" {
		t.Error("Unexpected result:", m, n)
		return
	}

	if m, n := splitModuleAndName("my.FuncCall"); m != "my" || n != "FuncCall" {
		t.Error("Unexpected result:", m, n)
		return
	}
}

func TestGetStdLibItems(t *testing.T) {
	dummyFunc := &ECALFunctionAdapter{}
	AddStdlibFunc("foo", "bar", dummyFunc)

	mathFuncMap["Println"] = &ECALFunctionAdapter{reflect.ValueOf(fmt.Println), "foo"}

	if f, _ := GetStdlibFunc("math.Println"); f != mathFuncMap["Println"] {
		t.Error("Unexpected resutl: functions should lookup correctly")
		return
	}

	if f, _ := GetStdlibFunc("foo.bar"); f != dummyFunc {
		t.Error("Unexpected resutl: functions should lookup correctly")
		return
	}

	if c, ok := GetStdlibFunc("foo"); c != nil || ok {
		t.Error("Unexpected resutl: func should lookup correctly")
		return
	}

	if c, _ := GetStdlibConst("math.Pi"); c != math.Pi {
		t.Error("Unexpected resutl: constants should lookup correctly")
		return
	}

	if c, ok := GetStdlibConst("foo"); c != nil || ok {
		t.Error("Unexpected resutl: constants should lookup correctly")
		return
	}
}

func TestAddStdLibFunc(t *testing.T) {
	dummyFunc := &ECALFunctionAdapter{}
	AddStdlibFunc("foo", "bar", dummyFunc)

	if f, _ := GetStdlibFunc("foo.bar"); f != dummyFunc {
		t.Error("Unexpected resutl: functions should lookup correctly")
		return
	}

	if err := AddStdlibFunc("foo2", "bar", dummyFunc); err == nil || err.Error() != "Package foo2 does not exist" {
		t.Error("Unexpected error:", err)
		return
	}
}

func TestAddPluginStdLibFunc(t *testing.T) {
	var err error

	// Uncomment the commented parts in this test to run against an
	// actual compiled plugin in the examples/plugin directory

	/*
		err = AddStdlibPluginFunc("foo", "bar", filepath.Join("..", "examples", "plugin", "myfunc.so"), "ECALmyfunc")

		if err != nil {
			t.Error("Unexpected result:", err)
			return
		}
	*/

	pluginTestLookup = &testLookup{&testECALPluginFunction{}, nil}

	err = AddStdlibPluginFunc("foo", "bar", "", "ECALmyfunc")

	pluginTestLookup = nil

	if err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	pluginTestLookup = &testLookup{&testECALPluginFunction{}, nil}

	errs := LoadStdlibPlugins([]interface{}{
		map[string]interface{}{
			"package": "foo",
			"name":    "bar",
			"path":    "",
			"symbol":  "ECALmyfunc",
		},
		map[string]interface{}{
			"package": "foo",
			"name":    "bar",
			"path":    "",
			"symbol":  "showerror",
		},
	})

	pluginTestLookup = nil

	if fmt.Sprint(errs) != "[Test lookup error]" {
		t.Error("Unexpected result:", errs)
		return
	}

	pfunc, ok := GetStdlibFunc("foo.bar")

	if !ok {
		t.Error("Unexpected result:", pfunc, ok)
		return
	}

	res, err := pfunc.Run("", nil, nil, 0, []interface{}{"John"})

	if err != nil || res != "Hello World for John" {
		t.Error("Unexpected result:", res, err)
		return
	}

	// Test errors

	/*
		err = AddStdlibPluginFunc("foo", "bar", filepath.Join("..", "examples", "plugin", "myfunc.so"), "Greeting")

		if err == nil || err.Error() != "Symbol Greeting is not a stdlib function" {
			t.Error("Unexpected result:", err)
			return
		}

		err = AddStdlibPluginFunc("foo", "bar", filepath.Join("..", "examples", "plugin", "myfunc.so"), "foo")

		if err == nil || !strings.Contains(err.Error(), "symbol foo not found") {
			t.Error("Unexpected result:", err)
			return
		}
	*/

	pluginTestLookup = &testLookup{"foo", nil}
	err = AddStdlibPluginFunc("foo", "bar", "", "Greeting")

	if err == nil || err.Error() != "Symbol Greeting is not a stdlib function" {
		t.Error("Unexpected result:", err)
		return
	}

	pluginTestLookup = &testLookup{nil, fmt.Errorf("symbol foo not found")}
	err = AddStdlibPluginFunc("foo", "bar", "", "foo")

	if err == nil || !strings.Contains(err.Error(), "symbol foo not found") {
		t.Error("Unexpected result:", err)
		return
	}
}

type testLookup struct {
	ret interface{}
	err error
}

func (tl *testLookup) Lookup(symName string) (plugin.Symbol, error) {
	if symName == "showerror" {
		return nil, fmt.Errorf("Test lookup error")
	}
	return tl.ret, tl.err
}

type testECALPluginFunction struct {
}

func (tf *testECALPluginFunction) Run(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("Need a name to greet as argument")
	}

	return fmt.Sprintf("Hello World for %v", args[0]), nil
}

func (tf *testECALPluginFunction) DocString() string {
	return "Myfunc is an example function"
}
