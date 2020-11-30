/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package stdlib

import (
	"fmt"
	"reflect"

	"devt.de/krotik/ecal/parser"
)

/*
ECALFunctionAdapter models a bridge adapter between an ECAL function to a Go function.
*/
type ECALFunctionAdapter struct {
	funcval   reflect.Value
	docstring string
}

/*
NewECALFunctionAdapter creates a new ECALFunctionAdapter.
*/
func NewECALFunctionAdapter(funcval reflect.Value, docstring string) *ECALFunctionAdapter {
	return &ECALFunctionAdapter{funcval, docstring}
}

/*
Run executes this function.
*/
func (ea *ECALFunctionAdapter) Run(instanceID string, vs parser.Scope,
	is map[string]interface{}, tid uint64, args []interface{}) (ret interface{}, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: %v", r)
		}
	}()

	funcType := ea.funcval.Type()

	// Build arguments

	fargs := make([]reflect.Value, 0, len(args))
	for i, arg := range args {

		if i == funcType.NumIn() {
			return nil, fmt.Errorf("Too many parameters - got %v expected %v",
				len(args), funcType.NumIn())
		}

		expectedType := funcType.In(i)

		// Try to convert into correct number types

		if float64Arg, ok := arg.(float64); ok {
			switch expectedType.Kind() {
			case reflect.Int:
				arg = int(float64Arg)
			case reflect.Int8:
				arg = int8(float64Arg)
			case reflect.Int16:
				arg = int16(float64Arg)
			case reflect.Int32:
				arg = int32(float64Arg)
			case reflect.Int64:
				arg = int64(float64Arg)
			case reflect.Uint:
				arg = uint(float64Arg)
			case reflect.Uint8:
				arg = uint8(float64Arg)
			case reflect.Uint16:
				arg = uint16(float64Arg)
			case reflect.Uint32:
				arg = uint32(float64Arg)
			case reflect.Uint64:
				arg = uint64(float64Arg)
			case reflect.Uintptr:
				arg = uintptr(float64Arg)
			case reflect.Float32:
				arg = float32(float64Arg)
			}
		}

		givenType := reflect.TypeOf(arg)

		// Check that the right types were given

		if givenType != expectedType &&
			!(expectedType.Kind() == reflect.Interface &&
				givenType.Kind() == reflect.Interface &&
				givenType.Implements(expectedType)) &&
			expectedType != reflect.TypeOf([]interface{}{}) {

			return nil, fmt.Errorf("Parameter %v should be of type %v but is of type %v",
				i+1, expectedType, givenType)
		}

		fargs = append(fargs, reflect.ValueOf(arg))
	}

	// Call the function

	vals := ea.funcval.Call(fargs)

	// Convert result value

	results := make([]interface{}, 0, len(vals))

	for i, v := range vals {
		res := v.Interface()

		if i == len(vals)-1 {

			// If the last item is an error then it is not part of the resutls
			// (it will be wrapped into a proper runtime error later)

			if funcType.Out(i) == reflect.TypeOf((*error)(nil)).Elem() {

				if res != nil {
					err = res.(error)
				}

				break
			}
		}

		// Convert result if it is a primitive type

		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			res = float64(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			res = float64(v.Uint())
		case reflect.Float32, reflect.Float64:
			res = v.Float()
		}

		results = append(results, res)
	}

	ret = results

	// Return a single value if results contains only a single item

	if len(results) == 1 {
		ret = results[0]
	}

	return ret, err
}

/*
DocString returns the docstring of the wrapped function.
*/
func (ea *ECALFunctionAdapter) DocString() (string, error) {
	return ea.docstring, nil
}
