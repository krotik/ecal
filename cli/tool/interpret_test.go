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
	"reflect"
	"strconv"
	"strings"
	"testing"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/fileutil"
	"devt.de/krotik/ecal/config"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/stdlib"
	"devt.de/krotik/ecal/util"
)

const testDir = "tooltest"

var testLogOut *bytes.Buffer
var testTerm *testConsoleLineTerminal

func newTestInterpreter() *CLIInterpreter {
	tin := NewCLIInterpreter()

	// Redirect I/O bits into internal buffers

	testTerm = &testConsoleLineTerminal{nil, bytes.Buffer{}}
	tin.Term = testTerm

	testLogOut = &bytes.Buffer{}
	tin.LogOut = testLogOut

	return tin
}

func newTestInterpreterWithConfig() *CLIInterpreter {
	tin := newTestInterpreter()

	if res, _ := fileutil.PathExists(testDir); res {
		os.RemoveAll(testDir)
	}

	err := os.Mkdir(testDir, 0770)
	if err != nil {
		fmt.Print("Could not create test directory:", err.Error())
		os.Exit(1)
	}

	l := testDir
	tin.Dir = &l

	tin.CustomWelcomeMessage = "123"

	return tin
}

func tearDown() {
	err := os.RemoveAll(testDir)
	if err != nil {
		fmt.Print("Could not remove test directory:", err.Error())
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing
}

func TestInterpretBasicFunctions(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing

	// Test normal initialisation

	tin := NewCLIInterpreter()
	if err := tin.CreateTerm(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tin = newTestInterpreter()

	// Test help output

	osArgs = []string{"foo", "bar", "-help"}

	flag.CommandLine.SetOutput(&testTerm.out)

	if stop := tin.ParseArgs(); !stop {
		t.Error("Asking for help should request to stop the program")
		return
	}

	if !strings.Contains(testTerm.out.String(), "Root directory for ECAL interpreter") {
		t.Error("Helptext does not contain expected string - output:", testTerm.out.String())
		return
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing

	// Test interpret

	tin = newTestInterpreter()

	osArgs = []string{"foo", "bar", "-help"}

	flag.CommandLine.SetOutput(&testTerm.out)

	errorutil.AssertOk(tin.Interpret(true))

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing

	// Test entry file parsing

	tin = NewCLIInterpreter()

	osArgs = []string{"foo", "bar", "myfile"}

	if stop := tin.ParseArgs(); stop {
		t.Error("Giving an entry file should not stop the program")
		return
	}

	if stop := tin.ParseArgs(); stop {
		t.Error("Giving an entry file should not stop the program")
		return
	}

	if tin.EntryFile != "myfile" {
		t.Error("Unexpected entryfile:", tin.EntryFile)
		return
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing

	osArgs = []string{"foo", "bar"}

	// Try to load non-existing plugins (success case is tested in stdlib)

	tin = newTestInterpreterWithConfig()
	defer tearDown()

	l1 := ""
	tin.LogFile = &l1
	l2 := ""
	tin.LogLevel = &l2

	ioutil.WriteFile(filepath.Join(testDir, ".ecal.json"), []byte(`{
  "stdlibPlugins" : [{
    "package" : "mypkg",
    "name" : "myfunc",
    "path" : "./myfunc.so",
    "symbol" : "ECALmyfunc"
  }]
}`), 0666)

	err := tin.Interpret(true)

	if err == nil || err.Error() != "Could not load plugins defined in .ecal.json" {
		t.Error("Unexpected result:", err.Error())
		return
	}

	if !strings.Contains(testLogOut.String(), "Error loading plugins") {
		t.Error("Unexpected result:", testLogOut.String())
		return
	}
}

func TestCreateRuntimeProvider(t *testing.T) {
	tin := newTestInterpreterWithConfig()
	defer tearDown()

	l := filepath.Join(testDir, "test.log")
	tin.LogFile = &l

	if err := tin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, ok := tin.RuntimeProvider.Logger.(*util.BufferLogger); !ok {
		t.Errorf("Unexpected logger: %#v", tin.RuntimeProvider.Logger)
		return
	}

	tin = newTestInterpreterWithConfig()
	defer tearDown()

	l = "error"
	tin.LogLevel = &l

	if err := tin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, ok := tin.RuntimeProvider.Logger.(*util.LogLevelLogger); !ok {
		t.Errorf("Unexpected logger: %#v", tin.RuntimeProvider.Logger)
		return
	}

	if err := tin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, ok := tin.RuntimeProvider.Logger.(*util.LogLevelLogger); !ok {
		t.Errorf("Unexpected logger: %#v", tin.RuntimeProvider.Logger)
		return
	}
}

func TestLoadInitialFile(t *testing.T) {
	tin := NewCLIDebugInterpreter(newTestInterpreterWithConfig())
	defer tearDown()

	if err := tin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tin.RuntimeProvider.Debugger = interpreter.NewECALDebugger(tin.GlobalVS)
	tin.RuntimeProvider.Logger = util.NewMemoryLogger(10)
	tin.RuntimeProvider.ImportLocator = &util.MemoryImportLocator{}

	tin.EntryFile = filepath.Join(testDir, "foo.ecal")

	ioutil.WriteFile(tin.EntryFile, []byte("a := 1"), 0777)

	if err := tin.CLIInterpreter.LoadInitialFile(1); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if tin.GlobalVS.String() != `GlobalScope {
    a (float64) : 1
}` {
		t.Error("Unexpected scope:", tin.GlobalVS)
		return
	}
}

func TestInterpret(t *testing.T) {
	tin := newTestInterpreterWithConfig()
	defer tearDown()

	if err := tin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tin.RuntimeProvider.Logger, _ = util.NewLogLevelLogger(util.NewMemoryLogger(10), "info")
	tin.RuntimeProvider.ImportLocator = &util.MemoryImportLocator{
		Files: map[string]string{
			"foo": "a := 1",
		},
	}

	l1 := ""
	tin.LogFile = &l1
	l2 := ""
	tin.LogLevel = &l2

	testTerm.in = []string{"xxx := 1", "q"}

	if err := tin.Interpret(true); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if testLogOut.String() != `ECAL `+config.ProductVersion+`
Log level: info - Root directory: tooltest
123
Type 'q' or 'quit' to exit the shell and '?' to get help
` {
		t.Error("Unexpected result:", testLogOut.String())
		return
	}

	if tin.GlobalVS.String() != `GlobalScope {
    xxx (float64) : 1
}` {
		t.Error("Unexpected scope:", tin.GlobalVS)
		return
	}
}

func TestHandleInput(t *testing.T) {
	tin := newTestInterpreterWithConfig()
	defer tearDown()

	tin.CustomHandler = &testCustomHandler{}

	if err := tin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	stdlib.AddStdlibPkg("foo", "bar")
	stdlib.AddStdlibFunc("foo", "Println",
		stdlib.NewECALFunctionAdapter(reflect.ValueOf(fmt.Println), "xxx"))
	stdlib.AddStdlibFunc("foo", "Atoi",
		stdlib.NewECALFunctionAdapter(reflect.ValueOf(strconv.Atoi), "xxx"))

	tin.RuntimeProvider.Logger, _ = util.NewLogLevelLogger(util.NewMemoryLogger(10), "info")
	tin.RuntimeProvider.ImportLocator = &util.MemoryImportLocator{}
	tin.CustomHelpString = "123"

	l1 := ""
	tin.LogFile = &l1
	l2 := ""
	tin.LogLevel = &l2

	testTerm.in = []string{"?", "@format", "@reload", "@sym", "@std", "@cus", "q"}

	if err := tin.Interpret(true); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	// Just check for a simple string no need for the whole thing

	if !strings.Contains(testTerm.out.String(), "New creates a new object instance.") {
		t.Error("Unexpected result:", testTerm.out.String())
		return
	}

	testTerm.out.Reset()

	testTerm.in = []string{"@sym raise", "@std math.Phi", "@std foo Print", "q"}

	if err := tin.Interpret(true); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if strings.HasSuffix(testTerm.out.String(), `╒═════════════════╤═══════════════════════════════╕
│Inbuild function │Description                    │
╞═════════════════╪═══════════════════════════════╡
│raise            │Raise returns an error object. │
│                 │                               │
╘═════════════════╧═══════════════════════════════╛
╒═════════╤══════════════════╕
│Constant │Value             │
╞═════════╪══════════════════╡
│math.Phi │1.618033988749895 │
│         │                  │
╘═════════╧══════════════════╛
╒════════════╤════════════╕
│Function    │Description │
╞════════════╪════════════╡
│foo.Println │xxx         │
│            │            │
╘════════════╧════════════╛

`) {
		t.Error("Unexpected result:", testTerm.out.String())
		return
	}

	testTerm.out.Reset()

	testTerm.in = []string{"1", "raise(123)", "q"}

	if err := tin.Interpret(true); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if testTerm.out.String() != `1
ECAL error in foo: 123 () (Line:1 Pos:1)
` {
		t.Error("Unexpected result:", testTerm.out.String())
		return
	}
}
