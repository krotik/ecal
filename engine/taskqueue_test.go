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
	"fmt"
	"testing"
)

func TestTaskQueue(t *testing.T) {
	UnitTestResetIDs()

	// Create dummy processor

	proc := NewProcessor(1)

	// Create dummy event

	event := &Event{
		"DummyEvent",
		[]string{"main"},
		nil,
	}

	// Create different root monitors with different IDs

	m1 := newRootMonitor(nil, NewRuleScope(map[string]bool{"": true}), proc.(*eventProcessor).messageQueue)

	// Create now different tasks which come from the different monitors

	t1 := &Task{proc, m1, event}

	tq := NewTaskQueue(proc.(*eventProcessor).messageQueue)

	tq.Push(t1)

	if res := tq.Size(); res != 1 {
		t.Error("Unexpected size:", res)
		return
	}

	tq.Clear()

	if res := tq.Size(); res != 0 {
		t.Error("Unexpected size:", res)
		return
	}

	if e := tq.Pop(); e != nil {
		t.Error("Unexpected event:", e)
		return
	}

	if res := tq.Size(); res != 0 {
		t.Error("Unexpected size:", res)
		return
	}

	testTaskQueuePushPop(t, tq, proc, event, t1)
}

func testTaskQueuePushPop(t *testing.T, tq *TaskQueue, proc Processor, event *Event, t1 *Task) {

	m2 := newRootMonitor(nil, NewRuleScope(map[string]bool{"": true}), proc.(*eventProcessor).messageQueue)
	m3 := newRootMonitor(nil, NewRuleScope(map[string]bool{"": true}), proc.(*eventProcessor).messageQueue)

	t2 := &Task{proc, m2, event}
	t3 := &Task{proc, m3, event}
	t4 := &Task{proc, m2.NewChildMonitor(5), event}
	t5 := &Task{proc, m2.NewChildMonitor(10), event}

	tq.Push(t1)
	tq.Push(t2)
	tq.Push(t3)
	tq.Push(t4)
	tq.Push(t5)

	if res := len(tq.queues); res != 3 {
		t.Error("Unexpected size:", res)
		return
	}

	if s := tq.queues[1].Size(); s != 1 {
		t.Error("Unexpected result:", s)
		return
	}

	if s := tq.queues[2].Size(); s != 3 {
		t.Error("Unexpected result:", s)
		return
	}

	if e := tq.Pop(); e != t1 && e != t2 && e != t3 {
		t.Error("Unexpected event:", e)
		return
	}

	if res := len(tq.queues); res != 3 {
		t.Error("Unexpected size:", res)
		return
	}

	tq.Pop()

	if res := len(tq.queues); res != 3 && res != 2 {
		t.Error("Unexpected size:", res)
		return
	}

	tq.Pop()

	if s := tq.Size(); s != 2 {
		t.Error("Unexpected result:", s)
		return
	}

	tq.Pop()

	tq.Pop()

	if s := tq.Size(); s != 0 {
		t.Error("Unexpected result:", s)
		return
	}

	if e := tq.Pop(); e != nil {
		t.Error("Unexpected event:", e)
		return
	}

	testTaskQueueMisc(t, tq, t5)
}

func testTaskQueueMisc(t *testing.T, tq *TaskQueue, t5 *Task) {
	tq.Push(t5)

	if fmt.Sprint(tq.queues) != "map[2:[ Task: EventProcessor 1 (workers:1) Monitor 5 (parent: Monitor 2 (parent: <nil> priority: 0 activated: false finished: false) priority: 10 activated: false finished: false) Event: DummyEvent main {} (10) ]]" {
		t.Error("Unexpected queue:", tq.queues)
		return
	}

	if e := tq.Pop(); e != t5 {
		t.Error("Unexpected event:", e)
		return
	}

	if fmt.Sprint(tq.queues) != "map[2:[ ]]" {
		t.Error("Unexpected list of ids:", tq.queues)
		return
	}

	if e := tq.Pop(); e != nil {
		t.Error("Unexpected event:", e)
		return
	}

	if fmt.Sprint(tq.queues) != "map[]" {
		t.Error("Unexpected list of ids:", tq.queues)
		return
	}

	if e := tq.Pop(); e != nil {
		t.Error("Unexpected event:", e)
		return
	}
}

func TestTaskQueueCorrectPriorities(t *testing.T) {
	UnitTestResetIDs()

	// Create dummy processor

	proc := NewProcessor(1)

	// Create dummy event

	event := &Event{
		"DummyEvent",
		[]string{"main"},
		nil,
	}

	// Create different root monitors with different IDs

	m1 := newRootMonitor(nil, NewRuleScope(map[string]bool{"": true}), proc.(*eventProcessor).messageQueue)

	// Create now different tasks which come from the different monitors

	t1 := &Task{proc, m1, event}
	t2 := &Task{proc, m1.NewChildMonitor(5), event}
	t3 := &Task{proc, m1.NewChildMonitor(10), event}

	tq := NewTaskQueue(proc.(*eventProcessor).messageQueue)

	tq.Push(t2)
	tq.Push(t1)
	tq.Push(t3)

	if s := tq.Size(); s != 3 {
		t.Error("Unexpected result:", s)
		return
	}

	var popList []int

	popList = append(popList, tq.Pop().(*Task).m.Priority())
	popList = append(popList, tq.Pop().(*Task).m.Priority())
	popList = append(popList, tq.Pop().(*Task).m.Priority())

	if fmt.Sprint(popList) != "[0 5 10]" {
		t.Error("Unexpected poplist:", popList)
		return
	}
}
