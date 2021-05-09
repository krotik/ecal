/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/engine/pool"
)

func TestProcessorSimpleCascade(t *testing.T) {
	UnitTestResetIDs()

	// Add debug logging

	var debugBuffer bytes.Buffer

	EventTracer.out = &debugBuffer
	EventTracer.MonitorEvent("core.*", map[interface{}]interface{}{
		"foo":  "bar",
		"foo2": nil,
	})
	EventTracer.MonitorEvent("core.*", map[interface{}]interface{}{
		"foo2": "test",
	})
	defer func() {
		EventTracer.Reset()
		EventTracer.out = os.Stdout
	}()

	// Do the normal testing

	var log bytes.Buffer

	proc := NewProcessor(1)

	StderrSave := os.Stderr
	os.Stderr = nil
	proc.ThreadPool().TooManyCallback()
	os.Stderr = StderrSave

	if res := fmt.Sprint(proc); res != "EventProcessor 1 (workers:1)" {
		t.Error("Unexpected result:", res)
		return
	}

	// Add rules to the processor

	rule1 := &Rule{
		"TestRule1",                            // Name
		"",                                     // Description
		[]string{"core.main.event1"},           // Kind match
		[]string{"data"},                       // Match on event cascade scope
		nil,                                    // No state match
		2,                                      // Priority of the rule
		[]string{"TestRule3", "TestRule3Copy"}, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			log.WriteString("TestRule1\n")

			// Add another event

			p.AddEvent(&Event{
				"InitialEvent",
				[]string{"core", "main", "event2"},
				map[interface{}]interface{}{
					"foo":  "bar",
					"foo2": "bla",
				},
			}, m.NewChildMonitor(1))

			return nil
		},
	}

	rule2 := &Rule{
		"TestRule2",             // Name
		"",                      // Description
		[]string{"core.main.*"}, // Kind match
		[]string{"data.read"},   // Match on event cascade scope
		nil,                     // No state match
		5,                       // Priority of the rule
		nil,                     // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			log.WriteString("TestRule2\n")
			return nil
		},
	}

	rule3 := &Rule{
		"TestRule3",             // Name
		"",                      // Description
		[]string{"core.main.*"}, // Kind match
		[]string{"data.read"},   // Match on event cascade scope
		nil,                     // No state match
		0,                       // Priority of the rule
		nil,                     // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			log.WriteString("TestRule3\n")
			return nil
		},
	}

	proc.AddRule(rule1)
	proc.AddRule(rule2)
	proc.AddRule(rule3)

	if r := len(proc.Rules()); r != 3 {
		t.Error("Unexpected rule number:", r)
		return
	}

	// Start the processor
	proc.Start()

	// Push a root event

	e := NewEvent(
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{
			"foo":  "bar",
			"foo2": "bla",
		},
	)

	if e.Name() != e.name || e.Kind() == nil || e.State() == nil {
		t.Error("Unepxected getter result:", e)
		return
	}

	rootm := proc.NewRootMonitor(nil, nil)
	rootm.SetFinishHandler(func(p Processor) {
		log.WriteString("finished!")
	})
	proc.AddEventAndWait(e, rootm)

	if err := proc.AddRule(rule3); err.Error() != "Cannot add rule if the processor has not stopped" {
		t.Error("Unexpected error:", err)
		return
	}

	if err := proc.Reset(); err.Error() != "Cannot reset processor if it has not stopped" {
		t.Error("Unexpected error:", err)
		return
	}

	// Finish the processor

	proc.Finish()

	// Finish the processor

	// Rule 1, 2 and 3 trigger on event1 but rule 3 is suppressed by rule 1
	// Rule 1 adds a new event which triggers only rule 2 and 3
	// Rule 3 comes first since it has the higher priority

	if log.String() != `TestRule1
TestRule2
TestRule3
TestRule2
finished!` {
		t.Error("Unexpected result:", log.String())
		return

	}

	log.Reset()

	if err := proc.AddRule(rule3.CopyAs("TestRule3Copy")); err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	// Start the processor

	proc.Start()

	// Push a root event

	proc.AddEventAndWait(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		nil,
	}, nil)

	// Finish the processor

	proc.Finish()

	if log.String() != `TestRule1
TestRule2
TestRule3
TestRule3
TestRule2
` {
		t.Error("Unexpected result:", log.String())
		return
	}

	// Test the case when the event is pointless

	log.Reset()

	proc.Start()

	proc.AddEvent(&Event{
		"InitialEventFoo",
		[]string{"core", "foo", "event1"},
		nil,
	}, nil)

	rm := proc.NewRootMonitor(nil, nil)

	proc.AddEvent(&Event{
		"InitialEventFoo",
		[]string{"core", "foo", "event1"},
		nil,
	}, rm)

	if !rm.IsFinished() {
		t.Error("Monitor which monitored a non-triggering event should still finished")
		return
	}

	proc.Finish()

	if log.String() != "" {
		t.Error("Unexpected result:", log.String())
		return
	}

	proc.Reset()

	if r := len(proc.Rules()); r != 0 {
		t.Error("Unexpected rule number:", r)
		return
	}

	if debugBuffer.String() == "" {
		t.Error("Nothing was recorded in the debug buffer")
		return
	}
}

func TestProcessorSimplePriorities(t *testing.T) {
	UnitTestResetIDs()

	var logLock = sync.Mutex{}
	var log bytes.Buffer

	testPriorities := func(p1, p2 int) int {

		proc := NewProcessor(2)

		// Add rules to the processor

		rule1 := &Rule{
			"TestRule1",                  // Name
			"",                           // Description
			[]string{"core.main.event1"}, // Kind match
			[]string{"data"},             // Match on event cascade scope
			nil,                          // No state match
			0,                            // Priority of the rule
			nil,                          // List of suppressed rules by this rule
			func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
				logLock.Lock()
				log.WriteString("TestRule1\n")
				logLock.Unlock()
				time.Sleep(2 * time.Millisecond)
				return nil
			},
		}

		rule2 := &Rule{
			"TestRule2",                  // Name
			"",                           // Description
			[]string{"core.main.event2"}, // Kind match
			[]string{"data"},             // Match on event cascade scope
			nil,                          // No state match
			0,                            // Priority of the rule
			nil,                          // List of suppressed rules by this rule
			func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
				logLock.Lock()
				log.WriteString("TestRule2\n")
				logLock.Unlock()
				time.Sleep(2 * time.Millisecond)
				return nil
			},
		}

		proc.AddRule(rule1)
		proc.AddRule(rule2)

		proc.Start()

		m := proc.NewRootMonitor(nil, nil)

		// Push a root event

		for i := 0; i < 3; i++ {
			proc.AddEvent(&Event{
				"InitialEvent1",
				[]string{"core", "main", "event1"},
				nil,
			}, m.NewChildMonitor(p1))
		}

		proc.AddEvent(&Event{
			"InitialEvent2",
			[]string{"core", "main", "event2"},
			nil,
		}, m.NewChildMonitor(p2))

		proc.AddEvent(&Event{
			"InitialEvent1",
			[]string{"core", "main", "event1"},
			nil,
		}, m.NewChildMonitor(p1))

		hp := m.HighestPriority()

		// Finish the processor

		proc.Finish()

		errorutil.AssertTrue(m.HighestPriority() == -1,
			"Highest priority should be -1 once a monitor has finished")

		return hp
	}

	// Since rule 1 has the higher priority it is more likely to be
	// executed

	if res := testPriorities(3, 5); res != 3 {
		t.Error("Unexpected highest priority:", res)
		return
	}

	if log.String() != `TestRule1
TestRule1
TestRule1
TestRule1
TestRule2
` && log.String() != `TestRule1
TestRule1
TestRule1
TestRule2
TestRule1
` {
		t.Error("Unexpected result:", log.String())
		return
	}

	log.Reset()

	// Since rule 2 has the higher priority it is more likely to be
	// executed

	if res := testPriorities(5, 2); res != 2 {
		t.Error("Unexpected highest priority:", res)
		return
	}

	if log.String() != `TestRule2
TestRule1
TestRule1
TestRule1
TestRule1
` && log.String() != `TestRule1
TestRule2
TestRule1
TestRule1
TestRule1
` && log.String() != `TestRule1
TestRule1
TestRule2
TestRule1
TestRule1
` {
		t.Error("Unexpected result:", log.String())
		return
	}
}

func TestProcessorScopeHandling(t *testing.T) {
	UnitTestResetIDs()

	var logLock = sync.Mutex{}
	var log bytes.Buffer

	proc := NewProcessor(10)

	// Add rules to the processor

	rule1 := &Rule{
		"TestRule1",             // Name
		"",                      // Description
		[]string{"core.main.*"}, // Kind match
		[]string{"data.write"},  // Match on event cascade scope
		nil,                     // No state match
		0,                       // Priority of the rule
		nil,                     // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			logLock.Lock()
			log.WriteString("TestRule1\n")
			logLock.Unlock()
			time.Sleep(2 * time.Millisecond)
			return nil
		},
	}

	rule2 := &Rule{
		"TestRule2",             // Name
		"",                      // Description
		[]string{"core.main.*"}, // Kind match
		[]string{"data"},        // Match on event cascade scope
		nil,                     // No state match
		0,                       // Priority of the rule
		nil,                     // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			logLock.Lock()
			log.WriteString("TestRule2\n")
			logLock.Unlock()
			time.Sleep(2 * time.Millisecond)
			return nil
		},
	}

	proc.AddRule(rule1)
	proc.AddRule(rule2)

	if proc.Status() != pool.StatusStopped || !proc.Stopped() {
		t.Error("Unexpected status:", proc.Status(), proc.Stopped())
		return
	}

	proc.Start()

	if proc.Status() != pool.StatusRunning || proc.Stopped() {
		t.Error("Unexpected status:", proc.Status(), proc.Stopped())
		return
	}

	scope1 := NewRuleScope(map[string]bool{
		"data":       true,
		"data.read":  true,
		"data.write": false,
	})

	m := proc.NewRootMonitor(nil, scope1)

	// Push a root event

	proc.AddEvent(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		nil,
	}, m)

	// Finish the processor

	proc.Finish()

	// Only rule 2 should trigger since the monitor has only access
	// to data and data.read

	if log.String() != `TestRule2
` {
		t.Error("Unexpected result:", log.String())
		return
	}

	log.Reset()

	proc.Start()

	scope2 := NewRuleScope(map[string]bool{
		"data":       true,
		"data.read":  true,
		"data.write": true,
	})

	m = proc.NewRootMonitor(nil, scope2)

	// Push a root event

	proc.AddEvent(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		nil,
	}, m)

	// Finish the processor

	proc.Finish()

	// Now both rules should trigger

	if log.String() != `TestRule1
TestRule2
` {
		t.Error("Unexpected result:", log.String())
		return
	}
}

func TestProcessorStateMatching(t *testing.T) {
	UnitTestResetIDs()

	var logLock = sync.Mutex{}
	var log bytes.Buffer

	proc := NewProcessor(10)

	if res := proc.Workers(); res != 10 {
		t.Error("Unexpected number of workers:", res)
		return
	}

	// Add rules to the processor

	rule1 := &Rule{
		"TestRule1",             // Name
		"",                      // Description
		[]string{"core.main.*"}, // Kind match
		[]string{"data"},        // Match on event cascade scope
		map[string]interface{}{"name": nil, "test": 1}, // Simple state match
		0,   // Priority of the rule
		nil, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			logLock.Lock()
			log.WriteString("TestRule1\n")
			logLock.Unlock()
			time.Sleep(2 * time.Millisecond)
			return nil
		},
	}

	rule2 := &Rule{
		"TestRule2",             // Name
		"",                      // Description
		[]string{"core.main.*"}, // Kind match
		[]string{"data"},        // Match on event cascade scope
		map[string]interface{}{"name": nil, "test": "123"}, // Simple state match
		0,   // Priority of the rule
		nil, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			logLock.Lock()
			log.WriteString("TestRule2\n")
			logLock.Unlock()
			time.Sleep(2 * time.Millisecond)
			return nil
		},
	}

	proc.AddRule(rule1)
	proc.AddRule(rule2)

	proc.Start()

	// Push a root event

	proc.AddEvent(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"name": "foo", "test": "123"},
	}, nil)

	proc.Finish()

	if log.String() != `TestRule2
` {
		t.Error("Unexpected result:", log.String())
		return
	}

	proc.Start()

	proc.AddEvent(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"name": nil, "test": 1, "foobar": 123},
	}, nil)

	proc.AddEvent(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"name": "bar", "test": 1},
	}, nil)

	// The following rule should not trigger as it is missing name

	proc.AddEvent(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"foobar": nil, "test": "123"},
	}, nil)

	proc.Finish()

	if log.String() != `TestRule2
TestRule1
TestRule1
` {
		t.Error("Unexpected result:", log.String())
		return
	}
}

func TestProcessorSimpleErrorHandling(t *testing.T) {
	UnitTestResetIDs()

	proc := NewProcessor(10)

	if proc.ThreadPool() == nil {
		t.Error("Should have a thread pool")
		return
	}

	// Add rules to the processor

	rule1 := &Rule{
		"TestRule1",                  // Name
		"",                           // Description
		[]string{"core.main.event1"}, // Kind match
		[]string{"data"},             // Match on event cascade scope
		nil,
		0,   // Priority of the rule
		nil, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			p.AddEvent(&Event{
				"event2",
				[]string{"core", "main", "event2"},
				nil,
			}, m.NewChildMonitor(1))
			return errors.New("testerror")
		},
	}

	rule2 := &Rule{
		"TestRule2",                  // Name
		"",                           // Description
		[]string{"core.main.event2"}, // Kind match
		[]string{"data"},             // Match on event cascade scope
		nil,
		0,   // Priority of the rule
		nil, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			p.AddEvent(&Event{
				"event3",
				[]string{"core", "main", "event3"},
				nil,
			}, m.NewChildMonitor(1))
			return nil
		},
	}

	rule3 := &Rule{
		"TestRule3", // Name
		"",          // Description
		[]string{"core.main.event3", "core.main.event1"}, // Kind match
		[]string{"data"}, // Match on event cascade scope
		nil,
		0,   // Priority of the rule
		nil, // List of suppressed rules by this rule
		func(p Processor, m Monitor, e *Event, tid uint64) error { // Action of the rule
			return errors.New("testerror2")
		},
	}

	// Add rule 1 twice

	proc.AddRule(rule1)
	proc.AddRule(rule1.CopyAs("TestRule1Copy"))

	proc.AddRule(rule2)
	proc.AddRule(rule3)

	recordedErrors := 0
	proc.SetRootMonitorErrorObserver(func(rm *RootMonitor) {
		recordedErrors = len(rm.AllErrors()[0].ErrorMap)
	})

	// First test will always execute all rules and collect all errors

	proc.SetFailOnFirstErrorInTriggerSequence(false)

	proc.Start()

	// Push a root event

	mon, err := proc.AddEventAndWait(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"name": "foo", "test": "123"},
	}, nil)

	rmon, ok := mon.(*RootMonitor)
	if !ok {
		t.Error("Root monitor expected:", mon, err)
		return
	}

	proc.Finish()

	if fmt.Sprint(mon) != "Monitor 1 (parent: <nil> priority: 0 activated: true finished: true)" {
		t.Error("Unexpected result:", mon)
		return
	}

	_, err = proc.AddEvent(&Event{}, nil)
	if err.Error() != "Cannot add event if the processor is stopping or not running" {
		t.Error("Unexpected error", err)
		return
	}

	// Two errors should have been collected

	errs := rmon.AllErrors()

	if len(errs) != 3 {
		t.Error("Unexpected number of errors:", len(errs))
		return
	}

	if recordedErrors != 3 {
		t.Error("Unexpected number of recorded errors:", recordedErrors)
		return
	}

	if fmt.Sprint(errs) != `[Taskerrors:
InitialEvent -> TestRule1 : testerror
InitialEvent -> TestRule1Copy : testerror
InitialEvent -> TestRule3 : testerror2 Taskerror:
InitialEvent -> event2 -> event3 -> TestRule3 : testerror2 Taskerror:
InitialEvent -> event2 -> event3 -> TestRule3 : testerror2]` {
		t.Error("Unexpected errors:", errs)
		return
	}

	testProcessorAdvancedErrorHandling(t, proc, &recordedErrors)
}

func testProcessorAdvancedErrorHandling(t *testing.T, proc Processor, recordedErrorsPtr *int) {

	// Second test will fail on the first failed rule in an event trigger sequence

	proc.SetFailOnFirstErrorInTriggerSequence(true)

	proc.Start()

	mon, err := proc.AddEventAndWait(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"name": "foo", "test": "123"},
	}, nil)
	rmon, ok := mon.(*RootMonitor)
	if !ok {
		t.Error("Root monitor expected:", mon, err)
		return
	}
	proc.Finish()

	errs := rmon.AllErrors()

	if len(errs) != 2 {
		t.Error("Unexpected number of errors:", len(errs))
		return
	}

	if fmt.Sprint(errs) != `[Taskerror:
InitialEvent -> TestRule1 : testerror Taskerror:
InitialEvent -> event2 -> event3 -> TestRule3 : testerror2]` {
		t.Error("Unexpected errors:", errs)
		return
	}

	if *recordedErrorsPtr != 1 {
		t.Error("Unexpected number of recorded errors:", *recordedErrorsPtr)
		return
	}

	// Now test AddEventAndWait

	proc.SetFailOnFirstErrorInTriggerSequence(false)
	proc.Start()

	mon, err = proc.AddEventAndWait(&Event{
		"InitialEvent1",
		[]string{"core", "main", "event5"},
		map[interface{}]interface{}{"name": "foo", "test": "123"},
	}, nil)

	if mon != nil || err != nil {
		t.Error("Nothing should have triggered: ", err)
		return
	}

	// Push a root event

	mon, err = proc.AddEventAndWait(&Event{
		"InitialEvent",
		[]string{"core", "main", "event1"},
		map[interface{}]interface{}{"name": "foo", "test": "123"},
	}, nil)

	rmon, ok = mon.(*RootMonitor)
	if !ok {
		t.Error("Root monitor expected:", mon, err)
		return
	}

	if fmt.Sprint(mon) != "Monitor 10 (parent: <nil> priority: 0 activated: true finished: true)" {
		t.Error("Unexpected result:", mon)
		return
	}

	if proc.Stopped() {
		t.Error("Processor should not be stopped at this point")
		return
	}

	errs = rmon.AllErrors()

	if len(errs) != 3 {
		t.Error("Unexpected number of errors:", len(errs))
		return
	}

	if *recordedErrorsPtr != 3 {
		t.Error("Unexpected number of recorded errors:", *recordedErrorsPtr)
		return
	}

	if fmt.Sprint(errs) != `[Taskerrors:
InitialEvent -> TestRule1 : testerror
InitialEvent -> TestRule1Copy : testerror
InitialEvent -> TestRule3 : testerror2 Taskerror:
InitialEvent -> event2 -> event3 -> TestRule3 : testerror2 Taskerror:
InitialEvent -> event2 -> event3 -> TestRule3 : testerror2]` {
		t.Error("Unexpected errors:", errs)
		return
	}

	proc.Finish()
}
