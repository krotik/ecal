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
	"math"
	"reflect"
	"strconv"
	"testing"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/scope"
)

func TestECALFunctionAdapterSimple(t *testing.T) {

	res, err := runAdapterTest(
		reflect.ValueOf(strconv.Atoi),
		[]interface{}{"1"},
	)

	if errorutil.AssertOk(err); res != float64(1) {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(strconv.ParseUint),
		[]interface{}{"123", float64(0), float64(0)},
	)

	if errorutil.AssertOk(err); res != float64(123) {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(strconv.ParseFloat),
		[]interface{}{"123.123", float64(0)},
	)

	if errorutil.AssertOk(err); res != float64(123.123) {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(fmt.Sprintf),
		[]interface{}{"foo %v", "bar"},
	)

	if errorutil.AssertOk(err); res != "foo bar" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(math.Float32bits),
		[]interface{}{float64(11)},
	)
	errorutil.AssertOk(err)

	if r := fmt.Sprintf("%X", uint32(res.(float64))); r != fmt.Sprintf("%X", math.Float32bits(11)) {
		t.Error("Unexpected result: ", r, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(math.Float32frombits),
		[]interface{}{float64(math.Float32bits(11))},
	)
	errorutil.AssertOk(err)

	if r := fmt.Sprintf("%v", res.(float64)); r != "11" {
		t.Error("Unexpected result: ", r, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(math.Float32frombits),
		[]interface{}{math.Float32bits(11)}, // Giving the correct type also works
	)
	errorutil.AssertOk(err)

	if r := fmt.Sprintf("%v", res.(float64)); r != "11" {
		t.Error("Unexpected result: ", r, err)
		return
	}
}

func TestECALFunctionAdapterSimple2(t *testing.T) {

	res, err := runAdapterTest(
		reflect.ValueOf(math.Float64bits),
		[]interface{}{float64(11)},
	)
	errorutil.AssertOk(err)

	if r := fmt.Sprintf("%X", uint64(res.(float64))); r != fmt.Sprintf("%X", math.Float64bits(11)) {
		t.Error("Unexpected result: ", r, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(math.Float64frombits),
		[]interface{}{float64(math.Float64bits(11))},
	)
	errorutil.AssertOk(err)

	if r := fmt.Sprintf("%v", res.(float64)); r != "11" {
		t.Error("Unexpected result: ", r, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyUint),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyUint8),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyUint16),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyUintptr),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyInt8),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyInt16),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyInt32),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(dummyInt64),
		[]interface{}{float64(1)},
	)

	if errorutil.AssertOk(err); res != "1" {
		t.Error("Unexpected result: ", res, err)
		return
	}
}

func TestECALFunctionAdapterErrors(t *testing.T) {

	// Test Error cases

	res, err := runAdapterTest(
		reflect.ValueOf(strconv.ParseFloat),
		[]interface{}{"123.123", 0, 0},
	)

	if err == nil || err.Error() != "Too many parameters - got 3 expected 2" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(strconv.ParseFloat),
		[]interface{}{"Hans", 0},
	)

	if err == nil || err.Error() != `strconv.ParseFloat: parsing "Hans": invalid syntax` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = runAdapterTest(
		reflect.ValueOf(strconv.ParseFloat),
		[]interface{}{123, 0},
	)

	if err == nil || err.Error() != `Parameter 1 should be of type string but is of type int` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	// Make sure we are never panicing but just returning an error

	res, err = runAdapterTest(
		reflect.ValueOf(errorutil.AssertTrue),
		[]interface{}{false, "Some Panic Description"},
	)

	if err == nil || err.Error() != `Error: Some Panic Description` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	// Get documentation

	afuncEcal := NewECALFunctionAdapter(reflect.ValueOf(fmt.Sprint), "test123")

	if s, err := afuncEcal.DocString(); s == "" || err != nil {
		t.Error("Docstring should return something")
		return
	}
}

func runAdapterTest(afunc reflect.Value, args []interface{}) (interface{}, error) {
	afuncEcal := &ECALFunctionAdapter{afunc, ""}
	return afuncEcal.Run("test", scope.NewScope(""), make(map[string]interface{}), 0, args)

}

func dummyUint(v uint) string {
	return fmt.Sprint(v)
}

func dummyUint8(v uint8) string {
	return fmt.Sprint(v)
}

func dummyUint16(v uint16) string {
	return fmt.Sprint(v)
}

func dummyUintptr(v uintptr) string {
	return fmt.Sprint(v)
}

func dummyInt8(v int8) string {
	return fmt.Sprint(v)
}

func dummyInt16(v int16) string {
	return fmt.Sprint(v)
}

func dummyInt32(v int32) string {
	return fmt.Sprint(v)
}

func dummyInt64(v int64) string {
	return fmt.Sprint(v)
}
