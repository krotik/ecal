/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package parser

/*
RuntimeProvider provides runtime components for a parse tree.
*/
type RuntimeProvider interface {

	/*
	   Runtime returns a runtime component for a given ASTNode.
	*/
	Runtime(node *ASTNode) Runtime
}

/*
Runtime provides the runtime for an ASTNode.
*/
type Runtime interface {

	/*
	   Validate this runtime component and all its child components.
	*/
	Validate() error

	/*
		Eval evaluate this runtime component. It gets passed the current variable
		scope an instance state and a thread ID.

		The instance state is created per execution instance and can be used
		for generator functions to store their current state. It gets replaced
		by a new object in certain situations (e.g. a function call).

		The thread ID can be used to identify a running process.
	*/
	Eval(Scope, map[string]interface{}, uint64) (interface{}, error)
}

/*
Scope models an environment which stores data.
*/
type Scope interface {

	/*
	   Name returns the name of this scope.
	*/
	Name() string

	/*
	   NewChild creates a new child scope.
	*/
	NewChild(name string) Scope

	/*
		Clear clears this scope of all stored values. This will clear children scopes
		but not remove parent scopes.
	*/
	Clear()

	/*
	   Parent returns the parent scope or nil.
	*/
	Parent() Scope

	/*
	   SetValue sets a new value for a variable.
	*/
	SetValue(varName string, varValue interface{}) error

	/*
	   SetLocalValue sets a new value for a local variable.
	*/
	SetLocalValue(varName string, varValue interface{}) error

	/*
	   GetValue gets the current value of a variable.
	*/
	GetValue(varName string) (interface{}, bool, error)

	/*
	   String returns a string representation of this scope.
	*/
	String() string

	/*
	   ToJSONObject returns this ASTNode and all its children as a JSON object.
	*/
	ToJSONObject() map[string]interface{}
}
