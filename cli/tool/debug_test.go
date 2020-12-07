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
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/stdlib"
	"devt.de/krotik/ecal/util"
)

var testDebugLogOut *bytes.Buffer

func newTestDebugWithConfig() *CLIDebugInterpreter {
	tdin := NewCLIDebugInterpreter(newTestInterpreterWithConfig())

	testDebugLogOut = &bytes.Buffer{}
	tdin.LogOut = testDebugLogOut

	return tdin
}

func TestDebugBasicFunctions(t *testing.T) {
	tdin := newTestDebugWithConfig()
	defer tearDown()

	// Test help output

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing

	osArgs = []string{"foo", "bar", "-help"}
	defer func() { osArgs = []string{} }()

	flag.CommandLine.SetOutput(&testTerm.out)

	if stop := tdin.ParseArgs(); !stop {
		t.Error("Asking for help should request to stop the program")
		return
	}

	if !strings.Contains(testTerm.out.String(), "Root directory for ECAL interpreter") {
		t.Error("Helptext does not contain expected string - output:", testTerm.out.String())
		return
	}

	if stop := tdin.ParseArgs(); stop {
		t.Error("Asking again should be caught by the short circuit")
		return
	}

	tdin = newTestDebugWithConfig()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset CLI parsing

	flag.CommandLine.SetOutput(&testTerm.out)

	if err := tdin.Interpret(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
}

func TestDebugInterpret(t *testing.T) {
	tdin := newTestDebugWithConfig()
	defer tearDown()

	if stop := tdin.ParseArgs(); stop {
		t.Error("Setting default args should be fine")
		return
	}

	if err := tdin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tdin.RuntimeProvider.Logger, _ = util.NewLogLevelLogger(util.NewMemoryLogger(10), "info")
	tdin.RuntimeProvider.ImportLocator = &util.MemoryImportLocator{
		Files: map[string]string{
			"foo": "a := 1",
		},
	}

	l1 := ""
	tdin.LogFile = &l1
	l2 := ""
	tdin.LogLevel = &l2
	l3 := true
	tdin.Interactive = &l3
	tdin.RunDebugServer = &l3

	testTerm.in = []string{"xxx := 1", "q"}

	// The interpret call takes quite long because the debug server is
	// closed by the defer call when the call returns

	if err := tdin.Interpret(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if !strings.Contains(testLogOut.String(),
		"Running in debug mode - with debug server on localhost:33274 - prefix debug commands with ##") {
		t.Error("Unexpected result:", testLogOut.String())
		return
	}

	if tdin.GlobalVS.String() != `GlobalScope {
    xxx (float64) : 1
}` {
		t.Error("Unexpected scope:", tdin.GlobalVS)
		return
	}
}

func TestDebugHandleInput(t *testing.T) {
	tdin := newTestDebugWithConfig()
	defer tearDown()

	if stop := tdin.ParseArgs(); stop {
		t.Error("Setting default args should be fine")
		return
	}

	if err := tdin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	stdlib.AddStdlibPkg("foo", "bar")
	stdlib.AddStdlibFunc("foo", "Println",
		stdlib.NewECALFunctionAdapter(reflect.ValueOf(fmt.Println), "xxx"))
	stdlib.AddStdlibFunc("foo", "Atoi",
		stdlib.NewECALFunctionAdapter(reflect.ValueOf(strconv.Atoi), "xxx"))

	tdin.RuntimeProvider.Logger, _ = util.NewLogLevelLogger(util.NewMemoryLogger(10), "info")
	tdin.RuntimeProvider.ImportLocator = &util.MemoryImportLocator{}
	tdin.CustomHelpString = "123"

	l1 := ""
	tdin.LogFile = &l1
	l2 := ""
	tdin.LogLevel = &l2
	l3 := true
	tdin.Interactive = &l3
	l4 := false
	tdin.RunDebugServer = &l4

	testTerm.in = []string{"?", "@dbg", "##status", "##foo", "q"}

	if err := tdin.Interpret(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	// Just check for a simple string no need for the whole thing
	time.Sleep(1 * time.Second)

	if !strings.Contains(testTerm.out.String(), "Set a breakpoint specifying <source>:<line>") {
		t.Error("Unexpected result:", testTerm.out.String())
		return
	}

	if !strings.Contains(testTerm.out.String(), "Unknown command: foo") {
		t.Error("Unexpected result:", testTerm.out.String())
		return
	}

	testTerm.out.Reset()

	testTerm.in = []string{"@dbg status", "q"}

	if err := tdin.Interpret(); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if testTerm.out.String() != `╒══════════════╤═════════════════════════════════════════╕
│Debug command │Description                              │
╞══════════════╪═════════════════════════════════════════╡
│status        │Shows breakpoints and suspended threads. │
│              │                                         │
╘══════════════╧═════════════════════════════════════════╛


` {
		t.Error("Unexpected result:", "#"+testTerm.out.String()+"#")
		return
	}

	testTerm.out.Reset()

	testTerm.in = []string{"1", "raise(123)", "q"}

	if err := tdin.Interpret(); err != nil {
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

func TestDebugTelnetServer(t *testing.T) {
	tdin := newTestDebugWithConfig()
	defer tearDown()

	if err := tdin.CreateRuntimeProvider("foo"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tdin.RuntimeProvider.Logger = util.NewMemoryLogger(10)
	tdin.RuntimeProvider.ImportLocator = &util.MemoryImportLocator{}
	tdin.RuntimeProvider.Debugger = interpreter.NewECALDebugger(tdin.GlobalVS)
	tdin.RuntimeProvider.Debugger.BreakOnError(false)
	tdin.CustomHandler = tdin

	addr := "localhost:33274"
	mlog := util.NewMemoryLogger(10)

	srv := &debugTelnetServer{
		address:     addr,
		logPrefix:   "testdebugserver",
		listener:    nil,
		listen:      true,
		echo:        true,
		interpreter: tdin,
		logger:      mlog,
	}
	defer func() {
		srv.listen = false
		srv.listener.Close() // Attempt to cleanup
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go srv.Run(wg)
	wg.Wait()

	conn, err := net.Dial("tcp", addr)
	errorutil.AssertOk(err)
	reader := bufio.NewReader(conn)

	fmt.Fprintf(conn, "a:= 1; a\n")

	line, err := reader.ReadString('}')
	errorutil.AssertOk(err)

	if line != `{
  "EncodedOutput": "MQo="
}` {
		t.Error("Unexpected output:", line)
		return
	}

	if tdin.GlobalVS.String() != `GlobalScope {
    a (float64) : 1
}` {
		t.Error("Unexpected result:", tdin.GlobalVS)
		return
	}

	fmt.Fprintf(conn, "##status\n")

	line, err = reader.ReadString('}')
	errorutil.AssertOk(err)
	l, err := reader.ReadString('}')
	errorutil.AssertOk(err)
	line += l
	l, err = reader.ReadString('}')
	errorutil.AssertOk(err)
	line += l
	line = strings.TrimSpace(line)

	if line != `{
  "breakonstart": false,
  "breakpoints": {},
  "sources": [
    "console input"
  ],
  "threads": {}
}` {
		t.Error("Unexpected output:", line)
		return
	}

	fmt.Fprintf(conn, "@sym\n")

	line, err = reader.ReadString('}')
	errorutil.AssertOk(err)

	if !strings.Contains(line, "KioqKioqKioqKioqKioqKioq") {
		t.Error("Unexpected output:", line)
		return
	}

	fmt.Fprintf(conn, "raise(123);1\n")

	line, err = reader.ReadString('}')
	errorutil.AssertOk(err)
	line = strings.TrimSpace(line)

	if line != `{
  "EncodedOutput": "RUNBTCBlcnJvciBpbiBmb286IDEyMyAoKSAoTGluZToxIFBvczoxKQo="
}` {
		t.Error("Unexpected output:", line)
		return
	}

	testDebugLogOut.Reset()

	errorutil.AssertOk(conn.Close())

	time.Sleep(10 * time.Millisecond)

	if !strings.Contains(testDebugLogOut.String(), "Disconnected") {
		t.Error("Unexpected output:", testDebugLogOut)
		return
	}

	testDebugLogOut.Reset()

	conn, err = net.Dial("tcp", addr)
	errorutil.AssertOk(err)

	if _, err := fmt.Fprintf(conn, "q\n"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	// Make sure we can't start a second server on the same port

	mlog2 := util.NewMemoryLogger(10)

	srv2 := &debugTelnetServer{
		address:     addr,
		logPrefix:   "testdebugserver",
		listener:    nil,
		listen:      true,
		echo:        true,
		interpreter: tdin,
		logger:      mlog2,
	}
	defer func() {
		srv2.listen = false
		srv2.listener.Close() // Attempt to cleanup
	}()

	mlog2.Reset()

	wg = &sync.WaitGroup{}
	wg.Add(1)
	go srv2.Run(wg)
	wg.Wait()

	if !strings.Contains(mlog2.String(), "address already in use") {
		t.Error("Unexpected output:", mlog2.String())
		return
	}

	mlog.Reset()

	srv.listener.Close()

	time.Sleep(5 * time.Millisecond)

	if !strings.Contains(mlog.String(), "use of closed network connection") {
		t.Error("Unexpected output:", mlog.String())
		return
	}
}
