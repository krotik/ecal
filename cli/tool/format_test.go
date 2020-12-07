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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/fileutil"
)

const formatTestDir = "formattest"

func setupFormatTestDir() {

	if res, _ := fileutil.PathExists(formatTestDir); res {
		os.RemoveAll(formatTestDir)
	}

	err := os.Mkdir(formatTestDir, 0770)
	if err != nil {
		fmt.Print("Could not create test directory:", err.Error())
		os.Exit(1)
	}
}

func tearDownFormatTestDir() {
	err := os.RemoveAll(formatTestDir)
	if err != nil {
		fmt.Print("Could not remove test directory:", err.Error())
	}
}

func TestFormat(t *testing.T) {
	setupFormatTestDir()
	defer tearDownFormatTestDir()

	out := bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{"foo", "bar", "-help"}

	if err := Format(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if !strings.Contains(out.String(), "Root directory for ECAL files") {
		t.Error("Unexpected output:", out.String())
		return
	}

	myfile := filepath.Join(formatTestDir, "myfile.ecal")
	myfile2 := filepath.Join(formatTestDir, "myfile.eca")
	myfile3 := filepath.Join(formatTestDir, "myinvalidfile.ecal")

	originalContent := "if a == 1 { b := 1 }"

	err := ioutil.WriteFile(myfile, []byte(originalContent), 0777)
	errorutil.AssertOk(err)

	err = ioutil.WriteFile(myfile2, []byte(originalContent), 0777)
	errorutil.AssertOk(err)

	err = ioutil.WriteFile(myfile3, []byte(originalContent[5:]), 0777)
	errorutil.AssertOk(err)

	out = bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{"foo", "bar", "-dir", formatTestDir}

	if err := Format(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if out.String() != `Formatting all .ecal files in formattest
Could not format formattest/myinvalidfile.ecal: Parse error in formattest/myinvalidfile.ecal: Term cannot start an expression (==) (Line:1 Pos:1)
` {
		t.Error("Unexpected output:", out.String())
		return
	}

	myfileContent, err := ioutil.ReadFile(myfile)
	errorutil.AssertOk(err)

	if string(myfileContent) != `if a == 1 {
    b := 1
}
` {
		t.Error("Unexpected result:", string(myfileContent))
		return
	}

	myfileContent, err = ioutil.ReadFile(myfile2)
	errorutil.AssertOk(err)

	if string(myfileContent) != originalContent {
		t.Error("Unexpected result:", string(myfileContent))
		return
	}
}
