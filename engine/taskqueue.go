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
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/sortutil"
	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/ecal/engine/pool"
	"devt.de/krotik/ecal/engine/pubsub"
)

/*
TaskError datastructure to collect all rule errors of an event.
*/
type TaskError struct {
	ErrorMap map[string]error // Rule errors (rule name -> error)
	Event    *Event           // Event which caused the error
	Monitor  Monitor          // Event monitor
}

/*
Error returns a string representation of this error.
*/
func (te *TaskError) Error() string {
	var ret bytes.Buffer

	// Collect all errors and sort them by name

	errNames := make([]string, 0, len(te.ErrorMap))

	for name := range te.ErrorMap {
		errNames = append(errNames, name)
	}

	sort.Strings(errNames)

	ret.WriteString(fmt.Sprintf("Taskerror%v:\n", stringutil.Plural(len(errNames))))
	for i, name := range errNames {
		ret.WriteString(te.Monitor.EventPathString())
		ret.WriteString(fmt.Sprintf(" -> %v : %v", name, te.ErrorMap[name]))
		if i < len(errNames)-1 {
			ret.WriteString("\n")
		}
	}

	return ret.String()
}

/*
Task models a task which is created and executed by the processor.
*/
type Task struct {
	p Processor // Processor which created the task
	m Monitor   // Monitor which observes the task execution
	e *Event    // Event which caused the task creation
}

/*
Run the task.
*/
func (t *Task) Run(tid uint64) error {
	EventTracer.record(t.e, "Task.Run", "Running task")

	errors := t.p.ProcessEvent(tid, t.e, t.m)

	if len(errors) > 0 {

		// Monitor is not declared finished until the errors have been handled

		EventTracer.record(t.e, "Task.Run", fmt.Sprint("Task had errors:", errors))
		return &TaskError{errors, t.e, t.m}
	}

	t.m.Finish()

	return nil
}

/*
Returns a string representation of this task.
*/
func (t *Task) String() string {
	return fmt.Sprintf("Task: %v %v %v", t.p, t.m, t.e)
}

/*
HandleError handles an error which occurred during the run method.
*/
func (t *Task) HandleError(e error) {
	t.m.SetErrors(e.(*TaskError))
	t.m.Finish()
	t.p.(*eventProcessor).notifyRootMonitorErrors(t.m.RootMonitor())
}

/*
TaskQueue models the queue of tasks for a processor.
*/
type TaskQueue struct {
	lock         *sync.Mutex                        // Lock for queue
	queues       map[uint64]*sortutil.PriorityQueue // Map from root monitor id to priority queue
	messageQueue *pubsub.EventPump                  // Queue for message passing between components
}

/*
NewTaskQueue creates a new TaskQueue object.
*/
func NewTaskQueue(ep *pubsub.EventPump) *TaskQueue {
	return &TaskQueue{&sync.Mutex{}, make(map[uint64]*sortutil.PriorityQueue), ep}
}

/*
Clear the queue of all pending tasks.
*/
func (tq *TaskQueue) Clear() {
	tq.lock.Lock()
	defer tq.lock.Unlock()

	tq.queues = make(map[uint64]*sortutil.PriorityQueue)
}

/*
Pop returns the next task from the queue.
*/
func (tq *TaskQueue) Pop() pool.Task {
	tq.lock.Lock()
	defer tq.lock.Unlock()

	var popQueue *sortutil.PriorityQueue
	var idx int

	// Pick a random number between 0 and len(tq.queues) - 1

	if lq := len(tq.queues); lq > 0 {
		idx = rand.Intn(lq)
	}

	// Go through all queues and pick one - clean up while we are at it

	for k, v := range tq.queues {

		if v.Size() > 0 {

			// Pick a random queue - pick the last if idx does not
			// reach 0 before the end of the iteration.

			idx--

			popQueue = v

			if idx <= 0 {
				break
			}

		} else {

			// Remove empty queues

			delete(tq.queues, k)
		}
	}

	if popQueue != nil {
		if res := popQueue.Pop(); res != nil {
			return res.(*Task)
		}
	}

	return nil
}

/*
Push adds another task to the queue.
*/
func (tq *TaskQueue) Push(t pool.Task) {
	tq.lock.Lock()
	defer tq.lock.Unlock()

	var q *sortutil.PriorityQueue
	var ok bool

	task := t.(*Task)

	rm := task.m.RootMonitor()
	id := rm.ID()

	if q, ok = tq.queues[id]; !ok {
		q = sortutil.NewPriorityQueue()
		tq.queues[id] = q

		// Add listener for finish

		tq.messageQueue.AddObserver(MessageRootMonitorFinished, rm,
			func(event string, eventSource interface{}) {
				tq.lock.Lock()
				defer tq.lock.Unlock()

				rm := eventSource.(*RootMonitor)
				q := tq.queues[rm.ID()]

				// Safeguard that no tasks are ever left over

				errorutil.AssertTrue(q == nil || q.Size() == 0,
					"Finished monitor left events behind")

				tq.messageQueue.RemoveObservers(event, eventSource)
			})
	}

	q.Push(task, task.m.Priority())
}

/*
Size returns the size of the queue.
*/
func (tq *TaskQueue) Size() int {
	tq.lock.Lock()
	defer tq.lock.Unlock()

	var ret int

	for _, q := range tq.queues {
		ret += q.Size()
	}

	return ret
}
