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
	"reflect"
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
