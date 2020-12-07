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
	"strings"

	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/stdlib"
	"devt.de/krotik/ecal/util"
)

/*
identifierRuntime is the runtime component for identifiers.
*/
type identifierRuntime struct {
	*baseRuntime
}

/*
identifierRuntimeInst returns a new runtime component instance.
*/
func identifierRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &identifierRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *identifierRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}
	_, err := rt.baseRuntime.Eval(vs, is, tid)
	if err == nil {
		res, err = rt.resolveValue(vs, is, tid, rt.node)
	}
	return res, err
}

/*
resolveValue resolves the value of this identifier.
*/
func (rt *identifierRuntime) resolveValue(vs parser.Scope, is map[string]interface{}, tid uint64, node *parser.ASTNode) (interface{}, error) {
	var anode *parser.ASTNode
	var astring string
	var result interface{}
	var err error

	functionResolved := func(astring string, rnode *parser.ASTNode) *parser.ASTNode {

		res := &parser.ASTNode{ // Create a dummy identifier which models the value evaluation so far
			Name: parser.NodeIDENTIFIER,
			Token: &parser.LexToken{
				ID:         node.Token.ID,
				Identifier: node.Token.Identifier,
				Lline:      node.Token.Lline,
				Lpos:       node.Token.Lpos,
				Pos:        node.Token.Pos,
				Val:        strings.Replace(astring, ".", ">", -1),
			},
			Children: nil,
		}

		for i, c := range rnode.Children {
			if c.Name == parser.NodeFUNCCALL {
				res.Children = rnode.Children[i+1:]
			}
		}

		return res
	}

	anode, astring, err = buildAccessString(rt.erp, vs, is, tid, node, node.Token.Val)

	if len(node.Children) == 0 {

		// Simple case we just have a variable

		result, _, err = vs.GetValue(node.Token.Val)

	} else if cval, ok := stdlib.GetStdlibConst(astring); ok {

		result = cval

	} else {

		if rerr, ok := err.(*util.RuntimeError); err == nil || ok && rerr.Type == util.ErrInvalidConstruct {
			funcCallInAccessStringExecuted := ok && rerr.Type == util.ErrInvalidConstruct

			if result, _, err = vs.GetValue(astring); err == nil {

				if funcCallInAccessStringExecuted {

					result, err = rt.resolveFunction(astring, vs, is, tid, rerr.Node, result, err)

					node = functionResolved(astring, anode)

					if len(node.Children) > 0 {

						// We have more identifiers after the func call - there is more to do ...

						vs = scope.NewScope("funcresult")
						vs.SetValue(node.Token.Val, result)

						result, err = rt.resolveValue(vs, is, tid, node)
					}
				} else {

					result, err = rt.resolveFunction(astring, vs, is, tid, node, result, err)
				}
			}
		}
	}

	return result, err
}

/*
resolveFunction execute function calls and return the result.
*/
func (rt *identifierRuntime) resolveFunction(astring string, vs parser.Scope, is map[string]interface{},
	tid uint64, node *parser.ASTNode, result interface{}, err error) (interface{}, error) {

	is["erp"] = rt.erp      // All functions have access to the ECAL Runtime Provider
	is["astnode"] = rt.node // ... and the AST node

	for _, funccall := range node.Children {

		if funccall.Name == parser.NodeFUNCCALL {

			funcObj, ok := rt.resolveFunctionObject(astring, result)

			if ok {
				var args []interface{}

				// Collect the parameter values

				for _, c := range funccall.Children {
					var val interface{}

					if err == nil {
						val, err = c.Runtime.Eval(vs, make(map[string]interface{}), tid)
						args = append(args, val)
					}
				}

				if err == nil {
					result, err = rt.executeFunction(astring, funcObj, args, vs, is, tid, node)
				}

			} else {

				err = rt.erp.NewRuntimeError(util.ErrUnknownConstruct,
					fmt.Sprintf("Unknown function: %v", node.Token.Val), node)
			}

			break
		}
	}

	return result, err
}

/*
resolveFunctionObject will resolve a given string or object into a concrete ECAL function.
*/
func (rt *identifierRuntime) resolveFunctionObject(astring string, result interface{}) (util.ECALFunction, bool) {
	var funcObj util.ECALFunction

	ok := astring == "log" || astring == "error" || astring == "debug"

	if !ok {

		funcObj, ok = result.(util.ECALFunction)

		if !ok {

			// Check for stdlib function

			funcObj, ok = stdlib.GetStdlibFunc(astring)

			if !ok {

				// Check for inbuild function

				funcObj, ok = InbuildFuncMap[astring]
			}
		}
	}

	return funcObj, ok
}

/*
executeFunction executes a function call with a given list of arguments and return the result.
*/
func (rt *identifierRuntime) executeFunction(astring string, funcObj util.ECALFunction, args []interface{},
	vs parser.Scope, is map[string]interface{}, tid uint64, node *parser.ASTNode) (interface{}, error) {

	var result interface{}
	var err error

	if stringutil.IndexOf(astring, []string{"log", "error", "debug"}) != -1 {

		// Convert non-string structures

		for i, a := range args {
			if _, ok := a.(string); !ok {
				args[i] = stringutil.ConvertToPrettyString(a)
			}
		}

		if astring == "log" {
			rt.erp.Logger.LogInfo(args...)
		} else if astring == "error" {
			rt.erp.Logger.LogError(args...)
		} else if astring == "debug" {
			rt.erp.Logger.LogDebug(args...)
		}

	} else {

		if rt.erp.Debugger != nil {
			rt.erp.Debugger.VisitStepInState(node, vs, tid)
		}

		// Execute the function

		result, err = funcObj.Run(rt.instanceID, vs, is, tid, args)

		if rt.erp.Debugger != nil {
			rt.erp.Debugger.VisitStepOutState(node, vs, tid, err)
		}

		_, ok1 := err.(*util.RuntimeError)
		_, ok2 := err.(*util.RuntimeErrorWithDetail)

		if err != nil && !ok1 && !ok2 {

			// Convert into a proper runtime error if necessary

			rerr := rt.erp.NewRuntimeError(util.ErrRuntimeError,
				err.Error(), node).(*util.RuntimeError)

			if stringutil.IndexOf(err.Error(), []string{util.ErrIsIterator.Error(),
				util.ErrEndOfIteration.Error(), util.ErrContinueIteration.Error()}) != -1 {

				rerr.Type = err
			}

			err = rerr
		}

		if tr, ok := err.(util.TraceableRuntimeError); ok {

			// Add tracing information to the error

			tr.AddTrace(rt.node)
		}
	}

	return result, err
}

/*
Set sets a value to this identifier.
*/
func (rt *identifierRuntime) Set(vs parser.Scope, is map[string]interface{}, tid uint64, value interface{}) error {
	var err error

	if len(rt.node.Children) == 0 {

		// Simple case we just have a variable

		err = vs.SetValue(rt.node.Token.Val, value)

	} else {
		var as string

		_, as, err = buildAccessString(rt.erp, vs, is, tid, rt.node, rt.node.Token.Val)

		if err == nil {

			// Collect all the children and find the right spot

			err = vs.SetValue(as, value)
		}
	}

	return err
}

/*
buildAccessString builds an access string using a given node and a prefix.
*/
func buildAccessString(erp *ECALRuntimeProvider, vs parser.Scope, is map[string]interface{},
	tid uint64, node *parser.ASTNode, prefix string) (*parser.ASTNode, string, error) {

	var err error
	res := prefix

	for i, c := range node.Children {

		if err == nil {

			// The unexpected construct error is used in two ways:
			// 1. Error message when a function call is used on the left hand of
			// an assignment.
			// 2. Signalling there is a function call involved on the right hand
			// of an assignment.

			if c.Name == parser.NodeCOMPACCESS {
				var val interface{}
				val, err = c.Children[0].Runtime.Eval(vs, is, tid)
				res = fmt.Sprintf("%v.%v", res, val)

				if len(node.Children) > i+1 && node.Children[i+1].Name == parser.NodeFUNCCALL {

					err = erp.NewRuntimeError(util.ErrInvalidConstruct,
						"Unexpected construct", node)
					break
				}

			} else if c.Name == parser.NodeIDENTIFIER {

				res = fmt.Sprintf("%v.%v", res, c.Token.Val)

				if len(c.Children) > 0 && c.Children[0].Name == parser.NodeFUNCCALL {
					node = c

					err = erp.NewRuntimeError(util.ErrInvalidConstruct,
						"Unexpected construct", node)
					break
				}

				node, res, err = buildAccessString(erp, vs, is, tid, c, res)
			}
		}
	}

	return node, res, err
}
