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
Package scope contains the block scope implementation for the event condition language ECAL.
*/
package scope

import (
	"fmt"

	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/ecal/parser"
)

/*
Default scope names
*/
const (
	GlobalScope = "GlobalScope"
	FuncPrefix  = "func:"
)

/*
NameFromASTNode returns a scope name from a given ASTNode.
*/
func NameFromASTNode(node *parser.ASTNode) string {
	return fmt.Sprintf("block: %v (Line:%d Pos:%d)", node.Name, node.Token.Lline, node.Token.Lpos)
}

/*
EvalToString should be used if a value should be converted into a string.
*/
func EvalToString(v interface{}) string {
	return stringutil.ConvertToString(v)
}

/*
ToObject converts a Scope into an object.
*/
func ToObject(vs parser.Scope) map[interface{}]interface{} {
	res := make(map[interface{}]interface{})
	for k, v := range vs.(*varsScope).storage {
		res[k] = v
	}
	return res
}

/*
ToScope converts a given object into a Scope.
*/
func ToScope(name string, o map[interface{}]interface{}) parser.Scope {
	vs := NewScope(name)
	for k, v := range o {
		vs.SetValue(fmt.Sprint(k), v)
	}
	return vs
}

/*
ConvertJSONToECALObject converts a JSON container structure into an object which
can be used by ECAL.
*/
func ConvertJSONToECALObject(v interface{}) interface{} {
	res := v

	if mapContainer, ok := v.(map[interface{}]interface{}); ok {
		newRes := make(map[interface{}]interface{})

		for mk, mv := range mapContainer {
			newRes[mk] = ConvertJSONToECALObject(mv)
		}

		res = newRes

	} else if mapContainer, ok := v.(map[string]interface{}); ok {
		newRes := make(map[interface{}]interface{})

		for mk, mv := range mapContainer {
			newRes[mk] = ConvertJSONToECALObject(mv)
		}

		res = newRes

	} else if mapList, ok := v.([]interface{}); ok {
		newRes := make([]interface{}, len(mapList))

		for i, lv := range mapList {
			newRes[i] = ConvertJSONToECALObject(lv)
		}

		res = newRes
	} else if mapList, ok := v.([]map[string]interface{}); ok {
		newRes := make([]interface{}, len(mapList))

		for i, lv := range mapList {
			newRes[i] = ConvertJSONToECALObject(lv)
		}

		res = newRes
	}

	return res
}

/*
ConvertECALToJSONObject converts an ECAL container structure into an object which
can be marshalled into a JSON string.
*/
func ConvertECALToJSONObject(v interface{}) interface{} {
	res := v

	if mapContainer, ok := v.(map[interface{}]interface{}); ok {
		newRes := make(map[string]interface{})

		for mk, mv := range mapContainer {
			newRes[fmt.Sprint(mk)] = ConvertECALToJSONObject(mv)
		}

		res = newRes

	} else if mapList, ok := v.([]interface{}); ok {
		newRes := make([]interface{}, len(mapList))

		for i, lv := range mapList {
			newRes[i] = ConvertECALToJSONObject(lv)
		}

		res = newRes
	}

	return res
}
