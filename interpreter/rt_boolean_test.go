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
	"testing"
)

func TestSimpleBoolean(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`2 >= 2`, nil,
		`
>=
  number: 2
  number: 2
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"foo" >= "bar"`, nil,
		`
>=
  string: 'foo'
  string: 'bar'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`3 > 2`, nil,
		`
>
  number: 3
  number: 2
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"foo" > "bar"`, nil,
		`
>
  string: 'foo'
  string: 'bar'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`2 <= 2`, nil,
		`
<=
  number: 2
  number: 2
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"bar" <= "foo"`, nil,
		`
<=
  string: 'bar'
  string: 'foo'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`2 < 3`, nil,
		`
<
  number: 2
  number: 3
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"bar" < "foo"`, nil,
		`
<
  string: 'bar'
  string: 'foo'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`3 == 3`, nil,
		`
==
  number: 3
  number: 3
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`3 == 3 == true`, nil,
		`
==
  ==
    number: 3
    number: 3
  true
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`2 != 3`, nil,
		`
!=
  number: 2
  number: 3
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`3 != 3 == false`, nil,
		`
==
  !=
    number: 3
    number: 3
  false
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`null == null`, nil,
		`
==
  null
  null
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`1 < 2 and 2 > 1`, nil,
		`
and
  <
    number: 1
    number: 2
  >
    number: 2
    number: 1
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`1 < 2 or 2 < 1`, nil,
		`
or
  <
    number: 1
    number: 2
  <
    number: 2
    number: 1
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`not (1 < 2 or 2 < 1)`, nil,
		`
not
  or
    <
      number: 1
      number: 2
    <
      number: 2
      number: 1
`[1:])

	if fmt.Sprint(res) != "false" || err != nil {
		t.Error(res, err)
		return
	}
}

func TestConditionOperators(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`"Hans" like "Ha*"`, nil,
		`
like
  string: 'Hans'
  string: 'Ha*'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"Hans" hasprefix "Ha"`, nil,
		`
hasprefix
  string: 'Hans'
  string: 'Ha'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"Hans" hassuffix "ns"`, nil,
		`
hassuffix
  string: 'Hans'
  string: 'ns'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`2 in [1,2,3]`, nil,
		`
in
  number: 2
  list
    number: 1
    number: 2
    number: 3
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"Hans" in [1,2,"Hans"]`, nil,
		`
in
  string: 'Hans'
  list
    number: 1
    number: 2
    string: 'Hans'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"NotHans" notin [1,2,"Hans"]`, nil,
		`
notin
  string: 'NotHans'
  list
    number: 1
    number: 2
    string: 'Hans'
`[1:])

	if fmt.Sprint(res) != "true" || err != nil {
		t.Error(res, err)
		return
	}
}
