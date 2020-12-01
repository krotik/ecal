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
	"sync"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/sortutil"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

// Statements Runtime
// ==================

/*
statementsRuntime is the runtime component for sequences of statements.
*/
type statementsRuntime struct {
	*baseRuntime
}

/*
statementsRuntimeInst returns a new runtime component instance.
*/
func statementsRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &statementsRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *statementsRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		for _, child := range rt.node.Children {
			if res, err = child.Runtime.Eval(vs, is, tid); err != nil {
				return nil, err
			}
		}
	}

	return res, err
}

// Condition statement
// ===================

/*
ifRuntime is the runtime for the if condition statement.
*/
type ifRuntime struct {
	*baseRuntime
}

/*
ifRuntimeInst returns a new runtime component instance.
*/
func ifRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &ifRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *ifRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		// Create a new variable scope

		vs = vs.NewChild(scope.NameFromASTNode(rt.node))

		for offset := 0; offset < len(rt.node.Children); offset += 2 {
			var guardres interface{}

			// Evaluate guard

			if err == nil {
				guardres, err = rt.node.Children[offset].Runtime.Eval(vs, is, tid)

				if err == nil && guardres.(bool) {

					// The guard holds true so we execture its statements

					return rt.node.Children[offset+1].Runtime.Eval(vs, is, tid)
				}
			}
		}
	}

	return nil, err
}

// Guard Runtime
// =============

/*
guardRuntime is the runtime for any guard condition (used in if, for, etc...).
*/
type guardRuntime struct {
	*baseRuntime
}

/*
guardRuntimeInst returns a new runtime component instance.
*/
func guardRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &guardRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *guardRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		var ret interface{}

		// Evaluate the condition

		ret, err = rt.node.Children[0].Runtime.Eval(vs, is, tid)

		// Guard returns always a boolean

		res = ret != nil && ret != false && ret != 0
	}

	return res, err
}

// Loop statement
// ==============

/*
loopRuntime is the runtime for the loop statement (for).
*/
type loopRuntime struct {
	*baseRuntime
	leftInVarName []string
}

/*
loopRuntimeInst returns a new runtime component instance.
*/
func loopRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &loopRuntime{newBaseRuntime(erp, node), nil}
}

/*
Validate this node and all its child nodes.
*/
func (rt *loopRuntime) Validate() error {

	err := rt.baseRuntime.Validate()

	if err == nil {

		if rt.node.Children[0].Name == parser.NodeIN {

			inVar := rt.node.Children[0].Children[0]

			if inVar.Name == parser.NodeIDENTIFIER {

				if len(inVar.Children) != 0 {
					return rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
						"Must have a simple variable on the left side of the In expression", rt.node)
				}

				rt.leftInVarName = []string{inVar.Token.Val}

			} else if inVar.Name == parser.NodeLIST {
				rt.leftInVarName = make([]string, 0, len(inVar.Children))

				for _, child := range inVar.Children {
					if child.Name != parser.NodeIDENTIFIER || len(child.Children) != 0 {
						return rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
							"Must have a list of simple variables on the left side of the In expression", rt.node)
					}

					rt.leftInVarName = append(rt.leftInVarName, child.Token.Val)
				}
			}
		}
	}

	return err
}

/*
Eval evaluate this runtime component.
*/
func (rt *loopRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		var guardres interface{}

		// Create a new variable scope

		vs = vs.NewChild(scope.NameFromASTNode(rt.node))

		// Create a new instance scope - elements in each loop iteration start from scratch

		is = make(map[string]interface{})

		if rt.node.Children[0].Name == parser.NodeGUARD {

			// Evaluate guard

			guardres, err = rt.node.Children[0].Runtime.Eval(vs, is, tid)

			for err == nil && guardres.(bool) {

				// Execute block

				_, err = rt.node.Children[1].Runtime.Eval(vs, is, tid)

				// Check for continue

				if err != nil {
					if eoi, ok := err.(*util.RuntimeError); ok {
						if eoi.Type == util.ErrContinueIteration {
							err = nil
						}
					}
				}

				if err == nil {

					// Evaluate guard

					guardres, err = rt.node.Children[0].Runtime.Eval(vs, is, tid)
				}
			}

		} else if rt.node.Children[0].Name == parser.NodeIN {
			var iterator func() (interface{}, error)
			var val interface{}

			it := rt.node.Children[0].Children[1]

			val, err = it.Runtime.Eval(vs, is, tid)

			// Create an iterator object

			if rterr, ok := err.(*util.RuntimeError); ok && rterr.Type == util.ErrIsIterator {

				// We got an iterator - all subsequent calls will return values

				iterator = func() (interface{}, error) {
					return it.Runtime.Eval(vs, is, tid)
				}
				err = nil

			} else {

				// We got a value over which we need to iterate

				if valList, isList := val.([]interface{}); isList {

					index := -1
					end := len(valList)

					iterator = func() (interface{}, error) {
						index++
						if index >= end {
							return nil, rt.erp.NewRuntimeError(util.ErrEndOfIteration, "", rt.node)
						}
						return valList[index], nil
					}

				} else if valMap, isMap := val.(map[interface{}]interface{}); isMap {
					var keys []interface{}

					index := -1

					for k := range valMap {
						keys = append(keys, k)
					}
					end := len(keys)

					// Try to sort according to string value

					sortutil.InterfaceStrings(keys)

					iterator = func() (interface{}, error) {
						index++
						if index >= end {
							return nil, rt.erp.NewRuntimeError(util.ErrEndOfIteration, "", rt.node)
						}
						key := keys[index]
						return []interface{}{key, valMap[key]}, nil
					}

				} else {

					// A single value will do exactly one iteration

					index := -1

					iterator = func() (interface{}, error) {
						index++
						if index > 0 {
							return nil, rt.erp.NewRuntimeError(util.ErrEndOfIteration, "", rt.node)
						}
						return val, nil
					}
				}
			}

			vars := rt.leftInVarName

			for err == nil {
				var res interface{}

				res, err = iterator()

				if err != nil {
					if eoi, ok := err.(*util.RuntimeError); ok {
						if eoi.Type == util.ErrIsIterator {
							err = nil
						}
					}
				}

				if err == nil {

					if len(vars) == 1 {
						err = vs.SetValue(vars[0], res)

					} else if resList, ok := res.([]interface{}); ok {

						if len(vars) != len(resList) {
							err = fmt.Errorf("Assigned number of variables is different to "+
								"number of values (%v variables vs %v values)",
								len(vars), len(resList))
						}

						if err == nil {
							for i, v := range vars {
								if err == nil {
									err = vs.SetValue(v, resList[i])
								}
							}
						}

					} else {

						err = fmt.Errorf("Result for loop variable is not a list (value is %v)", res)
					}

					if err != nil {
						return nil, rt.erp.NewRuntimeError(util.ErrRuntimeError,
							err.Error(), rt.node)
					}

					// Execute block

					_, err = rt.node.Children[1].Runtime.Eval(vs, is, tid)
				}

				// Check for continue

				if err != nil {
					if eoi, ok := err.(*util.RuntimeError); ok {
						if eoi.Type == util.ErrContinueIteration {
							err = nil
						}
					}
				}
			}

			// Check for end of iteration error

			if eoi, ok := err.(*util.RuntimeError); ok {
				if eoi.Type == util.ErrEndOfIteration {
					err = nil
				}
			}
		}
	}

	return nil, err
}

// Break statement
// ===============

/*
breakRuntime is the runtime for the break statement.
*/
type breakRuntime struct {
	*baseRuntime
}

/*
breakRuntimeInst returns a new runtime component instance.
*/
func breakRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &breakRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *breakRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		err = rt.erp.NewRuntimeError(util.ErrEndOfIteration, "", rt.node)
	}

	return nil, err
}

// Continue statement
// ==================

/*
continueRuntime is the runtime for the continue statement.
*/
type continueRuntime struct {
	*baseRuntime
}

/*
continueRuntimeInst returns a new runtime component instance.
*/
func continueRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &continueRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *continueRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		err = rt.erp.NewRuntimeError(util.ErrContinueIteration, "", rt.node)
	}

	return nil, err
}

// Try Runtime
// ===========

/*
tryRuntime is the runtime for try blocks.
*/
type tryRuntime struct {
	*baseRuntime
}

/*
tryRuntimeInst returns a new runtime component instance.
*/
func tryRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &tryRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *tryRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	evalExcept := func(errObj map[interface{}]interface{}, except *parser.ASTNode) bool {
		ret := false

		if len(except.Children) == 1 {

			// We only have statements - any exception is handled here

			evs := vs.NewChild(scope.NameFromASTNode(except))

			except.Children[0].Runtime.Eval(evs, is, tid)

			ret = true

		} else if len(except.Children) == 2 {

			// We have statements and the error object is available - any exception is handled here

			evs := vs.NewChild(scope.NameFromASTNode(except))
			evs.SetValue(except.Children[0].Token.Val, errObj)

			except.Children[1].Runtime.Eval(evs, is, tid)

			ret = true

		} else {
			errorVar := ""

			for i := 0; i < len(except.Children); i++ {
				child := except.Children[i]

				if !ret && child.Name == parser.NodeSTRING {
					exceptError, evalErr := child.Runtime.Eval(vs, is, tid)

					// If we fail evaluating the string we panic as otherwise
					// we would need to generate a new error while trying to handle another error
					errorutil.AssertOk(evalErr)

					ret = exceptError == fmt.Sprint(errObj["type"])

				} else if ret && child.Name == parser.NodeAS {
					errorVar = child.Children[0].Token.Val

				} else if ret && child.Name == parser.NodeSTATEMENTS {
					evs := vs.NewChild(scope.NameFromASTNode(except))

					if errorVar != "" {
						evs.SetValue(errorVar, errObj)
					}

					child.Runtime.Eval(evs, is, tid)
				}
			}
		}

		return ret
	}

	// Make sure the finally block is executed in any case

	if finally := rt.node.Children[len(rt.node.Children)-1]; finally.Name == parser.NodeFINALLY {
		fvs := vs.NewChild(scope.NameFromASTNode(finally))
		defer finally.Children[0].Runtime.Eval(fvs, is, tid)
	}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		tvs := vs.NewChild(scope.NameFromASTNode(rt.node))

		res, err = rt.node.Children[0].Runtime.Eval(tvs, is, tid)

		// Evaluate except clauses

		if err != nil {
			errObj := map[interface{}]interface{}{
				"type":  "UnexpectedError",
				"error": err.Error(),
			}

			if rtError, ok := err.(*util.RuntimeError); ok {
				errObj["type"] = rtError.Type.Error()
				errObj["detail"] = rtError.Detail
				errObj["pos"] = rtError.Pos
				errObj["line"] = rtError.Line
				errObj["source"] = rtError.Source

			} else if rtError, ok := err.(*util.RuntimeErrorWithDetail); ok {
				errObj["type"] = rtError.Type.Error()
				errObj["detail"] = rtError.Detail
				errObj["pos"] = rtError.Pos
				errObj["line"] = rtError.Line
				errObj["source"] = rtError.Source
				errObj["data"] = rtError.Data
			}

			if te, ok := err.(util.TraceableRuntimeError); ok {

				if ts := te.GetTraceString(); ts != nil {
					errObj["trace"] = ts
				}
			}

			res = nil

			for i := 1; i < len(rt.node.Children); i++ {
				if child := rt.node.Children[i]; child.Name == parser.NodeEXCEPT {
					if evalExcept(errObj, child) {
						err = nil
						break
					}
				}
			}
		}
	}

	return res, err
}

// Mutex Runtime
// =============

/*
mutexRuntime is the runtime for mutex blocks.
*/
type mutexRuntime struct {
	*baseRuntime
}

/*
mutexRuntimeInst returns a new runtime component instance.
*/
func mutexRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &mutexRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *mutexRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		// Get the name of the mutex

		name := rt.node.Children[0].Token.Val

		mutex, ok := rt.erp.Mutexes[name]
		if !ok {
			mutex = &sync.Mutex{}
			rt.erp.Mutexes[name] = mutex
		}

		tvs := vs.NewChild(scope.NameFromASTNode(rt.node))

		mutex.Lock()
		defer mutex.Unlock()

		res, err = rt.node.Children[0].Runtime.Eval(tvs, is, tid)
	}

	return res, err
}
