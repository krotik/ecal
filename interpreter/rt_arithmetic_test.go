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
	"testing"
)

func TestSimpleArithmetics(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`1 + 2`, nil,
		`
plus
  number: 1
  number: 2
`[1:])

	if err != nil || res != 3. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`1 + 2 + 3`, nil,
		`
plus
  plus
    number: 1
    number: 2
  number: 3
`[1:])

	if err != nil || res != 6. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`1 - 2 + 3`, nil,
		`
plus
  minus
    number: 1
    number: 2
  number: 3
`[1:])

	if err != nil || res != 2. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`1 - 2`, nil,
		`
minus
  number: 1
  number: 2
`[1:])

	if err != nil || res != -1. {
		t.Error("Unexpected result: ", res, err)
		return
	}
}

func TestSimpleArithmetics2(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`-5.2 - 2.2`, nil,
		`
minus
  minus
    number: 5.2
  number: 2.2
`[1:])

	if err != nil || res != -7.4 {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`+ 5.2 * 2`, nil,
		`
times
  plus
    number: 5.2
  number: 2
`[1:])

	if err != nil || res != 10.4 {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`5.2 / 2`, nil,
		`
div
  number: 5.2
  number: 2
`[1:])

	if err != nil || res != 2.6 {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`5.2 // 2`, nil,
		`
divint
  number: 5.2
  number: 2
`[1:])

	if err != nil || res != 2. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`5.2 % 2`, nil,
		`
modint
  number: 5.2
  number: 2
`[1:])

	if err != nil || res != 1. {
		t.Error("Unexpected result: ", res, err)
		return
	}
}
