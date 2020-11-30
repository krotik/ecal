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
	"sort"
	"strings"
)

// Globals
// =======

/*
RuleKindSeparator is the separator for rule kinds
*/
const RuleKindSeparator = "."

/*
RuleKindWildcard is a wildcard for rule kinds
*/
const RuleKindWildcard = "*"

// Messages
// ========

/*
MessageRootMonitorFinished is send from a root monitor when it has finished
*/
const MessageRootMonitorFinished = "MessageRootMonitorFinished"

// Rule Scope
// ==========

/*
RuleScope is a set of scope definitions for rules. Each definition allows or disallows
a set of rule types. Scope definitions and rule sets are usually expressed with
named paths (scope paths) in dot notation (e.g. core.data.read).
*/
type RuleScope struct {
	scopeDefs map[string]interface{}
}

const ruleScopeAllowFlag = "."

/*
NewRuleScope creates a new rule scope object with an initial set of definitions.
*/
func NewRuleScope(allows map[string]bool) *RuleScope {
	rs := &RuleScope{make(map[string]interface{})}
	rs.AddAll(allows)
	return rs
}

/*
IsAllowedAll checks if all given scopes are allowed.
*/
func (rs *RuleScope) IsAllowedAll(scopePaths []string) bool {
	for _, path := range scopePaths {
		if !rs.IsAllowed(path) {
			return false
		}
	}
	return true
}

/*
IsAllowed checks if a given scope path is allowed within this rule scope.
*/
func (rs *RuleScope) IsAllowed(scopePath string) bool {
	allowed := false
	scopeDefs := rs.scopeDefs

	if a, ok := scopeDefs[ruleScopeAllowFlag]; ok {
		allowed = a.(bool)
	}

	for _, scopeStep := range strings.Split(scopePath, ".") {
		val, ok := scopeDefs[scopeStep]

		if !ok {
			break
		}

		scopeDefs = val.(map[string]interface{})

		if a, ok := scopeDefs[ruleScopeAllowFlag]; ok {
			allowed = a.(bool)
		}
	}

	return allowed
}

/*
AddAll adds all given definitions to the rule scope.
*/
func (rs *RuleScope) AddAll(allows map[string]bool) {
	for scopePath, allow := range allows {
		rs.Add(scopePath, allow)
	}
}

/*
Add adds a given definition to the rule scope.
*/
func (rs *RuleScope) Add(scopePath string, allow bool) {
	scopeDefs := rs.scopeDefs

	if scopePath != "" {
		for _, scopeStep := range strings.Split(scopePath, ".") {
			val, ok := scopeDefs[scopeStep]

			if !ok {
				val = make(map[string]interface{})
				scopeDefs[scopeStep] = val
			}

			scopeDefs = val.(map[string]interface{})
		}
	}

	scopeDefs[ruleScopeAllowFlag] = allow
}

// Rule sorting
// ============

/*
RuleSlice is a slice of rules
*/
type RuleSlice []*Rule

func (s RuleSlice) Len() int           { return len(s) }
func (s RuleSlice) Less(i, j int) bool { return s[i].Priority < s[j].Priority }
func (s RuleSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

/*
SortRuleSlice sorts a slice of rules.
*/
func SortRuleSlice(a []*Rule) { sort.Sort(RuleSlice(a)) }

// Unit testing
// ============

/*
UnitTestResetIDs reset all counting IDs.
THIS FUNCTION SHOULD ONLY BE CALLED BY UNIT TESTS!
*/
func UnitTestResetIDs() {
	pidcounter = 1
	ruleindexidcounter = 1
	midcounter = 1
}
