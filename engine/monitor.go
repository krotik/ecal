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
	"container/heap"
	"fmt"
	"sync"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/sortutil"
	"devt.de/krotik/ecal/engine/pubsub"
)

/*
Monitor monitors events as they are cascading. Event cascades will produce tree
structures.
*/
type Monitor interface {

	/*
	   ID returns the monitor ID.
	*/
	ID() uint64

	/*
	   NewChildMonitor creates a new child monitor of this monitor.
	*/
	NewChildMonitor(priority int) Monitor

	/*
	   Scope returns the rule scope of this monitor.
	*/
	Scope() *RuleScope

	/*
	   Priority returns the monitors priority.
	*/
	Priority() int

	/*
		Activated returns if this monitor has been activated.
	*/
	IsActivated() bool

	/*
		Activate activates this monitor with a given event.
	*/
	Activate(e *Event)

	/*
	   Skip finishes this monitor without activation.
	*/
	Skip(e *Event)

	/*
		Finish finishes this monitor.
	*/
	Finish()

	/*
		Returns the root monitor of this monitor.
	*/
	RootMonitor() *RootMonitor

	/*
		Errors returns the error object of this monitor.
	*/
	Errors() *TaskError

	/*
		SetErrors adds an error object to this monitor.
	*/
	SetErrors(e *TaskError)

	/*
		EventPath returns the chain of events which created this monitor.
	*/
	EventPath() []*Event

	/*
	   EventPathString returns the event path as a string.
	*/
	EventPathString() string

	/*
		String returns a string representation of this monitor.
	*/
	String() string
}

/*
monitorBase provides the basic functions and fields for any monitor.
*/
type monitorBase struct {
	id      uint64                 // Monitor ID
	Parent  *monitorBase           // Parent monitor
	Context map[string]interface{} // Context object of the monitor
	Err     *TaskError             // Errors which occurred during event processing

	priority    int          // Priority of the monitor
	rootMonitor *RootMonitor // Root monitor
	event       *Event       // Event which activated this monitor
	activated   bool         // Flag indicating if the monitor was activated
	finished    bool         // Flag indicating if the monitor has finished
}

/*
NewMonitor creates a new monitor.
*/
func newMonitorBase(priority int, parent *monitorBase, context map[string]interface{}) *monitorBase {

	var ret *monitorBase

	if parent != nil {
		ret = &monitorBase{newMonID(), parent, context, nil, priority, parent.rootMonitor, nil, false, false}
	} else {
		ret = &monitorBase{newMonID(), nil, context, nil, priority, nil, nil, false, false}
	}

	return ret
}

/*
NewChildMonitor creates a new child monitor of this monitor.
*/
func (mb *monitorBase) NewChildMonitor(priority int) Monitor {
	child := &ChildMonitor{newMonitorBase(priority, mb, mb.Context)}

	mb.rootMonitor.descendantCreated(child)

	return child
}

/*
ID returns the monitor ID.
*/
func (mb *monitorBase) ID() uint64 {
	return mb.id
}

/*
RootMonitor returns the root monitor of this monitor.
*/
func (mb *monitorBase) RootMonitor() *RootMonitor {
	return mb.rootMonitor
}

/*
Scope returns the rule scope of this monitor.
*/
func (mb *monitorBase) Scope() *RuleScope {
	return mb.rootMonitor.ruleScope
}

/*
Priority returns the priority of this monitor.
*/
func (mb *monitorBase) Priority() int {
	return mb.priority
}

/*
IsActivated returns if this monitor has been activated.
*/
func (mb *monitorBase) IsActivated() bool {
	return mb.activated
}

/*
IsFinished returns if this monitor has finished.
*/
func (mb *monitorBase) IsFinished() bool {
	return mb.finished
}

/*
Activate activates this monitor.
*/
func (mb *monitorBase) Activate(e *Event) {
	errorutil.AssertTrue(!mb.finished, "Cannot activate a finished monitor")
	errorutil.AssertTrue(!mb.activated, "Cannot activate an active monitor")
	errorutil.AssertTrue(e != nil, "Monitor can only be activated with an event")

	mb.event = e
	mb.rootMonitor.descendantActivated(mb.priority)
	mb.activated = true
}

/*
Skip finishes this monitor without activation.
*/
func (mb *monitorBase) Skip(e *Event) {
	errorutil.AssertTrue(!mb.finished, "Cannot skip a finished monitor")
	errorutil.AssertTrue(!mb.activated, "Cannot skip an active monitor")

	mb.event = e
	mb.activated = true
	mb.Finish()
}

/*
Finish finishes this monitor.
*/
func (mb *monitorBase) Finish() {
	errorutil.AssertTrue(mb.activated, "Cannot finish a not active monitor")
	errorutil.AssertTrue(!mb.finished, "Cannot finish a finished monitor")

	mb.finished = true
	mb.rootMonitor.descendantFinished(mb)
}

/*
Errors returns the error object of this monitor.
*/
func (mb *monitorBase) Errors() *TaskError {
	errorutil.AssertTrue(mb.finished, "Cannot get errors on an unfinished monitor")
	return mb.Err
}

/*
SetErrors adds an error object to this monitor.
*/
func (mb *monitorBase) SetErrors(e *TaskError) {
	mb.Err = e
	mb.rootMonitor.descendantFailed(mb)
}

/*
EventPath returns the chain of events which created this monitor.
*/
func (mb *monitorBase) EventPath() []*Event {
	errorutil.AssertTrue(mb.finished, "Cannot get event path on an unfinished monitor")

	path := []*Event{mb.event}

	child := mb.Parent
	for child != nil {
		path = append(path, child.event)
		child = child.Parent
	}

	// Reverse path

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

/*
EventPathString returns the event path as a string.
*/
func (mb *monitorBase) EventPathString() string {
	var buf bytes.Buffer

	ep := mb.EventPath()
	last := len(ep) - 1

	for i, e := range mb.EventPath() {
		buf.WriteString(e.name)
		if i < last {
			buf.WriteString(" -> ")
		}
	}

	return buf.String()
}

/*
String returns a string representation of this monitor.
*/
func (mb *monitorBase) String() string {
	return fmt.Sprintf("Monitor %v (parent: %v priority: %v activated: %v finished: %v)",
		mb.ID(), mb.Parent, mb.priority, mb.activated, mb.finished)
}

// Root Monitor
// ============

/*
RootMonitor is a monitor which is at a beginning of an event cascade.
*/
type RootMonitor struct {
	*monitorBase
	lock         *sync.Mutex             // Lock for datastructures
	incomplete   map[int]int             // Priority -> Counters of incomplete trackers
	priorities   *sortutil.IntHeap       // List of handled priorities
	ruleScope    *RuleScope              // Rule scope definitions
	unfinished   int                     // Counter of all unfinished trackers
	messageQueue *pubsub.EventPump       // Message passing queue of the processor
	errors       map[uint64]*monitorBase // Monitors which got errors
	finished     func(Processor)         // Finish handler (can be used externally)
}

/*
NewRootMonitor creates a new root monitor.
*/
func newRootMonitor(context map[string]interface{}, scope *RuleScope,
	messageQueue *pubsub.EventPump) *RootMonitor {

	ret := &RootMonitor{newMonitorBase(0, nil, context), &sync.Mutex{},
		make(map[int]int), &sortutil.IntHeap{}, scope, 1, messageQueue,
		make(map[uint64]*monitorBase), nil}

	// A root monitor is its own parent

	ret.rootMonitor = ret

	heap.Init(ret.priorities)

	return ret
}

/*
SetFinishHandler adds a handler function to this monitor which is called once
this monitor has finished.
*/
func (rm *RootMonitor) SetFinishHandler(fh func(Processor)) {
	rm.finished = fh
}

/*
HighestPriority returns the highest priority which is handled by this monitor.
*/
func (rm *RootMonitor) HighestPriority() int {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	if len(*rm.priorities) > 0 {
		return (*rm.priorities)[0]
	}

	return -1
}

/*
AllErrors returns all error which have been collected in this root monitor.
*/
func (rm *RootMonitor) AllErrors() []*TaskError {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	ret := make([]*TaskError, 0, len(rm.errors))

	// Sort by monitor id - this should give the corrent order timewise

	var ids []uint64
	for id := range rm.errors {
		ids = append(ids, id)
	}

	sortutil.UInt64s(ids)

	for _, id := range ids {
		m := rm.errors[id]
		ret = append(ret, m.Errors())
	}

	return ret
}

/*
descendantCreated notifies this root monitor that a descendant has been created.
*/
func (rm *RootMonitor) descendantCreated(monitor Monitor) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	rm.unfinished++
}

/*
descendantActivated notifies this root monitor that a descendant has been activated.
*/
func (rm *RootMonitor) descendantActivated(priority int) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	val, ok := rm.incomplete[priority]
	if !ok {
		val = 0
		heap.Push(rm.priorities, priority)
	}

	rm.incomplete[priority] = val + 1
}

/*
descendantFailed notifies this root monitor that a descendant has failed.
*/
func (rm *RootMonitor) descendantFailed(monitor *monitorBase) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	rm.errors[monitor.ID()] = monitor
}

/*
descendantFinished records that this monitor has finished. If it is the last
active monitor in the event tree then send a notification.
*/
func (rm *RootMonitor) descendantFinished(m Monitor) {

	rm.lock.Lock()

	rm.unfinished--

	finished := rm.unfinished == 0

	if m.IsActivated() {
		priority := m.Priority()

		rm.incomplete[priority]--

		if rm.incomplete[priority] == 0 {
			rm.priorities.RemoveFirst(priority)
			delete(rm.incomplete, priority)
		}
	}

	rm.lock.Unlock()

	// Post notification

	if finished {
		rm.messageQueue.PostEvent(MessageRootMonitorFinished, rm)
	}
}

// Child Monitor
// =============

/*
ChildMonitor is a monitor which is a descendant of a root monitor.
*/
type ChildMonitor struct {
	*monitorBase
}

// Unique id creation
// ==================

var midcounter uint64 = 1
var midcounterLock = &sync.Mutex{}

/*
newId returns a new unique id.
*/
func newMonID() uint64 {
	midcounterLock.Lock()
	defer midcounterLock.Unlock()

	ret := midcounter
	midcounter++

	return ret
}
