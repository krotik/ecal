/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/common/termutil"
)

type testConsoleLineTerminal struct {
	in  []string
	out bytes.Buffer
}

func (t *testConsoleLineTerminal) StartTerm() error {
	return nil
}

func (t *testConsoleLineTerminal) AddKeyHandler(handler termutil.KeyHandler) {
}

func (t *testConsoleLineTerminal) NextLine() (string, error) {
	var err error
	var ret string

	if len(t.in) > 0 {
		ret = t.in[0]
		t.in = t.in[1:]
	} else {
		err = fmt.Errorf("Input is empty in testConsoleLineTerminal")
	}
	return ret, err
}

func (t *testConsoleLineTerminal) NextLinePrompt(prompt string, echo rune) (string, error) {
	return t.NextLine()
}

func (t *testConsoleLineTerminal) WriteString(s string) {
	t.out.WriteString(s)
}

func (t *testConsoleLineTerminal) Write(p []byte) (n int, err error) {
	return t.out.Write(p)
}

func (t *testConsoleLineTerminal) StopTerm() {
}

type testCustomHandler struct {
}

func (t *testCustomHandler) CanHandle(s string) bool {
	return s == "@cus"
}

func (t *testCustomHandler) Handle(ot OutputTerminal, input string) {}

func (t *testCustomHandler) LoadInitialFile(tid uint64) error {
	return nil
}

type testOutputTerminal struct {
	b bytes.Buffer
}

func (t *testOutputTerminal) WriteString(s string) {
	t.b.WriteString(s)
}

func TestMatchesFulltextSearch(t *testing.T) {
	ot := &testOutputTerminal{}

	ok := matchesFulltextSearch(ot, "abc", "s[")

	if !ok && strings.HasPrefix(ot.b.String(), "Invalid search expression") {
		t.Error("Unexpected result:", ot.b.String(), ok)
		return
	}

	ot.b = bytes.Buffer{}

	ok = matchesFulltextSearch(ot, "abc", "a*")

	if !ok || ot.b.String() != "" {
		t.Error("Unexpected result:", ot.b.String(), ok)
		return
	}

	ok = matchesFulltextSearch(ot, "abc", "ac*")

	if ok || ot.b.String() != "" {
		t.Error("Unexpected result:", ot.b.String(), ok)
		return
	}
}

func TestFillTableRow(t *testing.T) {

	res := fillTableRow([]string{}, "test", stringutil.GenerateRollingString("123 ", 100))

	b, _ := json.Marshal(&res)

	if string(b) != `["test","123 123 123 123 123 123 123 123 123 123 123 123 `+
		`123 123 123 123 123 123 123 123","","123 123 123 123 123","",""]` {
		t.Error("Unexpected result:", string(b))
		return
	}
}
