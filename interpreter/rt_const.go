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

import "devt.de/krotik/ecal/parser"

/*
trueRuntime is the runtime component for the true constant.
*/
type trueRuntime struct {
	*baseRuntime
}

/*
trueRuntimeInst returns a new runtime component instance.
*/
func trueRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &trueRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *trueRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)
	return true, err
}

/*
falseRuntime is the runtime component for the false constant.
*/
type falseRuntime struct {
	*baseRuntime
}

/*
falseRuntimeInst returns a new runtime component instance.
*/
func falseRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &falseRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *falseRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)
	return false, err
}

/*
nullRuntime is the runtime component for the null constant.
*/
type nullRuntime struct {
	*baseRuntime
}

/*
nullRuntimeInst returns a new runtime component instance.
*/
func nullRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &nullRuntime{newBaseRuntime(erp, node)}
}

/*
Eval evaluate this runtime component.
*/
func (rt *nullRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	_, err := rt.baseRuntime.Eval(vs, is, tid)
	return nil, err
}
