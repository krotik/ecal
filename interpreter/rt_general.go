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
	"fmt"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

// Base Runtime
// ============

/*
baseRuntime models a base runtime component which provides the essential fields and functions.
*/
type baseRuntime struct {
	instanceID string               // Unique identifier (should be used when instance state is stored)
	erp        *ECALRuntimeProvider // Runtime provider
	node       *parser.ASTNode      // AST node which this runtime component is servicing
	validated  bool
}

var instanceCounter uint64 // Global instance counter to create unique identifiers for every runtime component instance

/*
Validate this node and all its child nodes.
*/
func (rt *baseRuntime) Validate() error {
	rt.validated = true

	// Validate all children

	for _, child := range rt.node.Children {
		if err := child.Runtime.Validate(); err != nil {
			return err
		}
	}

	return nil
}

/*
Eval evaluate this runtime component.
*/
func (rt *baseRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var err error

	errorutil.AssertTrue(rt.validated, "Runtime component has not been validated - please call Validate() before Eval()")

	if rt.erp.Debugger != nil {
		err = rt.erp.Debugger.VisitState(rt.node, vs, tid)
		rt.erp.Debugger.SetLockingState(rt.erp.MutexeOwners, rt.erp.MutexLog)
		rt.erp.Debugger.SetThreadPool(rt.erp.Processor.ThreadPool())
	}

	return nil, err
}

/*
newBaseRuntime returns a new instance of baseRuntime.
*/
func newBaseRuntime(erp *ECALRuntimeProvider, node *parser.ASTNode) *baseRuntime {
	instanceCounter++
	return &baseRuntime{fmt.Sprint(instanceCounter), erp, node, false}
}

// Void Runtime
// ============

/*
voidRuntime is a special runtime for constructs which are only evaluated as part
of other components.
*/
type voidRuntime struct {
	*baseRuntime
}

/*
voidRuntimeInst returns a new runtime component instance.
*/
func voidRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &voidRuntime{newBaseRuntime(erp, node)}
}

/*
Validate this node and all its child nodes.
*/
func (rt *voidRuntime) Validate() error {
	return rt.baseRuntime.Validate()
}

/*
Eval evaluate this runtime component.
*/
func (rt *voidRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	return rt.baseRuntime.Eval(vs, is, tid)
}

// Import Runtime
// ==============

/*
importRuntime handles import statements.
*/
type importRuntime struct {
	*baseRuntime
}

/*
importRuntimeInst returns a new runtime component instance.
*/
func importRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &importRuntime{newBaseRuntime(erp, node)}
}

/*
Validate this node and all its child nodes.
*/
func (rt *importRuntime) Validate() error {
	return rt.baseRuntime.Validate()
}

/*
Eval evaluate this runtime component.
*/
func (rt *importRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if rt.erp.ImportLocator == nil {
		err = rt.erp.NewRuntimeError(util.ErrRuntimeError, "No import locator was specified", rt.node)
	}

	if err == nil {

		var importPath interface{}
		if importPath, err = rt.node.Children[0].Runtime.Eval(vs, is, tid); err == nil {

			var codeText string
			if codeText, err = rt.erp.ImportLocator.Resolve(fmt.Sprint(importPath)); err == nil {
				var ast *parser.ASTNode

				if ast, err = parser.ParseWithRuntime(fmt.Sprint(importPath), codeText, rt.erp); err == nil {
					if err = ast.Runtime.Validate(); err == nil {

						ivs := scope.NewScope(scope.GlobalScope)
						if _, err = ast.Runtime.Eval(ivs, make(map[string]interface{}), tid); err == nil {
							irt := rt.node.Children[1].Runtime.(*identifierRuntime)
							irt.Set(vs, is, tid, scope.ToObject(ivs))
						}
					}
				}
			}
		}
	}

	return nil, err
}

// Not Implemented Runtime
// =======================

/*
invalidRuntime is a special runtime for not implemented constructs.
*/
type invalidRuntime struct {
	*baseRuntime
}

/*
invalidRuntimeInst returns a new runtime component instance.
*/
func invalidRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &invalidRuntime{newBaseRuntime(erp, node)}
}

/*
Validate this node and all its child nodes.
*/
func (rt *invalidRuntime) Validate() error {
	err := rt.baseRuntime.Validate()
	if err == nil {
		err = rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
			fmt.Sprintf("Unknown node: %s", rt.node.Name), rt.node)
	}
	return err
}

/*
Eval evaluate this runtime component.
*/
func (rt *invalidRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)
	if err == nil {
		err = rt.erp.NewRuntimeError(util.ErrInvalidConstruct, fmt.Sprintf("Unknown node: %s", rt.node.Name), rt.node)
	}
	return nil, err
}

// General Operator Runtime
// ========================

/*
operatorRuntime is a general operator operation. Used for embedding.
*/
type operatorRuntime struct {
	*baseRuntime
}

/*
errorDetailString produces a detail string for errors.
*/
func (rt *operatorRuntime) errorDetailString(token *parser.LexToken, opVal interface{}) string {
	if !token.Identifier {
		return token.Val
	}

	if opVal == nil {
		opVal = "NULL"
	}

	return fmt.Sprintf("%v=%v", token.Val, opVal)
}

/*
numVal returns a transformed number value.
*/
func (rt *operatorRuntime) numVal(op func(float64) interface{}, vs parser.Scope,
	is map[string]interface{}, tid uint64) (interface{}, error) {

	var ret interface{}

	errorutil.AssertTrue(len(rt.node.Children) == 1,
		fmt.Sprint("Operation requires 1 operand", rt.node))

	res, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
	if err == nil {

		// Check if the value is a number

		resNum, ok := res.(float64)

		if !ok {

			// Produce a runtime error if the value is not a number

			return nil, rt.erp.NewRuntimeError(util.ErrNotANumber,
				rt.errorDetailString(rt.node.Children[0].Token, res), rt.node.Children[0])
		}

		ret = op(resNum)
	}

	return ret, err
}

/*
boolVal returns a transformed boolean value.
*/
func (rt *operatorRuntime) boolVal(op func(bool) interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	var ret interface{}

	errorutil.AssertTrue(len(rt.node.Children) == 1,
		fmt.Sprint("Operation requires 1 operand", rt.node))

	res, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
	if err == nil {

		resBool, ok := res.(bool)

		if !ok {
			return nil, rt.erp.NewRuntimeError(util.ErrNotABoolean,
				rt.errorDetailString(rt.node.Children[0].Token, res), rt.node.Children[0])
		}

		ret = op(resBool)
	}

	return ret, err
}

/*
numOp executes an operation on two number values.
*/
func (rt *operatorRuntime) numOp(op func(float64, float64) interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var ok bool
	var res1, res2 interface{}
	var err error

	errorutil.AssertTrue(len(rt.node.Children) == 2,
		fmt.Sprint("Operation requires 2 operands", rt.node))

	if res1, err = rt.node.Children[0].Runtime.Eval(vs, is, tid); err == nil {
		if res2, err = rt.node.Children[1].Runtime.Eval(vs, is, tid); err == nil {
			var res1Num, res2Num float64

			if res1Num, ok = res1.(float64); !ok {
				err = rt.erp.NewRuntimeError(util.ErrNotANumber,
					rt.errorDetailString(rt.node.Children[0].Token, res1), rt.node.Children[0])

			} else {
				if res2Num, ok = res2.(float64); !ok {
					err = rt.erp.NewRuntimeError(util.ErrNotANumber,
						rt.errorDetailString(rt.node.Children[1].Token, res2), rt.node.Children[1])

				} else {

					return op(res1Num, res2Num), err
				}
			}
		}
	}

	return nil, err
}

/*
genOp executes an operation on two general values.
*/
func (rt *operatorRuntime) genOp(op func(interface{}, interface{}) interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	var ret interface{}

	errorutil.AssertTrue(len(rt.node.Children) == 2,
		fmt.Sprint("Operation requires 2 operands", rt.node))

	res1, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
	if err == nil {
		var res2 interface{}

		if res2, err = rt.node.Children[1].Runtime.Eval(vs, is, tid); err == nil {
			ret = op(res1, res2)
		}
	}

	return ret, err
}

/*
strOp executes an operation on two string values.
*/
func (rt *operatorRuntime) strOp(op func(string, string) interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	var ret interface{}

	errorutil.AssertTrue(len(rt.node.Children) == 2,
		fmt.Sprint("Operation requires 2 operands", rt.node))

	res1, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
	if err == nil {
		var res2 interface{}

		if res2, err = rt.node.Children[1].Runtime.Eval(vs, is, tid); err == nil {
			ret = op(fmt.Sprint(res1), fmt.Sprint(res2))
		}
	}

	return ret, err
}

/*
boolOp executes an operation on two boolean values.
*/
func (rt *operatorRuntime) boolOp(op func(bool, bool) interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	var res interface{}

	errorutil.AssertTrue(len(rt.node.Children) == 2,
		fmt.Sprint("Operation requires 2 operands", rt.node))

	res1, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
	if err == nil {
		var res2 interface{}

		if res2, err = rt.node.Children[1].Runtime.Eval(vs, is, tid); err == nil {

			res1bool, ok := res1.(bool)
			if !ok {
				return nil, rt.erp.NewRuntimeError(util.ErrNotABoolean,
					rt.errorDetailString(rt.node.Children[0].Token, res1), rt.node.Children[0])
			}

			res2bool, ok := res2.(bool)
			if !ok {
				return nil, rt.erp.NewRuntimeError(util.ErrNotABoolean,
					rt.errorDetailString(rt.node.Children[1].Token, res2), rt.node.Children[0])
			}

			res = op(res1bool, res2bool)
		}
	}

	return res, err
}

/*
listOp executes an operation on a value and a list.
*/
func (rt *operatorRuntime) listOp(op func(interface{}, []interface{}) interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	var res interface{}

	errorutil.AssertTrue(len(rt.node.Children) == 2,
		fmt.Sprint("Operation requires 2 operands", rt.node))

	res1, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
	if err == nil {
		var res2 interface{}

		if res2, err = rt.node.Children[1].Runtime.Eval(vs, is, tid); err == nil {

			res2list, ok := res2.([]interface{})
			if !ok {
				err = rt.erp.NewRuntimeError(util.ErrNotAList,
					rt.errorDetailString(rt.node.Children[1].Token, res2), rt.node.Children[0])
			} else {
				res = op(res1, res2list)
			}
		}
	}

	return res, err
}
