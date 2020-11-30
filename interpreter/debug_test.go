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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

func TestSimpleDebugging(t *testing.T) {
	var err error

	defer func() {
		testDebugger = nil
	}()

	testDebugger = NewECALDebugger(nil)

	if _, err = testDebugger.HandleInput("break ECALEvalTest:3"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	if _, err = testDebugger.HandleInput("break ECALEvalTest:4"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	if _, err = testDebugger.HandleInput("disablebreak ECALEvalTest:4"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var tid uint64

	go func() {
		_, err = UnitTestEval(`
log("test1")
log("test2")
log("test3")
`, nil)
		if err != nil {
			t.Error(err)
		}

		testDebugger.RecordThreadFinished(tid)

		wg.Done()
	}()

	tid = waitForThreadSuspension(t)

	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ := json.MarshalIndent(out, "", "  ")
	outString := string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:3": true,
    "ECALEvalTest:4": false
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  }
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	out, err = testDebugger.HandleInput(fmt.Sprintf("describe %v", tid))

	outBytes, _ = json.MarshalIndent(out, "", "  ")
	outString = string(outBytes)

	if err != nil || outString != `{
  "callStack": [],
  "callStackNode": [],
  "callStackVsSnapshot": [],
  "callStackVsSnapshotGlobal": [],
  "code": "log(\"test2\")",
  "error": null,
  "node": {
    "allowescapes": false,
    "children": [
      {
        "children": [
          {
            "allowescapes": true,
            "id": 5,
            "identifier": false,
            "line": 3,
            "linepos": 5,
            "name": "string",
            "pos": 18,
            "source": "ECALEvalTest",
            "value": "test2"
          }
        ],
        "name": "funccall"
      }
    ],
    "id": 7,
    "identifier": true,
    "line": 3,
    "linepos": 1,
    "name": "identifier",
    "pos": 14,
    "source": "ECALEvalTest",
    "value": "log"
  },
  "threadRunning": false,
  "vs": {},
  "vsGlobal": {}
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	// Continue until the end

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v Resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg.Wait()

	if err != nil || testlogger.String() != `
test1
test2
test3`[1:] {
		t.Error("Unexpected result:", testlogger.String(), err)
		return
	}

	if _, err = testDebugger.HandleInput("rmbreak ECALEvalTest:4"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	out, err = testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ = json.MarshalIndent(out, "", "  ")
	outString = string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:3": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {}
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	if _, err = testDebugger.HandleInput("break ECALEvalTest:4"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("rmbreak ECALEvalTest"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	out, err = testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ = json.MarshalIndent(out, "", "  ")
	outString = string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {},
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {}
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}
}

func TestDebugReset(t *testing.T) {
	var err error

	defer func() {
		testDebugger = nil
	}()

	testDebugger = NewECALDebugger(nil)

	if _, err = testDebugger.HandleInput("break ECALEvalTest:3"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		_, err = UnitTestEval(`
log("test1")
log("test2")
log("test3")
`, nil)
		if err != nil {
			t.Error(err)
		}
	}()

	waitForThreadSuspension(t)

	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ := json.MarshalIndent(out, "", "  ")
	outString := string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:3": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  }
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	testDebugger.StopThreads(100 * time.Millisecond)

	wg.Wait()

	if err != nil || testlogger.String() != `
test1
test2`[1:] {
		t.Error("Unexpected result:", testlogger.String(), err)
		return
	}
}

func TestErrorStop(t *testing.T) {
	var err, evalError error

	defer func() {
		testDebugger = nil
	}()

	testDebugger = NewECALDebugger(nil)
	testDebugger.BreakOnError(true)

	if _, err = testDebugger.HandleInput("break ECALEvalTest:8"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		_, evalError = UnitTestEval(`
func err () {
	raise("foo")
}
log("test1")
log("test2")
err()
log("test3")
`, nil)
	}()

	waitForThreadSuspension(t)

	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ := json.MarshalIndent(out, "", "  ")
	outString := string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:8": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [
        "err() (ECALEvalTest:7)"
      ],
      "error": {
        "Data": null,
        "Detail": "",
        "Environment": {},
        "Node": {
          "Name": "identifier",
          "Token": {
            "ID": 7,
            "Pos": 16,
            "Val": "raise",
            "Identifier": true,
            "AllowEscapes": false,
            "Lsource": "ECALEvalTest",
            "Lline": 3,
            "Lpos": 2
          },
          "Meta": null,
          "Children": [
            {
              "Name": "funccall",
              "Token": null,
              "Meta": null,
              "Children": [
                {
                  "Name": "string",
                  "Token": {
                    "ID": 5,
                    "Pos": 22,
                    "Val": "foo",
                    "Identifier": false,
                    "AllowEscapes": true,
                    "Lsource": "ECALEvalTest",
                    "Lline": 3,
                    "Lpos": 8
                  },
                  "Meta": null,
                  "Children": [],
                  "Runtime": {}
                }
              ],
              "Runtime": {}
            }
          ],
          "Runtime": {}
        },
        "Source": "ECALTestRuntime",
        "Trace": null,
        "Type": "foo"
      },
      "threadRunning": false
    }
  }
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	if _, err = testDebugger.HandleInput(fmt.Sprintf("cont 1 Resume")); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg.Wait()

	if evalError == nil || testlogger.String() != `
test1
test2`[1:] || evalError.Error() != "ECAL error in ECALTestRuntime: foo () (Line:3 Pos:2)" {
		t.Error("Unexpected result:", testlogger.String(), err)
		return
	}
}

func TestConcurrentDebugging(t *testing.T) {
	var err error

	defer func() {
		testDebugger = nil
	}()

	testDebugger = NewECALDebugger(nil)

	if _, err = testDebugger.HandleInput("break ECALEvalTest:5"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	erp := NewECALRuntimeProvider("ECALTestRuntime", nil, nil)
	vs := scope.NewScope(scope.GlobalScope)

	go func() {
		_, err = UnitTestEvalWithRuntimeProvider(`
a := 1
b := 1
func test1() {
	log("test3")
	b := a + 1
}
log("test1")
log("test2")
test1()
log("test4")
`, vs, erp)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	go func() {
		_, err = UnitTestEvalWithRuntimeProvider(`
a := 1
c := 1
func test2() {
	log("test3")
	c := a + 1
}
log("test1")
log("test2")
test2()
log("test4")
`, vs, erp)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	waitForAllThreadSuspension(t)

	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ := json.MarshalIndent(out, "", "  ")
	outString := string(outBytes)

	if err != nil || (outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:5": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [
        "test1() (ECALEvalTest:10)"
      ],
      "error": null,
      "threadRunning": false
    },
    "2": {
      "callStack": [
        "test2() (ECALEvalTest:10)"
      ],
      "error": null,
      "threadRunning": false
    }
  }
}` && outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:5": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [
        "test2() (ECALEvalTest:10)"
      ],
      "error": null,
      "threadRunning": false
    },
    "2": {
      "callStack": [
        "test1() (ECALEvalTest:10)"
      ],
      "error": null,
      "threadRunning": false
    }
  }
}`) {
		t.Error("Unexpected result:", outString, err)
		return
	}

	// Continue until the end

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont 1 Resume")); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont 2 Resume")); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg.Wait()

	if vs.String() != `GlobalScope {
    a (float64) : 1
    b (float64) : 2
    c (float64) : 2
    test1 (*interpreter.function) : ecal.function: test1 (Line 4, Pos 1)
    test2 (*interpreter.function) : ecal.function: test2 (Line 4, Pos 1)
}` {
		t.Error("Unexpected result:", vs)
		return
	}
}

func waitForThreadSuspension(t *testing.T) uint64 {
	var tid uint64

	for i := 0; i < 100; i += 1 {
		state, err := testDebugger.HandleInput("status")
		errorutil.AssertOk(err)

		threads := state.(map[string]interface{})["threads"].(map[string]map[string]interface{})
		if len(threads) > 0 {
			for threadId, status := range threads {

				if r, ok := status["threadRunning"]; ok && !r.(bool) {
					threadIdNum, _ := strconv.ParseInt(threadId, 10, 0)
					tid = uint64(threadIdNum)
					return tid
				}
			}
		}

		time.Sleep(1 * time.Millisecond)
	}

	panic("No suspended thread")
}

func waitForAllThreadSuspension(t *testing.T) uint64 {
	var tid uint64

	for i := 0; i < 100; i += 1 {
		state, err := testDebugger.HandleInput("status")
		errorutil.AssertOk(err)

		threads := state.(map[string]interface{})["threads"].(map[string]map[string]interface{})
		if len(threads) > 0 {
			allSuspended := true
			for _, status := range threads {
				if r, ok := status["threadRunning"]; ok && !r.(bool) {
					allSuspended = false
					break
				}
			}
			if allSuspended {
				break
			}
		}

		time.Sleep(1 * time.Millisecond)
	}

	return tid
}

func TestStepDebugging(t *testing.T) {
	var err error
	defer func() {
		testDebugger = nil
	}()

	testDebugger = NewECALDebugger(nil)

	code := `
log("start")
func fa(x) {
  a := 1
  log("a enter")
  fb(x)
  log("a exit")
}
func fb(x) {
  b := 2
  log("b enter")
  fc()
  fc(fc())
  log("b exit")
}
func fc() {
  c := 3
  log("c enter")
  log("c exit")
}
fa(1)
func e() {
  log("e()")
}
func d() {
  e()
}
d(d())
log("finish")
`

	if _, err = testDebugger.HandleInput("break ECALEvalTest:10"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("breakonstart true"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err = UnitTestEval(code, nil)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	tid := waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "log(\"start\")",
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {}
}` {
		t.Error("Unexpected state:", state)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "b := 2",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "x": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Step in without a function

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepin", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "log(\"b enter\")",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "b": 2,
    "x": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Normal step over

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepover", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "fc()",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "b": 2,
    "x": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Normal step in

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepin", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "c := 3",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)",
        "fc() (ECALEvalTest:12)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {}
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Normal step out

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepout", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "fc(fc())",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "b": 2,
    "x": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Step in and step out - we should end up on the same line as before

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepin", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "c := 3",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)",
        "fc() (ECALEvalTest:13)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {}
}` {
		t.Error("Unexpected state:", state)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepout", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "fc(fc())",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)",
        "fb(x) (ECALEvalTest:6)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "b": 2,
    "x": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Normal step out

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepout", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true
  },
  "code": "log(\"a exit\")",
  "threads": {
    "1": {
      "callStack": [
        "fa(1) (ECALEvalTest:21)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "a": 1,
    "x": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Set a new breakpoint

	if _, err = testDebugger.HandleInput("break ECALEvalTest:28"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v Resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true,
    "ECALEvalTest:28": true
  },
  "code": "d(d())",
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "d": "ecal.function: d (Line 25, Pos 1)",
    "e": "ecal.function: e (Line 22, Pos 1)",
    "fa": "ecal.function: fa (Line 3, Pos 1)",
    "fb": "ecal.function: fb (Line 9, Pos 1)",
    "fc": "ecal.function: fc (Line 16, Pos 1)"
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Normal step over

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepover", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true,
    "ECALEvalTest:28": true
  },
  "code": "d(d())",
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "d": "ecal.function: d (Line 25, Pos 1)",
    "e": "ecal.function: e (Line 22, Pos 1)",
    "fa": "ecal.function: fa (Line 3, Pos 1)",
    "fb": "ecal.function: fb (Line 9, Pos 1)",
    "fc": "ecal.function: fc (Line 16, Pos 1)"
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v stepover", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:10": true,
    "ECALEvalTest:28": true
  },
  "code": "log(\"finish\")",
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "d": "ecal.function: d (Line 25, Pos 1)",
    "e": "ecal.function: e (Line 22, Pos 1)",
    "fa": "ecal.function: fa (Line 3, Pos 1)",
    "fb": "ecal.function: fb (Line 9, Pos 1)",
    "fc": "ecal.function: fc (Line 16, Pos 1)"
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Continue until the end

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v Resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg.Wait()

	if err != nil || testlogger.String() != `
start
a enter
b enter
c enter
c exit
c enter
c exit
c enter
c exit
b exit
a exit
e()
e()
finish`[1:] {
		t.Error("Unexpected result:", testlogger.String(), err)
		return
	}
}

func TestStepDebuggingWithImport(t *testing.T) {
	var err error
	defer func() {
		testDebugger = nil
	}()

	testDebugger = NewECALDebugger(nil)

	il := &util.MemoryImportLocator{Files: make(map[string]string)}
	il.Files["foo/bar"] = `
func myfunc(n) {
  if (n <= 1) {
      return n
  }
  n := n + 1
  return n
}
`
	code := `
a := 1
import "foo/bar" as foobar
log("start")
a := foobar.myfunc(a)
log("finish: ", a)
`

	if _, err = testDebugger.HandleInput("break ECALEvalTest:4"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("break foo/bar:4"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err = UnitTestEvalAndASTAndImport(code, nil, "", il)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	tid := waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:4": true,
    "foo/bar:4": true
  },
  "code": "log(\"start\")",
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "a": 1,
    "foobar": {
      "myfunc": "ecal.function: myfunc (Line 2, Pos 1)"
    }
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Resume execution

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}
	tid = waitForThreadSuspension(t)

	if state := getDebuggerState(tid, t); state != `{
  "breakpoints": {
    "ECALEvalTest:4": true,
    "foo/bar:4": true
  },
  "code": "return n",
  "threads": {
    "1": {
      "callStack": [
        "myfunc(a) (ECALEvalTest:5)"
      ],
      "error": null,
      "threadRunning": false
    }
  },
  "vs": {
    "n": 1
  }
}` {
		t.Error("Unexpected state:", state)
		return
	}

	// Continue until the end

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v Resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg.Wait()

	if err != nil || testlogger.String() != `
start
finish: 1`[1:] {
		t.Error("Unexpected result:", testlogger.String(), err)
		return
	}
}

func getDebuggerState(tid uint64, t *testing.T) string {
	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))
	if err != nil {
		t.Error(err)
		return ""
	}

	outMap := out.(map[string]interface{})

	out, err = testDebugger.HandleInput(fmt.Sprintf("describe %v", tid))
	if err != nil {
		t.Error(err)
		return ""
	}
	outMap2 := out.(map[string]interface{})

	outMap["vs"] = outMap2["vs"]
	outMap["code"] = outMap2["code"]

	delete(outMap, "breakonstart")
	delete(outMap, "sources")

	outBytes, _ := json.MarshalIndent(outMap, "", "  ")
	return string(outBytes)

}

func TestInjectAndExtractDebugging(t *testing.T) {
	var err error

	defer func() {
		testDebugger = nil
	}()

	vs := scope.NewScope(scope.GlobalScope)

	testDebugger = NewECALDebugger(vs)

	if _, err = testDebugger.HandleInput("break ECALEvalTest:5"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err = UnitTestEval(`
b := 49
func myfunc() {
	a := 56
	log("test2 a=", a)
}
log("test1")
myfunc()
log("test3 b=", b)
`, vs)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	tid := waitForThreadSuspension(t)

	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ := json.MarshalIndent(out, "", "  ")
	outString := string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:5": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [
        "myfunc() (ECALEvalTest:8)"
      ],
      "error": null,
      "threadRunning": false
    }
  }
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	out, err = testDebugger.HandleInput(fmt.Sprintf("describe %v", tid))

	outBytes, _ = json.MarshalIndent(out, "", "  ")
	outString = string(outBytes)

	if err != nil || outString != `{
  "callStack": [
    "myfunc() (ECALEvalTest:8)"
  ],
  "callStackNode": [
    {
      "allowescapes": false,
      "children": [
        {
          "name": "funccall"
        }
      ],
      "id": 7,
      "identifier": true,
      "line": 8,
      "linepos": 1,
      "name": "identifier",
      "pos": 69,
      "source": "ECALEvalTest",
      "value": "myfunc"
    }
  ],
  "callStackVsSnapshot": [
    {
      "b": 49,
      "myfunc": "ecal.function: myfunc (Line 3, Pos 1)"
    }
  ],
  "callStackVsSnapshotGlobal": [
    {
      "b": 49,
      "myfunc": "ecal.function: myfunc (Line 3, Pos 1)"
    }
  ],
  "code": "log(\"test2 a=\", a)",
  "error": null,
  "node": {
    "allowescapes": false,
    "children": [
      {
        "children": [
          {
            "allowescapes": true,
            "id": 5,
            "identifier": false,
            "line": 5,
            "linepos": 6,
            "name": "string",
            "pos": 39,
            "source": "ECALEvalTest",
            "value": "test2 a="
          },
          {
            "allowescapes": false,
            "id": 7,
            "identifier": true,
            "line": 5,
            "linepos": 18,
            "name": "identifier",
            "pos": 51,
            "source": "ECALEvalTest",
            "value": "a"
          }
        ],
        "name": "funccall"
      }
    ],
    "id": 7,
    "identifier": true,
    "line": 5,
    "linepos": 2,
    "name": "identifier",
    "pos": 35,
    "source": "ECALEvalTest",
    "value": "log"
  },
  "threadRunning": false,
  "vs": {
    "a": 56
  },
  "vsGlobal": {
    "b": 49,
    "myfunc": "ecal.function: myfunc (Line 3, Pos 1)"
  }
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("extract %v a foo", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := testDebugger.HandleInput(fmt.Sprintf("inject %v a x := b + 1; x", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	// Continue until the end

	if _, err := testDebugger.HandleInput(fmt.Sprintf("cont %v Resume", tid)); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg.Wait()

	if vs.String() != `
GlobalScope {
    b (float64) : 49
    foo (float64) : 56
    myfunc (*interpreter.function) : ecal.function: myfunc (Line 3, Pos 1)
}`[1:] {
		t.Error("Unexpected result:", vs.String(), err)
		return
	}

	if testlogger.String() != `
test1
test2 a=50
test3 b=49`[1:] {
		t.Error("Unexpected result:", testlogger.String(), err)
		return
	}
}

func TestSimpleStacktrace(t *testing.T) {

	res, err := UnitTestEval(`
func a() {
	b()
}
func b() {
	c()
}
func c() {
	raise("testerror")
}
a()
`, nil)

	if err == nil {
		t.Error("Unexpected result: ", res, err)
		return
	}

	ss := err.(util.TraceableRuntimeError)

	if out := fmt.Sprintf("%v\n  %v", err.Error(), strings.Join(ss.GetTraceString(), "\n  ")); out != `
ECAL error in ECALTestRuntime: testerror () (Line:9 Pos:2)
  raise("testerror") (ECALEvalTest:9)
  c() (ECALEvalTest:6)
  b() (ECALEvalTest:3)
  a() (ECALEvalTest:11)`[1:] {
		t.Error("Unexpected output:", out)
		return
	}
}

func TestDebugDocstrings(t *testing.T) {
	for k, v := range DebugCommandsMap {
		if res := v.DocString(); res == "" {
			t.Error("Docstring missing for ", k)
			return
		}
	}
}

func TestDebuggingErrorInput(t *testing.T) {
	var err error

	defer func() {
		testDebugger = nil
	}()

	vs := scope.NewScope(scope.GlobalScope)

	testDebugger = NewECALDebugger(vs)

	if _, err = testDebugger.HandleInput("uuu"); err == nil ||
		err.Error() != `Unknown command: uuu` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("break"); err == nil ||
		err.Error() != `Need a break target (<source>:<line>) as first parameter` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("break foo"); err == nil ||
		err.Error() != `Invalid break target - should be <source>:<line>` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("rmbreak"); err == nil ||
		err.Error() != `Need a break target (<source>[:<line>]) as first parameter` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("disablebreak"); err == nil ||
		err.Error() != `Need a break target (<source>:<line>) as first parameter` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("disablebreak foo"); err == nil ||
		err.Error() != `Invalid break target - should be <source>:<line>` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("break ECALEvalTest:3"); err != nil {
		t.Error("Unexpected result:", err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err = UnitTestEval(`
a:=1
log("test1")
log("test2")
log("test3")
`, vs)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	tid := waitForThreadSuspension(t)

	out, err := testDebugger.HandleInput(fmt.Sprintf("status"))

	outBytes, _ := json.MarshalIndent(out, "", "  ")
	outString := string(outBytes)

	if err != nil || outString != `{
  "breakonstart": false,
  "breakpoints": {
    "ECALEvalTest:3": true
  },
  "sources": [
    "ECALEvalTest"
  ],
  "threads": {
    "1": {
      "callStack": [],
      "error": null,
      "threadRunning": false
    }
  }
}` {
		t.Error("Unexpected result:", outString, err)
		return
	}

	if _, err = testDebugger.HandleInput("cont foo"); err == nil ||
		err.Error() != `Need a thread ID and a command Resume, StepIn, StepOver or StepOut` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("cont foo bar"); err == nil ||
		err.Error() != `Parameter 1 should be a number` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("cont 99 bar"); err == nil ||
		err.Error() != `Invalid command bar - must be resume, stepin, stepover or stepout` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput("describe"); err == nil ||
		err.Error() != `Need a thread ID` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput(fmt.Sprintf("extract %v foo", tid)); err == nil ||
		err.Error() != `Need a thread ID, a variable name and a destination variable name` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput(fmt.Sprintf("extract %v _foo foo", tid)); err == nil ||
		err.Error() != `Variable names may only contain [a-zA-Z] and [a-zA-Z0-9] from the second character` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput(fmt.Sprintf("extract %v foo foo", tid)); err == nil ||
		err.Error() != `No such value foo` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput(fmt.Sprintf("inject %v", tid)); err == nil ||
		err.Error() != `Need a thread ID, a variable name and an expression` {
		t.Error("Unexpected result:", err)
		return
	}

	testDebugger.(*ecalDebugger).globalScope = nil

	if _, err = testDebugger.HandleInput(fmt.Sprintf("extract %v foo foo", tid)); err == nil ||
		err.Error() != `Cannot access global scope` {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err = testDebugger.HandleInput(fmt.Sprintf("inject %v foo foo", tid)); err == nil ||
		err.Error() != `Cannot access global scope` {
		t.Error("Unexpected result:", err)
		return
	}
}
