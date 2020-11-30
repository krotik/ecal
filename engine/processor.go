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
	"sync"

	"devt.de/krotik/ecal/engine/pool"
	"devt.de/krotik/ecal/engine/pubsub"
)

/*
Processor is the main object of the event engine. It coordinates the thread pool
and rule index. Rules can only be added if the processor is stopped. Events
can only be added if the processor is not stopped.
*/
type Processor interface {

	/*
	   ID returns the processor ID.
	*/
	ID() uint64

	/*
	   ThreadPool returns the thread pool which this processor is using.
	*/
	ThreadPool() *pool.ThreadPool

	/*
	   Workers returns the number of threads of this processor.
	*/
	Workers() int

	/*
	   Reset removes all stored rules from this processor.
	*/
	Reset() error

	/*
	   AddRule adds a new rule to the processor.
	*/
	AddRule(rule *Rule) error

	/*
	   Rules returns all loaded rules.
	*/
	Rules() map[string]*Rule

	/*
	   Start starts this processor.
	*/
	Start()

	/*
	   Finish will finish all remaining tasks and then stop the processor.
	*/
	Finish()

	/*
	   Stopped returns if the processor is stopped.
	*/
	Stopped() bool

	/*
	   Status returns the status of the processor (Running / Stopping / Stopped).
	*/
	Status() string

	/*
	   NewRootMonitor creates a new root monitor for this processor. This monitor is used to add initial
	   root events.
	*/
	NewRootMonitor(context map[string]interface{}, scope *RuleScope) *RootMonitor

	/*
		SetRootMonitorErrorObserver specifies an observer which is triggered
		when a root monitor of this processor has finished and returns errors.
		By default this is set to nil (no observer).
	*/
	SetRootMonitorErrorObserver(func(rm *RootMonitor))

	/*
		SetFailOnFirstErrorInTriggerSequence sets the behavior when rules return errors.
		If set to false (default) then all rules in a trigger sequence for a specific event
		are executed. If set to true then the first rule which returns an error will stop
		the trigger sequence. Events which have been added by the failing rule are still processed.
	*/
	SetFailOnFirstErrorInTriggerSequence(bool)

	/*
	   AddEventAndWait adds a new event to the processor and waits for the resulting event cascade
	   to finish. If a monitor is passed then it must be a RootMonitor.
	*/
	AddEventAndWait(event *Event, monitor *RootMonitor) (Monitor, error)

	/*
	   AddEvent adds a new event to the processor. Returns the monitor if the event
	   triggered a rule and nil if the event was skipped.
	*/
	AddEvent(event *Event, parentMonitor Monitor) (Monitor, error)

	/*
	   IsTriggering checks if a given event triggers a loaded rule. This does not the
	   actual state matching for speed.
	*/
	IsTriggering(event *Event) bool

	/*
		ProcessEvent processes an event by determining which rules trigger and match
		the given event. This function must receive a unique thread ID from the
		executing thread.
	*/
	ProcessEvent(tid uint64, event *Event, parent Monitor) map[string]error

	/*
	   String returns a string representation the processor.
	*/
	String() string
}

/*
eventProcessor main implementation of the Processor interface.

Event cycle:

Process -> Triggering -> Matching -> Fire Rule

*/
type eventProcessor struct {
	id                  uint64                // Processor ID
	pool                *pool.ThreadPool      // Thread pool of this processor
	workerCount         int                   // Number of threads for this processor
	failOnFirstError    bool                  // Stop rule execution on first error in an event trigger sequence
	ruleIndex           RuleIndex             // Container for loaded rules
	triggeringCache     map[string]bool       // Cache which remembers which events are triggering
	triggeringCacheLock sync.Mutex            // Lock for triggeringg cache
	messageQueue        *pubsub.EventPump     // Queue for message passing between components
	rmErrorObserver     func(rm *RootMonitor) // Error observer for root monitors
}

/*
NewProcessor creates a new event processor with a given number of workers.
*/
func NewProcessor(workerCount int) Processor {
	ep := pubsub.NewEventPump()
	return &eventProcessor{newProcID(), pool.NewThreadPoolWithQueue(NewTaskQueue(ep)),
		workerCount, false, NewRuleIndex(), nil, sync.Mutex{}, ep, nil}
}

/*
ID returns the processor ID.
*/
func (p *eventProcessor) ID() uint64 {
	return p.id
}

/*
ThreadPool returns the thread pool which this processor is using.
*/
func (p *eventProcessor) ThreadPool() *pool.ThreadPool {
	return p.pool
}

/*
Workers returns the number of threads of this processor.
*/
func (p *eventProcessor) Workers() int {
	return p.workerCount
}

/*
Reset removes all stored rules from this processor.
*/
func (p *eventProcessor) Reset() error {

	// Check that the thread pool is stopped

	if p.pool.Status() != pool.StatusStopped {
		return fmt.Errorf("Cannot reset processor if it has not stopped")
	}

	// Invalidate triggering cache

	p.triggeringCacheLock.Lock()
	p.triggeringCache = nil
	p.triggeringCacheLock.Unlock()

	// Create a new rule index

	p.ruleIndex = NewRuleIndex()

	return nil
}

/*
AddRule adds a new rule to the processor.
*/
func (p *eventProcessor) AddRule(rule *Rule) error {

	// Check that the thread pool is stopped

	if p.pool.Status() != pool.StatusStopped {
		return fmt.Errorf("Cannot add rule if the processor has not stopped")
	}

	// Invalidate triggering cache

	p.triggeringCacheLock.Lock()
	p.triggeringCache = nil
	p.triggeringCacheLock.Unlock()

	return p.ruleIndex.AddRule(rule)
}

/*
Rules returns all loaded rules.
*/
func (p *eventProcessor) Rules() map[string]*Rule {
	return p.ruleIndex.Rules()
}

/*
Start starts this processor.
*/
func (p *eventProcessor) Start() {
	p.pool.SetWorkerCount(p.workerCount, false)
}

/*
Finish will finish all remaining tasks and then stop the processor.
*/
func (p *eventProcessor) Finish() {
	p.pool.JoinAll()
}

/*
Stopped returns if the processor is stopped.
*/
func (p *eventProcessor) Stopped() bool {
	return p.pool.Status() == pool.StatusStopped
}

/*
Status returns the status of the processor (Running / Stopping / Stopped).
*/
func (p *eventProcessor) Status() string {
	return p.pool.Status()
}

/*
NewRootMonitor creates a new root monitor for this processor. This monitor is used to add initial
root events.
*/
func (p *eventProcessor) NewRootMonitor(context map[string]interface{}, scope *RuleScope) *RootMonitor {

	if scope == nil {
		scope = NewRuleScope(map[string]bool{
			"": true, // Default root monitor has global scope
		})
	}

	return newRootMonitor(context, scope, p.messageQueue)
}

/*
SetRootMonitorErrorObserver specifies an observer which is triggered
when a root monitor of this processor has finished and returns errors.
By default this is set to nil (no observer).
*/
func (p *eventProcessor) SetRootMonitorErrorObserver(rmErrorObserver func(rm *RootMonitor)) {
	p.rmErrorObserver = rmErrorObserver
}

/*
SetFailOnFirstErrorInTriggerSequence sets the behavior when rules return errors.
If set to false (default) then all rules in a trigger sequence for a specific event
are executed. If set to true then the first rule which returns an error will stop
the trigger sequence. Events which have been added by the failing rule are still processed.
*/
func (p *eventProcessor) SetFailOnFirstErrorInTriggerSequence(v bool) {
	p.failOnFirstError = v
}

/*
Notify the root monitor error observer that an error occurred.
*/
func (p *eventProcessor) notifyRootMonitorErrors(rm *RootMonitor) {
	if p.rmErrorObserver != nil {
		p.rmErrorObserver(rm)
	}
}

/*
AddEventAndWait adds a new event to the processor and waits for the resulting event cascade
to finish. If a monitor is passed then it must be a RootMonitor.
*/
func (p *eventProcessor) AddEventAndWait(event *Event, monitor *RootMonitor) (Monitor, error) {
	var wg sync.WaitGroup
	wg.Add(1)

	if monitor == nil {
		monitor = p.NewRootMonitor(nil, nil)
	}

	p.messageQueue.AddObserver(MessageRootMonitorFinished, monitor,
		func(event string, eventSource interface{}) {

			// Everything has finished

			wg.Done()

			p.messageQueue.RemoveObservers(event, eventSource)
		})

	resMonitor, err := p.AddEvent(event, monitor)

	if resMonitor == nil {

		// Event was not added

		p.messageQueue.RemoveObservers(MessageRootMonitorFinished, monitor)

	} else {

		// Event was added now wait for it to finish

		wg.Wait()
	}

	return resMonitor, err
}

/*
AddEvent adds a new event to the processor. Returns the monitor if the event
triggered a rule and nil if the event was skipped.
*/
func (p *eventProcessor) AddEvent(event *Event, eventMonitor Monitor) (Monitor, error) {

	// Check that the thread pool is running

	if s := p.pool.Status(); s == pool.StatusStopped || s == pool.StatusStopping {
		return nil, fmt.Errorf("Cannot add event if the processor is stopping or not running")
	}

	EventTracer.record(event, "eventProcessor.AddEvent", "Event added to the processor")

	// First check if the event is triggering any rules at all

	if !p.IsTriggering(event) {

		EventTracer.record(event, "eventProcessor.AddEvent", "Event was skipped")

		if eventMonitor != nil {
			eventMonitor.Skip(event)
		}

		return nil, nil
	}

	// Check if we need to construct a new root monitor

	if eventMonitor == nil {
		eventMonitor = p.NewRootMonitor(nil, nil)
	}

	if rootMonitor, ok := eventMonitor.(*RootMonitor); ok {
		p.messageQueue.AddObserver(MessageRootMonitorFinished, rootMonitor,
			func(event string, eventSource interface{}) {

				// Call finish handler if there is one

				if rm := eventSource.(*RootMonitor); rm.finished != nil {
					rm.finished(p)
				}

				p.messageQueue.RemoveObservers(event, eventSource)
			})
	}

	eventMonitor.Activate(event)

	EventTracer.record(event, "eventProcessor.AddEvent", "Adding task to thread pool")

	// Kick off event processing (see Processor.ProcessEvent)

	p.pool.AddTask(&Task{p, eventMonitor, event})

	return eventMonitor, nil
}

/*
IsTriggering checks if a given event triggers a loaded rule. This does not the
actual state matching for speed.
*/
func (p *eventProcessor) IsTriggering(event *Event) bool {
	var res, ok bool

	p.triggeringCacheLock.Lock()
	defer p.triggeringCacheLock.Unlock()

	// Ensure the triggering cache exists

	if p.triggeringCache == nil {
		p.triggeringCache = make(map[string]bool)
	}

	name := event.Name()

	if res, ok = p.triggeringCache[name]; !ok {
		res = p.ruleIndex.IsTriggering(event)
		p.triggeringCache[name] = res
	}

	return res
}

/*
ProcessEvent processes an event by determining which rules trigger and match
the given event.
*/
func (p *eventProcessor) ProcessEvent(tid uint64, event *Event, parent Monitor) map[string]error {
	var rulesTriggering []*Rule
	var rulesExecuting []*Rule

	scope := parent.Scope()
	ruleCandidates := p.ruleIndex.Match(event)
	suppressedRules := make(map[string]bool)

	EventTracer.record(event, "eventProcessor.ProcessEvent", "Processing event")

	// Remove candidates which are out of scope

	for _, ruleCandidate := range ruleCandidates {

		if scope.IsAllowedAll(ruleCandidate.ScopeMatch) {
			rulesTriggering = append(rulesTriggering, ruleCandidate)

			// Build up a suppression list

			for _, suppressedRule := range ruleCandidate.SuppressionList {
				suppressedRules[suppressedRule] = true
			}
		}
	}

	// Remove suppressed rules

	for _, ruleTriggers := range rulesTriggering {
		if _, ok := suppressedRules[ruleTriggers.Name]; ok {
			continue
		}
		rulesExecuting = append(rulesExecuting, ruleTriggers)
	}

	// Sort rules according to their priority (0 is the highest)

	SortRuleSlice(rulesExecuting)

	// Run rules which are not suppressed

	errors := make(map[string]error)

	EventTracer.record(event, "eventProcessor.ProcessEvent", "Running rules: ", rulesExecuting)

	for _, rule := range rulesExecuting {
		if err := rule.Action(p, parent, event, tid); err != nil {
			errors[rule.Name] = err
		}
		if p.failOnFirstError && len(errors) > 0 {
			break
		}
	}

	return errors
}

/*
String returns a string representation the processor.
*/
func (p *eventProcessor) String() string {
	return fmt.Sprintf("RumbleProcessor %v (workers:%v)", p.ID(), p.workerCount)
}

// Unique id creation
// ==================

var pidcounter uint64 = 1
var pidcounterLock = &sync.Mutex{}

/*
newProcId returns a new unique id or processors.
*/
func newProcID() uint64 {
	pidcounterLock.Lock()
	defer pidcounterLock.Unlock()

	ret := pidcounter
	pidcounter++

	return ret
}
