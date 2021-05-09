/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package util

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
)

func TestRuntimeError(t *testing.T) {

	ast, _ := parser.Parse("foo", "a")

	err1 := NewRuntimeError("foo", fmt.Errorf("foo"), "bar", ast)

	if err1.Error() != "ECAL error in foo: foo (bar) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err1)
		return
	}

	ast.Token = nil

	err2 := NewRuntimeError("foo", fmt.Errorf("foo"), "bar", ast)

	if err2.Error() != "ECAL error in foo: foo (bar)" {
		t.Error("Unexpected result:", err2)
		return
	}

	ast, _ = parser.Parse("foo", "a:=1")
	err3 := NewRuntimeError("foo", fmt.Errorf("foo"), "bar", ast)

	ast, _ = parser.Parse("bar1", "print(b)")
	err3.(TraceableRuntimeError).AddTrace(ast)
	ast, _ = parser.Parse("bar2", "raise(c)")
	err3.(TraceableRuntimeError).AddTrace(ast)
	ast, _ = parser.Parse("bar3", "1 + d")
	err3.(TraceableRuntimeError).AddTrace(ast)

	trace := strings.Join(err3.(TraceableRuntimeError).GetTraceString(), "\n")

	if trace != `print(b) (bar1:1)
raise(c) (bar2:1)
1 + d (bar3:1)` {
		t.Error("Unexpected result:", trace)
		return
	}

	err4 := &RuntimeErrorWithDetail{err3.(*RuntimeError), nil, nil}

	res, _ := json.MarshalIndent(err4.RuntimeError, "", "  ")
	if string(res) != `{
  "Detail": "bar",
  "Node": {
    "Name": ":=",
    "Token": {
      "ID": 39,
      "Pos": 1,
      "Val": ":=",
      "Identifier": false,
      "AllowEscapes": false,
      "PrefixNewlines": 0,
      "Lsource": "foo",
      "Lline": 1,
      "Lpos": 2
    },
    "Meta": null,
    "Children": [
      {
        "Name": "identifier",
        "Token": {
          "ID": 7,
          "Pos": 0,
          "Val": "a",
          "Identifier": true,
          "AllowEscapes": false,
          "PrefixNewlines": 0,
          "Lsource": "foo",
          "Lline": 1,
          "Lpos": 1
        },
        "Meta": null,
        "Children": [],
        "Runtime": null
      },
      {
        "Name": "number",
        "Token": {
          "ID": 6,
          "Pos": 3,
          "Val": "1",
          "Identifier": false,
          "AllowEscapes": false,
          "PrefixNewlines": 0,
          "Lsource": "foo",
          "Lline": 1,
          "Lpos": 4
        },
        "Meta": null,
        "Children": [],
        "Runtime": null
      }
    ],
    "Runtime": null
  },
  "Source": "foo",
  "Trace": [
    {
      "Name": "identifier",
      "Token": {
        "ID": 7,
        "Pos": 0,
        "Val": "print",
        "Identifier": true,
        "AllowEscapes": false,
        "PrefixNewlines": 0,
        "Lsource": "bar1",
        "Lline": 1,
        "Lpos": 1
      },
      "Meta": null,
      "Children": [
        {
          "Name": "funccall",
          "Token": null,
          "Meta": null,
          "Children": [
            {
              "Name": "identifier",
              "Token": {
                "ID": 7,
                "Pos": 6,
                "Val": "b",
                "Identifier": true,
                "AllowEscapes": false,
                "PrefixNewlines": 0,
                "Lsource": "bar1",
                "Lline": 1,
                "Lpos": 7
              },
              "Meta": null,
              "Children": [],
              "Runtime": null
            }
          ],
          "Runtime": null
        }
      ],
      "Runtime": null
    },
    {
      "Name": "identifier",
      "Token": {
        "ID": 7,
        "Pos": 0,
        "Val": "raise",
        "Identifier": true,
        "AllowEscapes": false,
        "PrefixNewlines": 0,
        "Lsource": "bar2",
        "Lline": 1,
        "Lpos": 1
      },
      "Meta": null,
      "Children": [
        {
          "Name": "funccall",
          "Token": null,
          "Meta": null,
          "Children": [
            {
              "Name": "identifier",
              "Token": {
                "ID": 7,
                "Pos": 6,
                "Val": "c",
                "Identifier": true,
                "AllowEscapes": false,
                "PrefixNewlines": 0,
                "Lsource": "bar2",
                "Lline": 1,
                "Lpos": 7
              },
              "Meta": null,
              "Children": [],
              "Runtime": null
            }
          ],
          "Runtime": null
        }
      ],
      "Runtime": null
    },
    {
      "Name": "plus",
      "Token": {
        "ID": 33,
        "Pos": 2,
        "Val": "+",
        "Identifier": false,
        "AllowEscapes": false,
        "PrefixNewlines": 0,
        "Lsource": "bar3",
        "Lline": 1,
        "Lpos": 3
      },
      "Meta": null,
      "Children": [
        {
          "Name": "number",
          "Token": {
            "ID": 6,
            "Pos": 0,
            "Val": "1",
            "Identifier": false,
            "AllowEscapes": false,
            "PrefixNewlines": 0,
            "Lsource": "bar3",
            "Lline": 1,
            "Lpos": 1
          },
          "Meta": null,
          "Children": [],
          "Runtime": null
        },
        {
          "Name": "identifier",
          "Token": {
            "ID": 7,
            "Pos": 4,
            "Val": "d",
            "Identifier": true,
            "AllowEscapes": false,
            "PrefixNewlines": 0,
            "Lsource": "bar3",
            "Lline": 1,
            "Lpos": 5
          },
          "Meta": null,
          "Children": [],
          "Runtime": null
        }
      ],
      "Runtime": null
    }
  ],
  "Type": "foo"
}` {
		t.Error("Unexpected result:", string(res))
		return
	}

	s := scope.NewScope("aa")
	s.SetValue("xx", 123)
	err4 = &RuntimeErrorWithDetail{err3.(*RuntimeError), s, sync.Mutex{}}

	res, _ = json.MarshalIndent(err4, "", "  ")
	if string(res) != `{
  "Data": {},
  "Detail": "bar",
  "Environment": {
    "xx": 123
  },
  "Node": {
    "Name": ":=",
    "Token": {
      "ID": 39,
      "Pos": 1,
      "Val": ":=",
      "Identifier": false,
      "AllowEscapes": false,
      "PrefixNewlines": 0,
      "Lsource": "foo",
      "Lline": 1,
      "Lpos": 2
    },
    "Meta": null,
    "Children": [
      {
        "Name": "identifier",
        "Token": {
          "ID": 7,
          "Pos": 0,
          "Val": "a",
          "Identifier": true,
          "AllowEscapes": false,
          "PrefixNewlines": 0,
          "Lsource": "foo",
          "Lline": 1,
          "Lpos": 1
        },
        "Meta": null,
        "Children": [],
        "Runtime": null
      },
      {
        "Name": "number",
        "Token": {
          "ID": 6,
          "Pos": 3,
          "Val": "1",
          "Identifier": false,
          "AllowEscapes": false,
          "PrefixNewlines": 0,
          "Lsource": "foo",
          "Lline": 1,
          "Lpos": 4
        },
        "Meta": null,
        "Children": [],
        "Runtime": null
      }
    ],
    "Runtime": null
  },
  "Source": "foo",
  "Trace": [
    {
      "Name": "identifier",
      "Token": {
        "ID": 7,
        "Pos": 0,
        "Val": "print",
        "Identifier": true,
        "AllowEscapes": false,
        "PrefixNewlines": 0,
        "Lsource": "bar1",
        "Lline": 1,
        "Lpos": 1
      },
      "Meta": null,
      "Children": [
        {
          "Name": "funccall",
          "Token": null,
          "Meta": null,
          "Children": [
            {
              "Name": "identifier",
              "Token": {
                "ID": 7,
                "Pos": 6,
                "Val": "b",
                "Identifier": true,
                "AllowEscapes": false,
                "PrefixNewlines": 0,
                "Lsource": "bar1",
                "Lline": 1,
                "Lpos": 7
              },
              "Meta": null,
              "Children": [],
              "Runtime": null
            }
          ],
          "Runtime": null
        }
      ],
      "Runtime": null
    },
    {
      "Name": "identifier",
      "Token": {
        "ID": 7,
        "Pos": 0,
        "Val": "raise",
        "Identifier": true,
        "AllowEscapes": false,
        "PrefixNewlines": 0,
        "Lsource": "bar2",
        "Lline": 1,
        "Lpos": 1
      },
      "Meta": null,
      "Children": [
        {
          "Name": "funccall",
          "Token": null,
          "Meta": null,
          "Children": [
            {
              "Name": "identifier",
              "Token": {
                "ID": 7,
                "Pos": 6,
                "Val": "c",
                "Identifier": true,
                "AllowEscapes": false,
                "PrefixNewlines": 0,
                "Lsource": "bar2",
                "Lline": 1,
                "Lpos": 7
              },
              "Meta": null,
              "Children": [],
              "Runtime": null
            }
          ],
          "Runtime": null
        }
      ],
      "Runtime": null
    },
    {
      "Name": "plus",
      "Token": {
        "ID": 33,
        "Pos": 2,
        "Val": "+",
        "Identifier": false,
        "AllowEscapes": false,
        "PrefixNewlines": 0,
        "Lsource": "bar3",
        "Lline": 1,
        "Lpos": 3
      },
      "Meta": null,
      "Children": [
        {
          "Name": "number",
          "Token": {
            "ID": 6,
            "Pos": 0,
            "Val": "1",
            "Identifier": false,
            "AllowEscapes": false,
            "PrefixNewlines": 0,
            "Lsource": "bar3",
            "Lline": 1,
            "Lpos": 1
          },
          "Meta": null,
          "Children": [],
          "Runtime": null
        },
        {
          "Name": "identifier",
          "Token": {
            "ID": 7,
            "Pos": 4,
            "Val": "d",
            "Identifier": true,
            "AllowEscapes": false,
            "PrefixNewlines": 0,
            "Lsource": "bar3",
            "Lline": 1,
            "Lpos": 5
          },
          "Meta": null,
          "Children": [],
          "Runtime": null
        }
      ],
      "Runtime": null
    }
  ],
  "Type": "foo"
}` {
		t.Error("Unexpected result:", string(res))
		return
	}
}
