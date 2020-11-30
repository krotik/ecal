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
	"regexp"
	"strings"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/parser"
)

// Basic Boolean Operator Runtimes
// ===============================

type greaterequalOpRuntime struct {
	*operatorRuntime
}

/*
greaterequalOpRuntimeInst returns a new runtime component instance.
*/
func greaterequalOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &greaterequalOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *greaterequalOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 >= n2
		}, vs, is, tid)

		if err != nil {
			res, err = rt.strOp(func(n1 string, n2 string) interface{} {
				return n1 >= n2
			}, vs, is, tid)
		}
	}

	return res, err
}

type greaterOpRuntime struct {
	*operatorRuntime
}

/*
greaterOpRuntimeInst returns a new runtime component instance.
*/
func greaterOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &greaterOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *greaterOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 > n2
		}, vs, is, tid)

		if err != nil {
			res, err = rt.strOp(func(n1 string, n2 string) interface{} {
				return n1 > n2
			}, vs, is, tid)
		}
	}

	return res, err
}

type lessequalOpRuntime struct {
	*operatorRuntime
}

/*
lessequalOpRuntimeInst returns a new runtime component instance.
*/
func lessequalOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &lessequalOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *lessequalOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 <= n2
		}, vs, is, tid)

		if err != nil {
			res, err = rt.strOp(func(n1 string, n2 string) interface{} {
				return n1 <= n2
			}, vs, is, tid)
		}
	}

	return res, err
}

type lessOpRuntime struct {
	*operatorRuntime
}

/*
lessOpRuntimeInst returns a new runtime component instance.
*/
func lessOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &lessOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *lessOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.numOp(func(n1 float64, n2 float64) interface{} {
			return n1 < n2
		}, vs, is, tid)

		if err != nil {
			res, err = rt.strOp(func(n1 string, n2 string) interface{} {
				return n1 < n2
			}, vs, is, tid)
		}
	}

	return res, err
}

type equalOpRuntime struct {
	*operatorRuntime
}

/*
equalOpRuntimeInst returns a new runtime component instance.
*/
func equalOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &equalOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *equalOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.genOp(func(n1 interface{}, n2 interface{}) interface{} {
			return n1 == n2
		}, vs, is, tid)
	}

	return res, err
}

type notequalOpRuntime struct {
	*operatorRuntime
}

/*
notequalOpRuntimeInst returns a new runtime component instance.
*/
func notequalOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &notequalOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *notequalOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.genOp(func(n1 interface{}, n2 interface{}) interface{} {
			return n1 != n2
		}, vs, is, tid)
	}

	return res, err
}

type andOpRuntime struct {
	*operatorRuntime
}

/*
andOpRuntimeInst returns a new runtime component instance.
*/
func andOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &andOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *andOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.boolOp(func(b1 bool, b2 bool) interface{} {
			return b1 && b2
		}, vs, is, tid)
	}

	return res, err
}

type orOpRuntime struct {
	*operatorRuntime
}

/*
orOpRuntimeInst returns a new runtime component instance.
*/
func orOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &orOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *orOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.boolOp(func(b1 bool, b2 bool) interface{} {
			return b1 || b2
		}, vs, is, tid)
	}

	return res, err
}

type notOpRuntime struct {
	*operatorRuntime
}

/*
notOpRuntimeInst returns a new runtime component instance.
*/
func notOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &notOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *notOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		res, err = rt.boolVal(func(b bool) interface{} {
			return !b
		}, vs, is, tid)
	}

	return res, err
}

// In-build condition operators
// ============================

/*
likeOpRuntime is the pattern matching operator. The syntax of the regular
expressions accepted is the same general syntax used by Go, Perl, Python, and
other languages.
*/
type likeOpRuntime struct {
	*operatorRuntime
}

/*
likeOpRuntimeInst returns a new runtime component instance.
*/
func likeOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &likeOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *likeOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		errorutil.AssertTrue(len(rt.node.Children) == 2,
			fmt.Sprint("Operation requires 2 operands", rt.node))

		str, err := rt.node.Children[0].Runtime.Eval(vs, is, tid)
		if err == nil {
			var pattern interface{}

			pattern, err = rt.node.Children[1].Runtime.Eval(vs, is, tid)
			if err == nil {
				var re *regexp.Regexp

				re, err = regexp.Compile(fmt.Sprint(pattern))
				if err == nil {

					res = re.MatchString(fmt.Sprint(str))
				}
			}
		}
	}

	return res, err
}

type beginswithOpRuntime struct {
	*operatorRuntime
}

/*
beginswithOpRuntimeInst returns a new runtime component instance.
*/
func beginswithOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &beginswithOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *beginswithOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		res, err = rt.strOp(func(s1 string, s2 string) interface{} {
			return strings.HasPrefix(s1, s2)
		}, vs, is, tid)
	}

	return res, err
}

type endswithOpRuntime struct {
	*operatorRuntime
}

/*
endswithOpRuntimeInst returns a new runtime component instance.
*/
func endswithOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &endswithOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *endswithOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		res, err = rt.strOp(func(s1 string, s2 string) interface{} {
			return strings.HasSuffix(s1, s2)
		}, vs, is, tid)
	}

	return res, err
}

type inOpRuntime struct {
	*operatorRuntime
}

/*
inOpRuntimeInst returns a new runtime component instance.
*/
func inOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &inOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *inOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		res, err = rt.listOp(func(val interface{}, list []interface{}) interface{} {
			for _, i := range list {
				if val == i {
					return true
				}
			}
			return false
		}, vs, is, tid)
	}

	return res, err
}

type notinOpRuntime struct {
	*inOpRuntime
}

/*
notinOpRuntimeInst returns a new runtime component instance.
*/
func notinOpRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &notinOpRuntime{&inOpRuntime{&operatorRuntime{newBaseRuntime(erp, node)}}}
}

/*
Eval evaluate this runtime component.
*/
func (rt *notinOpRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var res interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		if res, err = rt.inOpRuntime.Eval(vs, is, tid); err == nil {
			res = !res.(bool)
		}
	}

	return res, err
}
