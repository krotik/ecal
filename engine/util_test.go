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

import "testing"

func TestRuleScope(t *testing.T) {

	rs := NewRuleScope(map[string]bool{
		"test.first.read":  true,
		"test.first.write": false,
		"test.second":      true,
	})

	if rs.IsAllowed("test.first") {
		t.Error("Unexpected result")
		return
	}

	if rs.IsAllowed("test.first.write") {
		t.Error("Unexpected result")
		return
	}

	if !rs.IsAllowed("test.first.read") {
		t.Error("Unexpected result")
		return
	}

	if !rs.IsAllowed("test.second") {
		t.Error("Unexpected result")
		return
	}

	if !rs.IsAllowed("test.second.bla") {
		t.Error("Unexpected result")
		return
	}

	// Test all is allowed

	rs = NewRuleScope(map[string]bool{
		"": true,
	})

	if !rs.IsAllowed("test.first") {
		t.Error("Unexpected result")
		return
	}

	if !rs.IsAllowed("test.first.write") {
		t.Error("Unexpected result")
		return
	}

	// Test nothing is allowed

	rs = NewRuleScope(nil)

	if rs.IsAllowed("test.first") {
		t.Error("Unexpected result")
		return
	}

	if rs.IsAllowed("test.first.write") {
		t.Error("Unexpected result")
		return
	}
}
