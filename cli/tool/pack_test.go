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
	"devt.de/krotik/common/stringutil"
)

const packTestDir = "packtest"

var testPackOut *bytes.Buffer

var lastReturnCode = 0
var lastRuntimeError error

func setupPackTestDir() {
	if res, _ := fileutil.PathExists(packTestDir); res {
		os.RemoveAll(packTestDir)
	}

	err := os.Mkdir(packTestDir, 0770)
	if err != nil {
		fmt.Print("Could not create test directory:", err.Error())
		os.Exit(1)
	}

	err = os.Mkdir(filepath.Join(packTestDir, "sub"), 0770)
	if err != nil {
		fmt.Print("Could not create test directory:", err.Error())
		os.Exit(1)
	}

	osExit = func(code int) {
		lastReturnCode = code
	}

	handleError = func(err error) {
		lastRuntimeError = err
	}
}

func tearDownPackTestDir() {
	err := os.RemoveAll(packTestDir)
	if err != nil {
		fmt.Print("Could not remove test directory:", err.Error())
	}
}

func newTestCLIPacker() *CLIPacker {
	clip := NewCLIPacker()

	testPackOut = &bytes.Buffer{}
	clip.LogOut = testPackOut

	return clip
}

func TestPackParseArgs(t *testing.T) {
	setupPackTestDir()
	defer tearDownPackTestDir()

	clip := newTestCLIPacker()

	packTestSrcBin := filepath.Join(packTestDir, "source.bin")
	out := bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{packTestSrcBin, "foo", "-help"}

	if ok := clip.ParseArgs(); !ok {
		t.Error("Asking for help should ask to finish the program")
		return
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{packTestSrcBin, "foo", "-help"}

	if err := clip.Pack(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if !strings.Contains(out.String(), "Root directory for ECAL interpreter") {
		t.Error("Unexpected output:", out.String())
		return
	}

	out = bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{packTestSrcBin, "foo", "myentryfile"}

	if ok := clip.ParseArgs(); ok {
		t.Error("Only asking for help should finish the program")
		return
	}

	if ok := clip.ParseArgs(); ok {
		t.Error("Only asking for help should finish the program")
		return
	}

	if clip.EntryFile != "myentryfile" {
		t.Error("Unexpected output:", clip.EntryFile)
		return
	}
}

func TestPackPacking(t *testing.T) {
	setupPackTestDir()
	defer tearDownPackTestDir()

	clip := newTestCLIPacker()

	packTestSrcBin := filepath.Join(packTestDir, "source.bin")
	packTestEntry := filepath.Join(packTestDir, "myentry.ecal")
	packAnotherFile := filepath.Join(packTestDir, "sub", "anotherfile.ecal")
	packTestDestBin := filepath.Join(packTestDir, "dest.exe")

	b1 = 5
	b2 = len(packmarker) + 11

	err := ioutil.WriteFile(packTestSrcBin, []byte("mybinaryfilecontent#somemorecontent"+
		stringutil.GenerateRollingString("123", 30)), 0777)
	errorutil.AssertOk(err)

	err = ioutil.WriteFile(packTestEntry, []byte("myvar := 1; 5"), 0777)
	errorutil.AssertOk(err)

	err = ioutil.WriteFile(packAnotherFile, []byte("func f() { raise(123) };f()"), 0777)
	errorutil.AssertOk(err)

	out := bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	// Write a binary with return code

	osArgs = []string{packTestSrcBin, "foo", "-dir", packTestDir, "-target",
		packTestDestBin, packTestEntry}

	// Simulate that whitespaces are added around the pack marker

	oldpackmarker := packmarker
	packmarker = fmt.Sprintf("\n\n\n%v\n\n\n", packmarker)

	if err := clip.Pack(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	packmarker = oldpackmarker

	if !strings.Contains(testPackOut.String(), "bytes for intro") {
		t.Error("Unexpected output:", testPackOut.String())
		return
	}

	// Write a binary with which errors

	clip = newTestCLIPacker()

	out = bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{packTestSrcBin, "foo", "-dir", packTestDir, "-target",
		packTestDestBin + ".error", packAnotherFile}

	if err := clip.Pack(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if !strings.Contains(testPackOut.String(), "bytes for intro") {
		t.Error("Unexpected output:", testPackOut.String())
		return
	}

	// Write also a corrupted binary

	err = ioutil.WriteFile(packTestDestBin+".corrupted", []byte(
		"mybinaryfilecontent#somemorecontent"+
			stringutil.GenerateRollingString("123", 30)+
			"\n"+
			packmarker+
			"\n"+
			stringutil.GenerateRollingString("123", 30)), 0777)

	errorutil.AssertOk(err)

	testRunningPackedBinary(t)
}

func testRunningPackedBinary(t *testing.T) {
	packTestDestBin := filepath.Join(packTestDir, "dest") // Suffix .exe should be appended

	out := bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{packTestDestBin + ".exe.corrupted"}

	RunPackedBinary()

	if lastRuntimeError == nil || lastRuntimeError.Error() != "zip: not a valid zip file" {
		t.Error("Unexpected result:", lastRuntimeError)
		return
	}

	out = bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)

	osArgs = []string{packTestDestBin}

	RunPackedBinary()

	if lastRuntimeError != nil {
		t.Error("Unexpected result:", lastRuntimeError)
		return
	}

	if lastReturnCode != 5 {
		t.Error("Unexpected result:", lastReturnCode)
		return
	}

	out = bytes.Buffer{}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Reset CLI parsing
	flag.CommandLine.SetOutput(&out)
	osStderr = &out

	osArgs = []string{packTestDestBin + ".exe.error"}

	RunPackedBinary()

	if lastRuntimeError != nil {
		t.Error("Unexpected result:", lastRuntimeError)
		return
	}

	if !strings.HasPrefix(out.String(), "ECAL error in packtest/dest.exe.error") ||
		!strings.Contains(out.String(), "raise(123)") {
		t.Error("Unexpected result:", out.String())
		return
	}
}
