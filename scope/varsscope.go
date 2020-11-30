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
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/ecal/parser"
)

/*
varsScope models a scope for variables in ECAL.
*/
type varsScope struct {
	name     string                 // Name of the scope
	parent   parser.Scope           // Parent scope
	children []*varsScope           // Children of this scope (only if tracking is enabled)
	storage  map[string]interface{} // Storage for variables
	lock     *sync.RWMutex          // Lock for this scope
}

/*
NewScope creates a new variable scope.
*/
func NewScope(name string) parser.Scope {
	return NewScopeWithParent(name, nil)
}

/*
NewScopeWithParent creates a new variable scope with a parent. This can be
used to create scope structures without children links.
*/
func NewScopeWithParent(name string, parent parser.Scope) parser.Scope {
	res := &varsScope{name, nil, nil, make(map[string]interface{}), &sync.RWMutex{}}
	SetParentOfScope(res, parent)
	return res
}

/*
SetParentOfScope sets the parent of a given scope. This assumes that the given scope
is a varsScope.
*/
func SetParentOfScope(scope parser.Scope, parent parser.Scope) {
	if pvs, ok := parent.(*varsScope); ok {
		if vs, ok := scope.(*varsScope); ok {

			vs.lock.Lock()
			defer vs.lock.Unlock()
			pvs.lock.Lock()
			defer pvs.lock.Unlock()

			vs.parent = parent
			vs.lock = pvs.lock
		}
	}
}

/*
NewChild creates a new child scope for variables. The new child scope is tracked
by the parent scope. This means it should not be used for global scopes with
many children.
*/
func (s *varsScope) NewChild(name string) parser.Scope {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, c := range s.children {
		if c.name == name {
			return c
		}
	}

	child := NewScope(name).(*varsScope)
	child.parent = s
	child.lock = s.lock
	s.children = append(s.children, child)

	return child
}

/*
Name returns the name of this scope.
*/
func (s *varsScope) Name() string {
	return s.name
}

/*
Clear clears this scope of all stored values. This will clear children scopes
but not remove parent scopes.
*/
func (s *varsScope) Clear() {
	s.children = nil
	s.storage = make(map[string]interface{})
}

/*
Parent returns the parent scope or nil.
*/
func (s *varsScope) Parent() parser.Scope {
	return s.parent
}

/*
SetValue sets a new value for a variable.
*/
func (s *varsScope) SetValue(varName string, varValue interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.setValue(varName, varValue)
}

/*
SetLocalValue sets a new value for a local variable.
*/
func (s *varsScope) SetLocalValue(varName string, varValue interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Ensure the variable exists in the local scope

	localVarName := strings.Split(varName, ".")[0]
	s.storage[localVarName] = nil

	return s.setValue(varName, varValue)
}

/*
setValue sets a new value for a variable.
*/
func (s *varsScope) setValue(varName string, varValue interface{}) error {
	var err error

	// Check for dotted names which access a container structure

	if cFields := strings.Split(varName, "."); len(cFields) > 1 {

		// Get the container

		if container, ok, _ := s.getValue(cFields[0]); ok {

			if len(cFields) > 2 {

				var containerAccess func(fields []string)

				containerAccess = func(fields []string) {

					// Get inner container

					if mapContainer, ok := container.(map[interface{}]interface{}); ok {

						if container, ok = mapContainer[fields[0]]; !ok {
							err = fmt.Errorf("Container field %v does not exist",
								strings.Join(cFields[:len(cFields)-len(fields)+1], "."))
						}

					} else if listContainer, ok := container.([]interface{}); ok {
						var index int

						if index, err = strconv.Atoi(fmt.Sprint(fields[0])); err == nil {

							if index < 0 {

								// Handle negative numbers

								index = len(listContainer) + index
							}

							if index < len(listContainer) {
								container = listContainer[index]
							} else {
								err = fmt.Errorf("Out of bounds access to list %v with index: %v",
									strings.Join(cFields[:len(cFields)-len(fields)], "."), index)
							}

						} else {
							container = nil
							err = fmt.Errorf("List %v needs a number index not: %v",
								strings.Join(cFields[:len(cFields)-len(fields)], "."), fields[0])
						}

					} else {
						container = nil
						err = fmt.Errorf("Variable %v is not a container",
							strings.Join(cFields[:len(cFields)-len(fields)], "."))
					}

					if err == nil && len(fields) > 2 {
						containerAccess(fields[1:])
					}
				}

				containerAccess(cFields[1:])
			}

			if err == nil && container != nil {

				fieldIndex := cFields[len(cFields)-1]

				if mapContainer, ok := container.(map[interface{}]interface{}); ok {

					mapContainer[fieldIndex] = varValue

				} else if listContainer, ok := container.([]interface{}); ok {
					var index int

					if index, err = strconv.Atoi(fieldIndex); err == nil {

						if index < 0 {

							// Handle negative numbers

							index = len(listContainer) + index
						}

						if index < len(listContainer) {
							listContainer[index] = varValue
						} else {
							err = fmt.Errorf("Out of bounds access to list %v with index: %v",
								strings.Join(cFields[:len(cFields)-1], "."), index)
						}
					} else {
						err = fmt.Errorf("List %v needs a number index not: %v",
							strings.Join(cFields[:len(cFields)-1], "."), fieldIndex)
					}
				} else {
					err = fmt.Errorf("Variable %v is not a container",
						strings.Join(cFields[:len(cFields)-1], "."))
				}
			}

		} else {
			err = fmt.Errorf("Variable %v is not a container", cFields[0])
		}

		return err
	}

	// Check if the variable is already defined in a parent scope

	if vs := s.getScopeForVariable(varName); vs != nil {
		s = vs
	}

	// Set value newly in scope

	s.storage[varName] = varValue

	return err
}

/*
getScopeForVariable returns the scope (this or a parent scope) which holds a
given variable.
*/
func (s *varsScope) getScopeForVariable(varName string) *varsScope {

	_, ok := s.storage[varName]

	if ok {
		return s
	} else if s.parent != nil {
		return s.parent.(*varsScope).getScopeForVariable(varName)
	}

	return nil
}

/*
GetValue gets the current value of a variable.
*/
func (s *varsScope) GetValue(varName string) (interface{}, bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getValue(varName)
}

/*
getValue gets the current value of a variable.
*/
func (s *varsScope) getValue(varName string) (interface{}, bool, error) {

	// Check for dotted names which access a container structure

	if cFields := strings.Split(varName, "."); len(cFields) > 1 {
		var err error
		var containerAccess func(fields []string, container interface{}) (interface{}, bool, error)

		// Get the container

		container, ok, _ := s.getValue(cFields[0])

		if !ok {
			return nil, ok, err
		}

		// Now look into the container and get the value

		containerAccess = func(fields []string, container interface{}) (interface{}, bool, error) {
			var retContainer interface{}

			if mapContainer, ok := container.(map[interface{}]interface{}); ok {
				var ok bool

				if index, err := strconv.Atoi(fmt.Sprint(fields[0])); err == nil {

					// Numbers are usually converted to float64

					retContainer, ok = mapContainer[float64(index)]
				}

				if !ok {
					retContainer = mapContainer[fields[0]]
				}

			} else if listContainer, ok := container.([]interface{}); ok {
				var index int

				if index, err = strconv.Atoi(fmt.Sprint(fields[0])); err == nil {

					if index < 0 {

						// Handle negative numbers

						index = len(listContainer) + index
					}

					if index < len(listContainer) {
						retContainer = listContainer[index]
					} else {
						err = fmt.Errorf("Out of bounds access to list %v with index: %v",
							strings.Join(cFields[:len(cFields)-len(fields)], "."), index)
					}

				} else {
					err = fmt.Errorf("List %v needs a number index not: %v",
						strings.Join(cFields[:len(cFields)-len(fields)], "."), fields[0])
				}

			} else {
				err = fmt.Errorf("Variable %v is not a container",
					strings.Join(cFields[:len(cFields)-len(fields)], "."))
			}

			if err == nil && len(fields) > 1 {
				return containerAccess(fields[1:], retContainer)
			}

			return retContainer, retContainer != nil, err
		}

		return containerAccess(cFields[1:], container)
	}

	if vs := s.getScopeForVariable(varName); vs != nil {

		ret := vs.storage[varName]

		return ret, true, nil
	}

	return nil, false, nil
}

/*
String returns a string representation of this varsScope and all its
parents.
*/
func (s *varsScope) String() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.scopeStringParents(s.scopeStringChildren())
}

/*
ToJSONObject returns this ASTNode and all its children as a JSON object.
*/
func (s *varsScope) ToJSONObject() map[string]interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ret := make(map[string]interface{})

	for k, v := range s.storage {
		var value interface{}

		value = fmt.Sprintf("ComplexDataStructure: %#v", v)

		bytes, err := json.Marshal(v)
		if err != nil {
			bytes, err = json.Marshal(stringutil.ConvertToJSONMarshalableObject(v))

		}
		if err == nil {
			json.Unmarshal(bytes, &value)
		}

		ret[k] = value
	}

	return ret
}

/*
scopeStringChildren returns a string representation of all children scopes.
*/
func (s *varsScope) scopeStringChildren() string {
	var buf bytes.Buffer

	// Write the known child scopes

	for i, c := range s.children {
		buf.WriteString(c.scopeString(c.scopeStringChildren()))
		if i < len(s.children)-1 {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

/*
scopeStringParents returns a string representation of this varsScope
with initial children and all its parents.
*/
func (s *varsScope) scopeStringParents(childrenString string) string {
	ss := s.scopeString(childrenString)

	if s.parent != nil {
		return s.parent.(*varsScope).scopeStringParents(ss)
	}

	return fmt.Sprint(ss)
}

/*
scopeString returns a string representation of this varsScope.
*/
func (s *varsScope) scopeString(childrenString string) string {
	buf := bytes.Buffer{}
	varList := []string{}

	buf.WriteString(fmt.Sprintf("%v {\n", s.name))

	for k := range s.storage {
		varList = append(varList, k)
	}

	sort.Strings(varList)

	for _, v := range varList {
		buf.WriteString(fmt.Sprintf("    %s (%T) : %v\n", v, s.storage[v],
			EvalToString(s.storage[v])))
	}

	if childrenString != "" {

		// Indent all

		buf.WriteString("    ")
		buf.WriteString(strings.Replace(childrenString, "\n", "\n    ", -1))
		buf.WriteString("\n")
	}

	buf.WriteString("}")

	return buf.String()
}
