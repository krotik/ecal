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
	"testing"

	"devt.de/krotik/ecal/parser"
)

func TestNameFromASTNode(t *testing.T) {
	n, _ := parser.Parse("", "foo")

	if res := NameFromASTNode(n); res != "block: identifier (Line:1 Pos:1)" {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestScopeConversion(t *testing.T) {
	vs := NewScope("foo")

	vs.SetValue("a", 1)
	vs.SetValue("b", 2)
	vs.SetValue("c", 3)

	vs2 := ToScope("foo", ToObject(vs))

	if vs.String() != vs2.String() {
		t.Error("Unexpected result:", vs.String(), vs2.String())
		return
	}
}
