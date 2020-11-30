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
	"strconv"
	"strings"

	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
)

/*
numberValueRuntime is the runtime component for constant numeric values.
*/
type numberValueRuntime struct {
	*baseRuntime
	numValue float64 // Numeric value
}

/*
numberValueRuntimeInst returns a new runtime component instance.
*/
func numberValueRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &numberValueRuntime{newBaseRuntime(erp, node), 0}
}

/*
Validate this node and all its child nodes.
*/
func (rt *numberValueRuntime) Validate() error {
	err := rt.baseRuntime.Validate()

	if err == nil {
		rt.numValue, err = strconv.ParseFloat(rt.node.Token.Val, 64)
	}

	return err
}

/*
Eval evaluate this runtime component.
*/
func (rt *numberValueRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	return rt.numValue, err
}

/*
stringValueRuntime is the runtime component for constant string values.
*/
type stringValueRuntime struct {
	*baseRuntime
}

/*
stringValueRuntimeInst returns a new runtime component instance.
*/
func stringValueRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &stringValueRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *stringValueRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	ret := rt.node.Token.Val

	if rt.node.Token.AllowEscapes {

		// Do allow string interpolation if escape sequences are allowed

		for {
			var replace string

			code, ok := rt.GetInfix(ret, "{{", "}}")
			if !ok {
				break
			}

			ast, ierr := parser.ParseWithRuntime(
				fmt.Sprintf("String interpolation: %v", code), code, rt.erp)

			if ierr == nil {

				if ierr = ast.Runtime.Validate(); ierr == nil {
					var res interface{}

					res, ierr = ast.Runtime.Eval(
						vs.NewChild(scope.NameFromASTNode(rt.node)),
						make(map[string]interface{}), tid)

					if ierr == nil {
						replace = fmt.Sprint(res)
					}
				}
			}

			if ierr != nil {
				replace = fmt.Sprintf("#%v", ierr.Error())
			}

			ret = strings.Replace(ret, fmt.Sprintf("{{%v}}", code), replace, 1)
		}

	}

	return ret, err
}

func (rt *stringValueRuntime) GetInfix(str string, start string, end string) (string, bool) {
	res := str

	if s := strings.Index(str, start); s >= 0 {
		s += len(start)
		if e := strings.Index(str, end); e >= 0 {
			res = str[s:e]
		}
	}

	return res, res != str
}

/*
mapValueRuntime is the runtime component for map values.
*/
type mapValueRuntime struct {
	*baseRuntime
}

/*
mapValueRuntimeInst returns a new runtime component instance.
*/
func mapValueRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &mapValueRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *mapValueRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	m := make(map[interface{}]interface{})

	if err == nil {
		for _, kvp := range rt.node.Children {
			var key, val interface{}

			if err == nil {
				if key, err = kvp.Children[0].Runtime.Eval(vs, is, tid); err == nil {
					if val, err = kvp.Children[1].Runtime.Eval(vs, is, tid); err == nil {
						m[key] = val
					}
				}
			}
		}
	}

	return m, err
}

/*
listValueRuntime is the runtime component for list values.
*/
type listValueRuntime struct {
	*baseRuntime
}

/*
listValueRuntimeInst returns a new runtime component instance.
*/
func listValueRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &listValueRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *listValueRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)

	var l []interface{}

	if err == nil {
		for _, item := range rt.node.Children {
			if err == nil {
				var val interface{}
				if val, err = item.Runtime.Eval(vs, is, tid); err == nil {
					l = append(l, val)
				}

			}
		}
	}

	return l, err
}
