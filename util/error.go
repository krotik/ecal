/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

/*
Package util contains utility definitions and functions for the event condition language ECAL.
*/
package util

import (
	"encoding/json"
	"errors"
	"fmt"

	"devt.de/krotik/ecal/parser"
)

/*
TraceableRuntimeError can record and show a stack trace.
*/
type TraceableRuntimeError interface {
	error

	/*
		AddTrace adds a trace step.
	*/
	AddTrace(*parser.ASTNode)

	/*
		GetTrace returns the current stacktrace.
	*/
	GetTrace() []*parser.ASTNode

	/*
		GetTrace returns the current stacktrace as a string.
	*/
	GetTraceString() []string
}

/*
RuntimeError is a runtime related error.
*/
type RuntimeError struct {
	Source string            // Name of the source which was given to the parser
	Type   error             // Error type (to be used for equal checks)
	Detail string            // Details of this error
	Node   *parser.ASTNode   // AST Node where the error occurred
	Line   int               // Line of the error
	Pos    int               // Position of the error
	Trace  []*parser.ASTNode // Stacktrace
}

/*
Runtime related error types.
*/
var (
	ErrRuntimeError     = errors.New("Runtime error")
	ErrUnknownConstruct = errors.New("Unknown construct")
	ErrInvalidConstruct = errors.New("Invalid construct")
	ErrInvalidState     = errors.New("Invalid state")
	ErrVarAccess        = errors.New("Cannot access variable")
	ErrNotANumber       = errors.New("Operand is not a number")
	ErrNotABoolean      = errors.New("Operand is not a boolean")
	ErrNotAList         = errors.New("Operand is not a list")
	ErrNotAMap          = errors.New("Operand is not a map")
	ErrNotAListOrMap    = errors.New("Operand is not a list nor a map")
	ErrSink             = errors.New("Error in sink")

	// ErrReturn is not an error. It is used to return when executing a function
	ErrReturn = errors.New("*** return ***")

	// Error codes for loop operations
	ErrIsIterator        = errors.New("Function is an iterator")
	ErrEndOfIteration    = errors.New("End of iteration was reached")
	ErrContinueIteration = errors.New("End of iteration step - Continue iteration")
)

/*
NewRuntimeError creates a new RuntimeError object.
*/
func NewRuntimeError(source string, t error, d string, node *parser.ASTNode) error {
	if node.Token != nil {
		return &RuntimeError{source, t, d, node, node.Token.Lline, node.Token.Lpos, nil}
	}
	return &RuntimeError{source, t, d, node, 0, 0, nil}
}

/*
Error returns a human-readable string representation of this error.
*/
func (re *RuntimeError) Error() string {
	ret := fmt.Sprintf("ECAL error in %s: %v (%v)", re.Source, re.Type, re.Detail)

	if re.Line != 0 {

		// Add line if available

		ret = fmt.Sprintf("%s (Line:%d Pos:%d)", ret, re.Line, re.Pos)
	}

	return ret
}

/*
AddTrace adds a trace step.
*/
func (re *RuntimeError) AddTrace(n *parser.ASTNode) {
	re.Trace = append(re.Trace, n)
}

/*
GetTrace returns the current stacktrace.
*/
func (re *RuntimeError) GetTrace() []*parser.ASTNode {
	return re.Trace
}

/*
GetTrace returns the current stacktrace as a string.
*/
func (re *RuntimeError) GetTraceString() []string {
	res := []string{}
	for _, t := range re.GetTrace() {
		pp, _ := parser.PrettyPrint(t)
		res = append(res, fmt.Sprintf("%v (%v:%v)", pp, t.Token.Lsource, t.Token.Lline))
	}
	return res
}

/*
ToJSONObject returns this RuntimeError and all its children as a JSON object.
*/
func (re *RuntimeError) ToJSONObject() map[string]interface{} {
	t := ""
	if re.Type != nil {
		t = re.Type.Error()
	}
	return map[string]interface{}{
		"Source": re.Source,
		"Type":   t,
		"Detail": re.Detail,
		"Node":   re.Node,
		"Trace":  re.Trace,
	}
}

/*
MarshalJSON serializes this RuntimeError into a JSON string.
*/
func (re *RuntimeError) MarshalJSON() ([]byte, error) {
	return json.Marshal(re.ToJSONObject())
}

/*
RuntimeErrorWithDetail is a runtime error with additional environment information.
*/
type RuntimeErrorWithDetail struct {
	*RuntimeError
	Environment parser.Scope
	Data        interface{}
}

/*
ToJSONObject returns this RuntimeErrorWithDetail and all its children as a JSON object.
*/
func (re *RuntimeErrorWithDetail) ToJSONObject() map[string]interface{} {
	res := re.RuntimeError.ToJSONObject()
	e := map[string]interface{}{}
	if re.Environment != nil {
		e = re.Environment.ToJSONObject()
	}
	res["Environment"] = e
	res["Data"] = re.Data
	return res
}

/*
MarshalJSON serializes this RuntimeErrorWithDetail into a JSON string.
*/
func (re *RuntimeErrorWithDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(re.ToJSONObject())
}
