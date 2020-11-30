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
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/sortutil"
)

/*
Rule models a matching rule for event receivers (actions). A rule has 3 possible
matching criteria:

- Match on event kinds: A list of strings in dot notation which describes event kinds. May
contain '*' characters as wildcards (e.g. core.tests.*).

- Match on event cascade scope: A list of strings in dot notation which describe the
required scopes of an event cascade.

- Match on event state: A simple list of required key / value states in the event
state. Nil values can be used as wildcards (i.e. match is only on key).

Rules have priorities (0 being the highest) and may suppress each other.
*/
type Rule struct {
	Name            string                 // Name of the rule
	Desc            string                 // Description of the rule (optional)
	KindMatch       []string               // Match on event kinds
	ScopeMatch      []string               // Match on event cascade scope
	StateMatch      map[string]interface{} // Match on event state
	Priority        int                    // Priority of the rule
	SuppressionList []string               // List of suppressed rules by this rule
	Action          RuleAction             // Action of the rule
}

/*
CopyAs returns a shallow copy of this rule with a new name.
*/
func (r *Rule) CopyAs(newName string) *Rule {
	return &Rule{
		Name:            newName,
		Desc:            r.Desc,
		KindMatch:       r.KindMatch,
		ScopeMatch:      r.ScopeMatch,
		StateMatch:      r.StateMatch,
		Priority:        r.Priority,
		SuppressionList: r.SuppressionList,
		Action:          r.Action,
	}
}

func (r *Rule) String() string {
	sm, _ := json.Marshal(r.StateMatch)
	return fmt.Sprintf("Rule:%s [%s] (Priority:%v Kind:%v Scope:%v StateMatch:%s Suppress:%v)",
		r.Name, strings.TrimSpace(r.Desc), r.Priority, r.KindMatch, r.ScopeMatch, sm, r.SuppressionList)
}

/*
RuleAction is an action which is executed by a matching rule. The action gets
a unique thread ID from the executing thread.
*/
type RuleAction func(p Processor, m Monitor, e *Event, tid uint64) error

/*
RuleIndex is an index for rules. It takes the form of a tree structure in which
incoming events are matched level by level (e.g. event of kind core.task1.step1
is first matched by kind "core" then "task1" and then "step1". At the leaf of
the index tree it may then be matched on a state condition).
*/
type RuleIndex interface {

	/*
	   AddRule adds a new rule to the index.
	*/
	AddRule(rule *Rule) error

	/*
	   IsTriggering checks if a given event triggers a rule in this index.
	*/

	IsTriggering(event *Event) bool

	/*
		Match returns all rules in this index which match a given event. This
		method does a full matching check including state matching.
	*/
	Match(event *Event) []*Rule

	/*
		String returns a string representation of this rule index and all subindexes.
	*/
	String() string

	/*
		Rules returns all rules with the given prefix in the name. Use the empty
		string to return all rules.
	*/
	Rules() map[string]*Rule
}

/*
ruleSubIndex is a sub index used by a rule index.
*/
type ruleSubIndex interface {

	/*
		type returns the type of the rule sub index.
	*/
	Type() string

	/*
		addRuleAtLevel adds a new rule to the index at a specific level. The
		level is described by a part of the rule kind match.
	*/
	addRuleAtLevel(rule *Rule, kindMatchLevel []string)

	/*
		isTriggeringAtLevel checks if a given event triggers a rule at the given
		level of the index.
	*/
	isTriggeringAtLevel(event *Event, level int) bool

	/*
		matchAtLevel returns all rules in this index which match a given event
		at the given level. This method does a full matching check including
		state matching.
	*/
	matchAtLevel(event *Event, level int) []*Rule

	/*
		stringIndent returns a string representation with a given indentation of this
		rule index and all subindexes.
	*/
	stringIndent(indent string) string
}

/*
ruleIndexRoot models the index root node.
*/
type ruleIndexRoot struct {
	*RuleIndexKind
	rules map[string]*Rule
}

/*
   AddRule adds a new rule to the index.
*/
func (r *ruleIndexRoot) AddRule(rule *Rule) error {

	if _, ok := r.rules[rule.Name]; ok {
		return fmt.Errorf("Cannot add rule %v twice", rule.Name)
	}

	r.rules[rule.Name] = rule

	return r.RuleIndexKind.AddRule(rule)
}

/*
Rules returns all rules with the given prefix in the name. Use the empty
string to return all rules.
*/
func (r *ruleIndexRoot) Rules() map[string]*Rule {
	return r.rules
}

/*
NewRuleIndex creates a new rule container for efficient event matching.
*/
func NewRuleIndex() RuleIndex {
	return &ruleIndexRoot{newRuleIndexKind(), make(map[string]*Rule)}
}

/*
Rule index types
*/
const (
	typeRuleIndexKind  = "RuleIndexKind"
	typeRuleIndexState = "RuleIndexState"
	typeRuleIndexAll   = "RuleIndexAll"
)

// Rule Index Kind
// ===============

/*
RuleIndexKind data structure.
*/
type RuleIndexKind struct {
	id              uint64                    // Id of this rule index
	kindAllMatch    []ruleSubIndex            // Rules with target all events of a specific category
	kindSingleMatch map[string][]ruleSubIndex // Rules which target specific event kinds
	count           int                       // Number of loaded rules
}

/*
newRuleIndexKind creates a new rule index matching on event kind.
*/
func newRuleIndexKind() *RuleIndexKind {
	return &RuleIndexKind{
		newRuleIndexID(),
		make([]ruleSubIndex, 0),
		make(map[string][]ruleSubIndex),
		0,
	}
}

/*
Type returns the type of the rule sub index.
*/
func (ri *RuleIndexKind) Type() string {
	return typeRuleIndexKind
}

/*
AddRule adds a new rule to the index.
*/
func (ri *RuleIndexKind) AddRule(rule *Rule) error {

	// Check essential rule attributes

	if rule.KindMatch == nil || len(rule.KindMatch) == 0 {
		return fmt.Errorf("Cannot add rule without a kind match: %v", rule.Name)
	} else if rule.ScopeMatch == nil {
		return fmt.Errorf("Cannot add rule without a scope match: %v", rule.Name)
	}

	// Add rule to the index for all kind matches

	for _, kindMatch := range rule.KindMatch {
		ri.addRuleAtLevel(rule, strings.Split(kindMatch, RuleKindSeparator))
		ri.count++
	}

	return nil
}

/*
addRuleAtLevel adds a new rule to the index at a specific level. The
level is described by a part of the rule kind match.
*/
func (ri *RuleIndexKind) addRuleAtLevel(rule *Rule, kindMatchLevel []string) {
	var indexType string
	var index ruleSubIndex
	var ruleSubIndexList []ruleSubIndex
	var ok bool

	// Pick the right index type

	if len(kindMatchLevel) == 1 {
		if rule.StateMatch != nil {
			indexType = typeRuleIndexState
		} else {
			indexType = typeRuleIndexAll
		}
	} else {
		indexType = typeRuleIndexKind
	}

	// Get (create when necessary) a sub index of a specific type for the
	// match item of this level

	matchItem := kindMatchLevel[0]

	// Select the correct ruleSubIndexList

	if matchItem == RuleKindWildcard {
		ruleSubIndexList = ri.kindAllMatch
	} else {
		if ruleSubIndexList, ok = ri.kindSingleMatch[matchItem]; !ok {
			ruleSubIndexList = make([]ruleSubIndex, 0)
			ri.kindSingleMatch[matchItem] = ruleSubIndexList
		}
	}

	// Check if the required index is already existing

	for _, item := range ruleSubIndexList {
		if item.Type() == indexType {
			index = item
			break
		}
	}

	// Create a new index if no index was found

	if index == nil {
		switch indexType {
		case typeRuleIndexState:
			index = newRuleIndexState()
		case typeRuleIndexAll:
			index = newRuleIndexAll()
		case typeRuleIndexKind:
			index = newRuleIndexKind()
		}

		// Add the new index to the correct list

		if matchItem == RuleKindWildcard {
			ri.kindAllMatch = append(ruleSubIndexList, index)
		} else {
			ri.kindSingleMatch[matchItem] = append(ruleSubIndexList, index)
		}
	}

	// Recurse into the next level

	index.addRuleAtLevel(rule, kindMatchLevel[1:])
}

/*
IsTriggering checks if a given event triggers a rule in this index.
*/
func (ri *RuleIndexKind) IsTriggering(event *Event) bool {
	return ri.isTriggeringAtLevel(event, 0)
}

/*
isTriggeringAtLevel checks if a given event triggers a rule at the given
level of the index.
*/
func (ri *RuleIndexKind) isTriggeringAtLevel(event *Event, level int) bool {

	// Check if the event kind is too general (e.g. rule is defined as a.b.c
	// and the event kind is a.b)

	if len(event.kind) <= level {
		return false
	}

	levelKind := event.kind[level]
	nextLevel := level + 1

	// Check rules targeting all events

	for _, index := range ri.kindAllMatch {
		if index.isTriggeringAtLevel(event, nextLevel) {
			return true
		}
	}

	// Check rules targeting specific events

	if ruleSubIndexList, ok := ri.kindSingleMatch[levelKind]; ok {
		for _, index := range ruleSubIndexList {
			if index.isTriggeringAtLevel(event, nextLevel) {
				return true
			}
		}
	}

	return false
}

/*
Match returns all rules in this index which match a given event. This method
does a full matching check including state matching.
*/
func (ri *RuleIndexKind) Match(event *Event) []*Rule {
	return ri.matchAtLevel(event, 0)
}

/*
matchAtLevel returns all rules in this index which match a given event
at the given level. This method does a full matching check including
state matching.
*/
func (ri *RuleIndexKind) matchAtLevel(event *Event, level int) []*Rule {

	// Check if the event kind is too general (e.g. rule is defined as a.b.c
	// and the event kind is a.b)

	if len(event.kind) <= level {
		return nil
	}

	var ret []*Rule

	levelKind := event.kind[level]
	nextLevel := level + 1

	// Check rules targeting all events

	for _, index := range ri.kindAllMatch {
		ret = append(ret, index.matchAtLevel(event, nextLevel)...)
	}

	// Check rules targeting specific events

	if ruleSubIndexList, ok := ri.kindSingleMatch[levelKind]; ok {
		for _, index := range ruleSubIndexList {
			ret = append(ret, index.matchAtLevel(event, nextLevel)...)
		}
	}

	return ret
}

/*
String returns a string representation of this rule index and all subindexes.
*/
func (ri *RuleIndexKind) String() string {
	return ri.stringIndent("")
}

/*
stringIndent returns a string representation with a given indentation of this
rule index and all subindexes.
*/
func (ri *RuleIndexKind) stringIndent(indent string) string {
	var buf bytes.Buffer

	newIndent := indent + "  "

	writeIndexList := func(name string, indexList []ruleSubIndex) {
		if len(indexList) > 0 {

			buf.WriteString(fmt.Sprint(indent, name))
			buf.WriteString(fmt.Sprintf(" - %v (%v)\n", ri.Type(), ri.id))

			for _, index := range indexList {
				buf.WriteString(index.stringIndent(newIndent))
			}
		}
	}

	writeIndexList("*", ri.kindAllMatch)

	var keys []string
	for k := range ri.kindSingleMatch {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		indexList := ri.kindSingleMatch[key]
		writeIndexList(key, indexList)
	}

	return buf.String()
}

// Rule Index State
// ================

/*
RuleMatcherKey is used for pure key - value state matches.
*/
type RuleMatcherKey struct {
	bits        uint64
	bitsAny     uint64
	bitsValue   map[interface{}]uint64
	bitsRegexes map[uint64]*regexp.Regexp
}

/*
addRule adds a new rule to this key matcher.
*/
func (rm *RuleMatcherKey) addRule(num uint, bit uint64, key string, value interface{}) {

	// Register rule bit

	rm.bits |= bit

	if value == nil {
		rm.bitsAny |= bit

	} else if regex, ok := value.(*regexp.Regexp); ok {

		// For regex match we add a bit to the any mask so the presence of
		// the key is checked before the actual regex is checked

		rm.bitsAny |= bit
		rm.bitsRegexes[bit] = regex

	} else {
		rm.bitsValue[value] |= bit
	}
}

/*
match adds matching rules to a given bit mask.
*/
func (rm *RuleMatcherKey) match(bits uint64, value interface{}) uint64 {
	toRemove := rm.bitsAny ^ rm.bits

	if value != nil {
		if additionalBits, ok := rm.bitsValue[value]; ok {
			toRemove = rm.bitsAny | additionalBits ^ rm.bits
		}
	}

	keyMatchedBits := bits ^ (bits & toRemove)

	for bm, r := range rm.bitsRegexes {

		if keyMatchedBits&bm > 0 && !r.MatchString(fmt.Sprint(value)) {

			// Regex does not match remove the bit

			keyMatchedBits ^= keyMatchedBits & bm
		}
	}

	return keyMatchedBits
}

/*
unmatch removes all registered rules in this
*/
func (rm *RuleMatcherKey) unmatch(bits uint64) uint64 {
	return bits ^ (bits & rm.bits)
}

/*
String returns a string representation of this key matcher.
*/
func (rm *RuleMatcherKey) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%08X *:%08X", rm.bits, rm.bitsAny))

	buf.WriteString(" [")

	var keys []interface{}
	for k := range rm.bitsValue {
		keys = append(keys, k)
	}

	sortutil.InterfaceStrings(keys)

	for _, k := range keys {
		m := rm.bitsValue[k]
		buf.WriteString(fmt.Sprintf("%v:%08X ", k, m))
	}

	buf.WriteString("] [")

	var rkeys []uint64
	for k := range rm.bitsRegexes {
		rkeys = append(rkeys, k)
	}

	sortutil.UInt64s(rkeys)

	for _, k := range rkeys {
		r := rm.bitsRegexes[k]
		buf.WriteString(fmt.Sprintf("%08X:%v ", k, r))
	}

	buf.WriteString("]")

	return buf.String()
}

/*
RuleIndexState data structure
*/
type RuleIndexState struct {
	id     uint64                     // Id of this rule index
	rules  []*Rule                    // All rules stored in this index
	keyMap map[string]*RuleMatcherKey // Map of keys (key or key and value) to KeyMatcher
}

/*
newRuleIndexState creates a new rule index matching on event state.
*/
func newRuleIndexState() *RuleIndexState {
	return &RuleIndexState{newRuleIndexID(), make([]*Rule, 0),
		make(map[string]*RuleMatcherKey)}
}

/*
Type returns the type of the rule sub index.
*/
func (ri *RuleIndexState) Type() string {
	return typeRuleIndexState
}

/*
addRuleAtLevel adds a new rule to the index at a specific level. The
level is described by a part of the rule kind match.
*/
func (ri *RuleIndexState) addRuleAtLevel(rule *Rule, kindMatchLevel []string) {
	errorutil.AssertTrue(len(kindMatchLevel) == 0,
		fmt.Sprint("RuleIndexState must be a leaf - level is:", kindMatchLevel))

	num := uint(len(ri.rules))
	var bit uint64 = 1 << num

	ri.rules = append(ri.rules, rule)

	for k, v := range rule.StateMatch {
		var ok bool
		var keyMatcher *RuleMatcherKey

		if keyMatcher, ok = ri.keyMap[k]; !ok {
			keyMatcher = &RuleMatcherKey{0, 0, make(map[interface{}]uint64), make(map[uint64]*regexp.Regexp)}
			ri.keyMap[k] = keyMatcher
		}

		keyMatcher.addRule(num, bit, k, v)
	}
}

/*
isTriggeringAtLevel checks if a given event triggers a rule at the given
level of the index.
*/
func (ri *RuleIndexState) isTriggeringAtLevel(event *Event, level int) bool {
	return len(event.kind) == level
}

/*
matchAtLevel returns all rules in this index which match a given event
at the given level. This method does a full matching check including
state matching.
*/
func (ri *RuleIndexState) matchAtLevel(event *Event, level int) []*Rule {
	if len(event.kind) != level {
		return nil
	}

	// Assume all rules match and remove the ones with don't

	var matchBits uint64 = (1 << uint(len(ri.rules))) - 1

	// Match key and values

	for key, matcher := range ri.keyMap {
		if val, ok := event.state[key]; ok {

			// Key is present in event

			matchBits = matcher.match(matchBits, val)

		} else {

			// Key is not present in event - remove all rules which require the key

			matchBits = matcher.unmatch(matchBits)
		}

		if matchBits == 0 {

			// All rules have been excluded

			return nil
		}
	}

	var ret []*Rule
	var collectionBits uint64 = 1

	// Collect matched rules

	for i := 0; collectionBits <= matchBits; i++ {
		if matchBits&collectionBits > 0 {
			ret = append(ret, ri.rules[i])
		}

		collectionBits <<= 1
	}

	return ret
}

/*
stringIndent returns a string representation with a given indentation of this
rule index and all subindexes.
*/
func (ri *RuleIndexState) stringIndent(indent string) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%v%v (%v) ", indent, ri.Type(), ri.id))
	buf.WriteString("[")
	for _, r := range ri.rules {
		buf.WriteString(fmt.Sprintf("%v ", r.Name))
	}
	buf.WriteString("]\n")

	newIndent := indent + "  "

	var keys []string
	for k := range ri.keyMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		m := ri.keyMap[k]
		buf.WriteString(fmt.Sprintf("%v%v - %v\n", newIndent, k, m))
	}

	return buf.String()
}

// Rule Index All
// ==============

/*
RuleIndexAll data structure.
*/
type RuleIndexAll struct {
	id    uint64  // Id of this rule index
	rules []*Rule // Rules with target all events of a specific category
}

/*
newRuleIndexAll creates a new leaf rule index matching on all events.
*/
func newRuleIndexAll() *RuleIndexAll {
	return &RuleIndexAll{newRuleIndexID(), make([]*Rule, 0)}
}

/*
Type returns the type of the rule sub index.
*/
func (ri *RuleIndexAll) Type() string {
	return typeRuleIndexAll
}

/*
addRuleAtLevel adds a new rule to the index at a specific level. The
level is described by a part of the rule kind match.
*/
func (ri *RuleIndexAll) addRuleAtLevel(rule *Rule, kindMatchLevel []string) {
	ri.rules = append(ri.rules, rule)
}

/*
isTriggeringAtLevel checks if a given event triggers a rule at the given
level of the index.
*/
func (ri *RuleIndexAll) isTriggeringAtLevel(event *Event, level int) bool {
	return len(event.kind) == level
}

/*
matchAtLevel returns all rules in this index which match a given event
at the given level. This method does a full matching check including
state matching.
*/
func (ri *RuleIndexAll) matchAtLevel(event *Event, level int) []*Rule {
	if len(event.kind) != level {
		return nil
	}

	return ri.rules
}

/*
stringIndent returns a string representation with a given indentation of this
rule index and all subindexes.
*/
func (ri *RuleIndexAll) stringIndent(indent string) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%v%v (%v)\n", indent, ri.Type(), ri.id))

	newIndent := indent + "  "

	for _, rule := range ri.rules {
		buf.WriteString(fmt.Sprintf("%v%v\n", newIndent, rule))
	}

	return buf.String()
}

// Unique id creation
// ==================

var ruleindexidcounter uint64 = 1
var ruleindexidcounterLock = &sync.Mutex{}

/*
newId returns a new unique id.
*/
func newRuleIndexID() uint64 {
	ruleindexidcounterLock.Lock()
	defer ruleindexidcounterLock.Unlock()

	ret := ruleindexidcounter
	ruleindexidcounter++

	return ret
}
