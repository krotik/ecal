/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package interpreter

import (
	"encoding/json"
	"fmt"
	"strings"

	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

/*
returnRuntime is a special runtime for return statements in functions.
*/
type returnRuntime struct {
	*baseRuntime
}

/*
voidRuntimeInst returns a new runtime component instance.
*/
func returnRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &returnRuntime{newBaseRuntime(erp, node)}
}

/*
Validate this node and all its child nodes.
*/
func (rt *returnRuntime) Validate() error {
	return rt.baseRuntime.Validate()
}

/*
Eval evaluate this runtime component.
*/
func (rt *returnRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		var res interface{}

		if res, err = rt.node.Children[0].Runtime.Eval(vs, is, tid); err == nil {
			rerr := rt.erp.NewRuntimeError(util.ErrReturn, fmt.Sprintf("Return value: %v", res), rt.node)
			err = &returnValue{
				rerr.(*util.RuntimeError),
				res,
			}

		}
	}

	return nil, err
}

type returnValue struct {
	*util.RuntimeError
	returnValue interface{}
}

/*
funcRuntime is the runtime component for function declarations.
*/
type funcRuntime struct {
	*baseRuntime
}

/*
funcRuntimeInst returns a new runtime component instance.
*/
func funcRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &funcRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *funcRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var fc interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		name := ""

		if rt.node.Children[0].Name == parser.NodeIDENTIFIER {
			name = rt.node.Children[0].Token.Val
		}

		fc = &function{name, nil, nil, rt.node, vs}

		if name != "" {
			vs.SetValue(name, fc)
		}
	}

	return fc, err
}

/*
function models a function in ECAL. It can have a context object attached - this.
*/
type function struct {
	name          string
	super         []interface{}   // Super function pointer
	this          interface{}     // Function context
	declaration   *parser.ASTNode // Function declaration node
	declarationVS parser.Scope    // Function declaration scope
}

/*
Run executes this function. The function is called with parameters and might also
have a reference to a context state - this.
*/
func (f *function) Run(instanceID string, vs parser.Scope, is map[string]interface{}, tid uint64, args []interface{}) (interface{}, error) {
	var res interface{}
	var err error

	nameOffset := 0
	if f.declaration.Children[0].Name == parser.NodeIDENTIFIER {
		nameOffset = 1
	}
	params := f.declaration.Children[0+nameOffset].Children
	body := f.declaration.Children[1+nameOffset]

	// Create varscope for the body - not a child scope but a new root

	fvs := scope.NewScope(fmt.Sprintf("%v %v", scope.FuncPrefix, f.name))

	if f.this != nil {
		fvs.SetValue("this", f.this)
	}
	if f.super != nil {
		fvs.SetValue("super", f.super)
	}

	for i, p := range params {
		var name string
		var val interface{}

		if err == nil {
			name = ""

			if p.Name == parser.NodeIDENTIFIER {
				name = p.Token.Val

				if i < len(args) {
					val = args[i]
				}
			} else if p.Name == parser.NodePRESET {
				name = p.Children[0].Token.Val

				if i < len(args) {
					val = args[i]
				} else {
					val, err = p.Children[1].Runtime.Eval(vs, is, tid)
				}
			}

			if name != "" {
				fvs.SetValue(name, val)
			}
		}
	}

	if err == nil {

		scope.SetParentOfScope(fvs, f.declarationVS)

		res, err = body.Runtime.Eval(fvs, make(map[string]interface{}), tid)

		// Check for return value (delivered as error object)

		if rval, ok := err.(*returnValue); ok {
			res = rval.returnValue
			err = nil
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (f *function) DocString() (string, error) {

	if len(f.declaration.Meta) > 0 {
		return strings.TrimSpace(f.declaration.Meta[0].Value()), nil
	}

	return fmt.Sprintf("Declared function: %v (%v)", f.name, f.declaration.Token.PosString()), nil
}

/*
String returns a string representation of this function.
*/
func (f *function) String() string {
	return fmt.Sprintf("ecal.function: %v (%v)", f.name, f.declaration.Token.PosString())
}

/*
MarshalJSON returns a string representation of this function - a function cannot
be JSON encoded.
*/
func (f *function) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}
