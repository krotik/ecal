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
	"math"

	"devt.de/krotik/ecal/parser"
)

// Basic Arithmetic Operator Runtimes
// ==================================

type plusOpRuntime struct {
	*operatorRuntime
}

/*
plusOpRuntimeInst returns a new runtime component instance.
*/
func plusOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &plusOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *plusOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		// Use as prefix

		if len(rt.node.Children) == 1 {
			return rt.numVal(func(n float64) interface{} {
				return n
			}, vs, is, tid)
		}

		// Use as operation

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 + n2
		}, vs, is, tid)
	}

	return res, err
}

type minusOpRuntime struct {
	*operatorRuntime
}

/*
minusOpRuntimeInst returns a new runtime component instance.
*/
func minusOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &minusOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *minusOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		// Use as prefix

		if len(rt.node.Children) == 1 {
			return rt.numVal(func(n float64) interface{} {
				return -n
			}, vs, is, tid)
		}

		// Use as operation

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 - n2
		}, vs, is, tid)
	}

	return res, err
}

type timesOpRuntime struct {
	*operatorRuntime
}

/*
timesOpRuntimeInst returns a new runtime component instance.
*/
func timesOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &timesOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *timesOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 * n2
		}, vs, is, tid)
	}

	return res, err
}

type divOpRuntime struct {
	*operatorRuntime
}

/*
divOpRuntimeInst returns a new runtime component instance.
*/
func divOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &divOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *divOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 / n2
		}, vs, is, tid)
	}

	return res, err
}

type divintOpRuntime struct {
	*operatorRuntime
}

/*
divintOpRuntimeInst returns a new runtime component instance.
*/
func divintOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &divintOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *divintOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return math.Floor(n1 / n2)
		}, vs, is, tid)
	}

	return res, err
}

type modintOpRuntime struct {
	*operatorRuntime
}

/*
divOpRuntimeInst returns a new runtime component instance.
*/
func modintOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &modintOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *modintOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return float64(int64(n1) % int64(n2))
		}, vs, is, tid)
	}

	return res, err
}
