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

	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/util"
)

/*
assignmentRuntime is the runtime component for assignment of values.
*/
type assignmentRuntime struct {
	*baseRuntime
	leftSide []*identifierRuntime
}

/*
assignmentRuntimeInst returns a new runtime component instance.
*/
func assignmentRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &assignmentRuntime{newBaseRuntime(erp, node), nil}
}

/*
Validate this node and all its child nodes.
*/
func (rt *assignmentRuntime) Validate() error {
	err := rt.baseRuntime.Validate()

	if err == nil {

		leftVar := rt.node.Children[0]

		if _, ok := leftVar.Runtime.(*letRuntime); ok {
			leftVar = leftVar.Children[0]
		}

		if leftRuntime, ok := leftVar.Runtime.(*identifierRuntime); ok {

			rt.leftSide = []*identifierRuntime{leftRuntime}

		} else if leftVar.Name == parser.NodeLIST {

			rt.leftSide = make([]*identifierRuntime, 0, len(leftVar.Children))

			for _, child := range leftVar.Children {
				childRuntime, ok := child.Runtime.(*identifierRuntime)

				if !ok {
					err = rt.erp.NewRuntimeError(util.ErrVarAccess,
						"Must have a list of variables on the left side of the assignment", rt.node)
					break
				}

				rt.leftSide = append(rt.leftSide, childRuntime)
			}

		} else {

			err = rt.erp.NewRuntimeError(util.ErrVarAccess,
				"Must have a variable or list of variables on the left side of the assignment", rt.node)
		}
	}

	return err
}

/*
Eval evaluate this runtime component.
*/
func (rt *assignmentRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		// Execute let statements on the right before evaluating the left side

		if _, err = rt.node.Children[0].Runtime.Eval(vs, is, tid); err == nil {
			var val interface{}

			val, err = rt.node.Children[1].Runtime.Eval(vs, is, tid)

			if err == nil {
				if len(rt.leftSide) == 1 {

					err = rt.leftSide[0].Set(vs, is, tid, val)

				} else if valList, ok := val.([]interface{}); ok {

					if len(rt.leftSide) != len(valList) {

						err = rt.erp.NewRuntimeError(util.ErrInvalidState,
							fmt.Sprintf("Assigned number of variables is different to "+
								"number of values (%v variables vs %v values)",
								len(rt.leftSide), len(valList)), rt.node)

					} else {

						for i, v := range rt.leftSide {

							if err = v.Set(vs, is, tid, valList[i]); err != nil {
								err = rt.erp.NewRuntimeError(util.ErrVarAccess,
									err.Error(), rt.node)
								break
							}
						}
					}

				} else {

					err = rt.erp.NewRuntimeError(util.ErrInvalidState,
						fmt.Sprintf("Result is not a list (value is %v)", val),
						rt.node)
				}
			}
		}
	}

	return nil, err
}

/*
letRuntime is the runtime component for let statements
*/
type letRuntime struct {
	*baseRuntime
	declared []*identifierRuntime
}

/*
letRuntimeInst returns a new runtime component instance.
*/
func letRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &letRuntime{newBaseRuntime(erp, node), nil}
}

/*
Validate this node and all its child nodes.
*/
func (rt *letRuntime) Validate() error {
	err := rt.baseRuntime.Validate()

	if err == nil {

		leftVar := rt.node.Children[0]

		if leftRuntime, ok := leftVar.Runtime.(*identifierRuntime); ok {

			rt.declared = []*identifierRuntime{leftRuntime}

		} else if leftVar.Name == parser.NodeLIST {

			rt.declared = make([]*identifierRuntime, 0, len(leftVar.Children))

			for _, child := range leftVar.Children {
				childRuntime, ok := child.Runtime.(*identifierRuntime)

				if !ok {
					err = rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
						"Let can only declare variables within a list", rt.node)
					break
				}

				rt.declared = append(rt.declared, childRuntime)
			}

		} else {

			err = rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
				"Let must declare a variable or list of variables", rt.node)
		}
	}

	return err
}

/*
Eval evaluate this runtime component.
*/
func (rt *letRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		for _, v := range rt.declared {
			if len(v.node.Children) == 0 {
				vs.SetLocalValue(v.node.Token.Val, nil)
			} else {
				err = rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
					"Let can only declare simple variables", rt.node)
				break
			}
		}

		if err == nil {
			res, err = rt.node.Children[0].Runtime.Eval(vs, is, tid)
		}
	}

	return res, err
}
