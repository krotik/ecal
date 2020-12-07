/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

/*
Package interpreter contains the ECAL interpreter.
*/
package interpreter

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"devt.de/krotik/common/datautil"
	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

/*
ecalDebugger is the inbuild default debugger.
*/
type ecalDebugger struct {
	breakPoints                map[string]bool                     // Break points (active or not)
	interrogationStates        map[uint64]*interrogationState      // Collection of threads which are interrogated
	callStacks                 map[uint64][]*parser.ASTNode        // Call stack locations of threads
	callStackVsSnapshots       map[uint64][]map[string]interface{} // Call stack variable scope snapshots of threads
	callStackGlobalVsSnapshots map[uint64][]map[string]interface{} // Call stack global variable scope snapshots of threads
	sources                    map[string]bool                     // All known sources
	breakOnStart               bool                                // Flag to stop at the start of the next execution
	breakOnError               bool                                // Flag to stop if an error occurs
	globalScope                parser.Scope                        // Global variable scope which can be used to transfer data
	lock                       *sync.RWMutex                       // Lock for this debugger
	lastVisit                  int64                               // Last time the debugger had a state visit
}

/*
interrogationState contains state information of a thread interrogation.
*/
type interrogationState struct {
	cond         *sync.Cond        // Condition on which the thread is waiting when suspended
	running      bool              // Flag if the thread is running or waiting
	cmd          interrogationCmd  // Next interrogation command for the thread
	stepOutStack []*parser.ASTNode // Target stack when doing a step out
	node         *parser.ASTNode   // Node on which the thread was last stopped
	vs           parser.Scope      // Variable scope of the thread when it was last stopped
	err          error             // Error which was returned by a function call
}

/*
interrogationCmd represents a command for a thread interrogation.
*/
type interrogationCmd int

/*
Interrogation commands
*/
const (
	Stop     interrogationCmd = iota // Stop the execution (default)
	StepIn                           // Step into the next function
	StepOut                          // Step out of the current function
	StepOver                         // Step over the next function
	Resume                           // Resume execution - do not break again on the same line
	Kill                             // Resume execution - and kill the thread on the next state change
)

/*
newInterrogationState creates a new interrogation state.
*/
func newInterrogationState(node *parser.ASTNode, vs parser.Scope) *interrogationState {
	return &interrogationState{
		sync.NewCond(&sync.Mutex{}),
		false,
		Stop,
		nil,
		node,
		vs,
		nil,
	}
}

/*
NewECALDebugger returns a new debugger object.
*/
func NewECALDebugger(globalVS parser.Scope) util.ECALDebugger {
	return &ecalDebugger{
		breakPoints:                make(map[string]bool),
		interrogationStates:        make(map[uint64]*interrogationState),
		callStacks:                 make(map[uint64][]*parser.ASTNode),
		callStackVsSnapshots:       make(map[uint64][]map[string]interface{}),
		callStackGlobalVsSnapshots: make(map[uint64][]map[string]interface{}),
		sources:                    make(map[string]bool),
		breakOnStart:               false,
		breakOnError:               true,
		globalScope:                globalVS,
		lock:                       &sync.RWMutex{},
		lastVisit:                  0,
	}
}

/*
HandleInput handles a given debug instruction from a console.
*/
func (ed *ecalDebugger) HandleInput(input string) (interface{}, error) {
	var res interface{}
	var err error

	args := strings.Fields(input)

	if len(args) > 0 {
		if cmd, ok := DebugCommandsMap[args[0]]; ok {
			if len(args) > 1 {
				res, err = cmd.Run(ed, args[1:])
			} else {
				res, err = cmd.Run(ed, nil)
			}
		} else {
			err = fmt.Errorf("Unknown command: %v", args[0])
		}
	}

	return res, err
}

/*
StopThreads will continue all suspended threads and set them to be killed.
Returns true if a waiting thread was resumed. Can wait for threads to end
by ensuring that for at least d time no state change occurred.
*/
func (ed *ecalDebugger) StopThreads(d time.Duration) bool {
	var ret = false

	for _, is := range ed.interrogationStates {
		if is.running == false {
			ret = true
			is.cmd = Kill
			is.running = true
			is.cond.L.Lock()
			is.cond.Broadcast()
			is.cond.L.Unlock()
		}
	}

	if ret && d > 0 {
		var lastVisit int64 = -1
		for lastVisit != ed.lastVisit {
			lastVisit = ed.lastVisit
			time.Sleep(d)
		}
	}

	return ret
}

/*
BreakOnStart breaks on the start of the next execution.
*/
func (ed *ecalDebugger) BreakOnStart(flag bool) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakOnStart = flag
}

/*
BreakOnError breaks if an error occurs.
*/
func (ed *ecalDebugger) BreakOnError(flag bool) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakOnError = flag
}

/*
VisitState is called for every state during the execution of a program.
*/
func (ed *ecalDebugger) VisitState(node *parser.ASTNode, vs parser.Scope, tid uint64) util.TraceableRuntimeError {

	ed.lock.RLock()
	_, ok := ed.callStacks[tid]
	ed.lastVisit = time.Now().UnixNano()
	ed.lock.RUnlock()

	if !ok {

		// Make the debugger aware of running threads

		ed.lock.Lock()
		ed.callStacks[tid] = make([]*parser.ASTNode, 0, 10)
		ed.callStackVsSnapshots[tid] = make([]map[string]interface{}, 0, 10)
		ed.callStackGlobalVsSnapshots[tid] = make([]map[string]interface{}, 0, 10)
		ed.lock.Unlock()
	}

	if node.Token != nil { // Statements are excluded here
		targetIdentifier := fmt.Sprintf("%v:%v", node.Token.Lsource, node.Token.Lline)

		ed.lock.RLock()
		is, ok := ed.interrogationStates[tid]
		_, sourceKnown := ed.sources[node.Token.Lsource]
		ed.lock.RUnlock()

		if !sourceKnown {
			ed.RecordSource(node.Token.Lsource)
		}

		if ok {

			// The thread is being interrogated

			switch is.cmd {
			case Resume, Kill:
				if is.node.Token.Lline != node.Token.Lline {

					// Remove the resume command once we are on a different line

					ed.lock.Lock()
					delete(ed.interrogationStates, tid)
					ed.lock.Unlock()

					if is.cmd == Kill {
						runtime.Goexit()
					}

					return ed.VisitState(node, vs, tid)
				}
			case Stop, StepIn, StepOver:

				if is.node.Token.Lline != node.Token.Lline || is.cmd == Stop {
					is.node = node
					is.vs = vs
					is.running = false

					is.cond.L.Lock()
					is.cond.Wait()
					is.cond.L.Unlock()
				}
			}

		} else if active, ok := ed.breakPoints[targetIdentifier]; (ok && active) || ed.breakOnStart {

			// A globally defined breakpoint has been hit - note the position
			// in the thread specific map and wait

			is := newInterrogationState(node, vs)

			ed.lock.Lock()
			ed.breakOnStart = false
			ed.interrogationStates[tid] = is
			ed.lock.Unlock()

			is.cond.L.Lock()
			is.cond.Wait()
			is.cond.L.Unlock()
		}
	}

	return nil
}

/*
VisitStepInState is called before entering a function call.
*/
func (ed *ecalDebugger) VisitStepInState(node *parser.ASTNode, vs parser.Scope, tid uint64) util.TraceableRuntimeError {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	var err util.TraceableRuntimeError

	threadCallStack := ed.callStacks[tid]
	threadCallStackVs := ed.callStackVsSnapshots[tid]
	threadCallStackGlobalVs := ed.callStackGlobalVsSnapshots[tid]

	is, ok := ed.interrogationStates[tid]

	if ok {

		if is.cmd == Stop {

			// Special case a parameter of a function was resolved by another
			// function call - the debugger should stop before entering

			ed.lock.Unlock()
			err = ed.VisitState(node, vs, tid)
			ed.lock.Lock()
		}

		if err == nil {
			// The thread is being interrogated

			switch is.cmd {
			case StepIn:
				is.cmd = Stop
			case StepOver:
				is.cmd = StepOut
				is.stepOutStack = threadCallStack
			}
		}
	}

	ed.callStacks[tid] = append(threadCallStack, node)
	ed.callStackVsSnapshots[tid] = append(threadCallStackVs, ed.buildVsSnapshot(vs))
	ed.callStackGlobalVsSnapshots[tid] = append(threadCallStackGlobalVs, ed.buildGlobalVsSnapshot(vs))

	return err
}

/*
VisitStepOutState is called after returning from a function call.
*/
func (ed *ecalDebugger) VisitStepOutState(node *parser.ASTNode, vs parser.Scope, tid uint64, soErr error) util.TraceableRuntimeError {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	threadCallStack := ed.callStacks[tid]
	threadCallStackVs := ed.callStackVsSnapshots[tid]
	threadCallStackGlobalVs := ed.callStackGlobalVsSnapshots[tid]
	lastIndex := len(threadCallStack) - 1

	ok, cerr := threadCallStack[lastIndex].Equals(node, false) // Sanity check step in node must be the same as step out node
	errorutil.AssertTrue(ok,
		fmt.Sprintf("Unexpected callstack when stepping out - callstack: %v - funccall: %v - comparison error: %v",
			threadCallStack, node, cerr))

	ed.callStacks[tid] = threadCallStack[:lastIndex] // Remove the last item
	ed.callStackVsSnapshots[tid] = threadCallStackVs[:lastIndex]
	ed.callStackGlobalVsSnapshots[tid] = threadCallStackGlobalVs[:lastIndex]

	is, ok := ed.interrogationStates[tid]

	if ed.breakOnError && soErr != nil {

		if !ok {
			is = newInterrogationState(node, vs)

			ed.breakOnStart = false
			ed.interrogationStates[tid] = is

		} else {
			is.node = node
			is.vs = vs
			is.running = false
		}

		if is.err == nil {

			// Only stop if the error is being set

			is.err = soErr

			ed.lock.Unlock()
			is.cond.L.Lock()
			is.cond.Wait()
			is.cond.L.Unlock()
			ed.lock.Lock()
		}

	} else if ok {

		is.err = soErr

		// The thread is being interrogated

		switch is.cmd {
		case StepOver, StepOut:

			if len(ed.callStacks[tid]) == len(is.stepOutStack) {
				is.cmd = Stop
			}
		}
	}

	return nil
}

/*
RecordSource records a code source.
*/
func (ed *ecalDebugger) RecordSource(source string) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.sources[source] = true
}

/*
RecordThreadFinished lets the debugger know that a thread has finished.
*/
func (ed *ecalDebugger) RecordThreadFinished(tid uint64) {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	if is, ok := ed.interrogationStates[tid]; !ok || !is.running {
		delete(ed.interrogationStates, tid)
		delete(ed.callStacks, tid)
		delete(ed.callStackVsSnapshots, tid)
		delete(ed.callStackGlobalVsSnapshots, tid)
	}
}

/*
SetBreakPoint sets a break point.
*/
func (ed *ecalDebugger) SetBreakPoint(source string, line int) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakPoints[fmt.Sprintf("%v:%v", source, line)] = true
}

/*
DisableBreakPoint disables a break point but keeps the code reference.
*/
func (ed *ecalDebugger) DisableBreakPoint(source string, line int) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakPoints[fmt.Sprintf("%v:%v", source, line)] = false
}

/*
RemoveBreakPoint removes a break point.
*/
func (ed *ecalDebugger) RemoveBreakPoint(source string, line int) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	if line > 0 {
		delete(ed.breakPoints, fmt.Sprintf("%v:%v", source, line))
	} else {
		for k := range ed.breakPoints {
			if ksource := strings.Split(k, ":")[0]; ksource == source {
				delete(ed.breakPoints, k)
			}
		}
	}
}

/*
ExtractValue copies a value from a suspended thread into the
global variable scope.
*/
func (ed *ecalDebugger) ExtractValue(threadID uint64, varName string, destVarName string) error {
	if ed.globalScope == nil {
		return fmt.Errorf("Cannot access global scope")
	}

	err := fmt.Errorf("Cannot find suspended thread %v", threadID)

	ed.lock.Lock()
	defer ed.lock.Unlock()

	is, ok := ed.interrogationStates[threadID]

	if ok && !is.running {
		var val interface{}
		var ok bool

		if val, ok, err = is.vs.GetValue(varName); ok {
			err = ed.globalScope.SetValue(destVarName, val)
		} else if err == nil {
			err = fmt.Errorf("No such value %v", varName)
		}
	}

	return err
}

/*
InjectValue copies a value from an expression (using the global variable scope) into
a suspended thread.
*/
func (ed *ecalDebugger) InjectValue(threadID uint64, varName string, expression string) error {
	if ed.globalScope == nil {
		return fmt.Errorf("Cannot access global scope")
	}

	err := fmt.Errorf("Cannot find suspended thread %v", threadID)

	ed.lock.Lock()
	defer ed.lock.Unlock()

	is, ok := ed.interrogationStates[threadID]

	if ok && !is.running {
		var ast *parser.ASTNode
		var val interface{}

		// Eval expression

		ast, err = parser.ParseWithRuntime("InjectValueExpression", expression,
			NewECALRuntimeProvider("InjectValueExpression2", nil, nil))

		if err == nil {
			if err = ast.Runtime.Validate(); err == nil {

				ivs := scope.NewScopeWithParent("InjectValueExpressionScope", ed.globalScope)
				val, err = ast.Runtime.Eval(ivs, make(map[string]interface{}), 999)

				if err == nil {
					err = is.vs.SetValue(varName, val)
				}
			}
		}
	}

	return err
}

/*
Continue will continue a suspended thread.
*/
func (ed *ecalDebugger) Continue(threadID uint64, contType util.ContType) {
	ed.lock.RLock()
	defer ed.lock.RUnlock()

	if is, ok := ed.interrogationStates[threadID]; ok && !is.running {

		switch contType {
		case util.Resume:
			is.cmd = Resume
		case util.StepIn:
			is.cmd = StepIn
		case util.StepOver:
			is.cmd = StepOver
		case util.StepOut:
			is.cmd = StepOut
			stack := ed.callStacks[threadID]
			is.stepOutStack = stack[:len(stack)-1]
		}

		is.running = true

		is.cond.L.Lock()
		is.cond.Broadcast()
		is.cond.L.Unlock()
	}
}

/*
Status returns the current status of the debugger.
*/
func (ed *ecalDebugger) Status() interface{} {
	ed.lock.RLock()
	defer ed.lock.RUnlock()

	var sources []string

	threadStates := make(map[string]map[string]interface{})

	res := map[string]interface{}{
		"breakpoints":  ed.breakPoints,
		"breakonstart": ed.breakOnStart,
		"threads":      threadStates,
	}

	for k := range ed.sources {
		sources = append(sources, k)
	}
	res["sources"] = sources

	for k, v := range ed.callStacks {
		s := map[string]interface{}{
			"callStack": ed.prettyPrintCallStack(v),
		}

		if is, ok := ed.interrogationStates[k]; ok {
			s["threadRunning"] = is.running
			s["error"] = is.err
		}

		threadStates[fmt.Sprint(k)] = s
	}

	return res
}

/*
Describe describes a thread currently observed by the debugger.
*/
func (ed *ecalDebugger) Describe(threadID uint64) interface{} {
	ed.lock.RLock()
	defer ed.lock.RUnlock()

	var res map[string]interface{}

	threadCallStack, ok1 := ed.callStacks[threadID]

	if is, ok2 := ed.interrogationStates[threadID]; ok1 && ok2 {
		callStackNode := make([]map[string]interface{}, 0)

		for _, sn := range threadCallStack {
			callStackNode = append(callStackNode, sn.ToJSONObject())
		}

		res = map[string]interface{}{
			"threadRunning":             is.running,
			"error":                     is.err,
			"callStack":                 ed.prettyPrintCallStack(threadCallStack),
			"callStackNode":             callStackNode,
			"callStackVsSnapshot":       ed.callStackVsSnapshots[threadID],
			"callStackVsSnapshotGlobal": ed.callStackGlobalVsSnapshots[threadID],
		}

		if !is.running {

			codeString, _ := parser.PrettyPrint(is.node)
			res["code"] = codeString
			res["node"] = is.node.ToJSONObject()
			res["vs"] = ed.buildVsSnapshot(is.vs)
			res["vsGlobal"] = ed.buildGlobalVsSnapshot(is.vs)
		}
	}

	return res
}

func (ed *ecalDebugger) buildVsSnapshot(vs parser.Scope) map[string]interface{} {
	vsValues := make(map[string]interface{})

	// Collect all parent scopes except the global scope

	parent := vs.Parent()
	for parent != nil &&
		parent.Name() != scope.GlobalScope {

		vsValues = datautil.MergeMaps(vsValues, parent.ToJSONObject())

		parent = parent.Parent()
	}

	return ed.MergeMaps(vsValues, vs.ToJSONObject())
}

func (ed *ecalDebugger) buildGlobalVsSnapshot(vs parser.Scope) map[string]interface{} {
	vsValues := make(map[string]interface{})

	globalVs := vs
	for globalVs != nil &&
		globalVs.Name() != scope.GlobalScope {
		globalVs = globalVs.Parent()
	}

	if globalVs != nil && globalVs.Name() == scope.GlobalScope {
		vsValues = globalVs.ToJSONObject()
	}

	return vsValues
}

/*
MergeMaps merges all given maps into a new map. Contents are shallow copies
and conflicts are resolved as first-one-wins.
*/
func (ed *ecalDebugger) MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})

	for _, m := range maps {
		for k, v := range m {
			if _, ok := ret[k]; !ok {
				ret[k] = v
			}
		}
	}

	return ret
}

/*
Describe describes a thread currently observed by the debugger.
*/
func (ed *ecalDebugger) prettyPrintCallStack(threadCallStack []*parser.ASTNode) []string {
	cs := []string{}
	for _, s := range threadCallStack {
		pp, _ := parser.PrettyPrint(s)
		cs = append(cs, fmt.Sprintf("%v (%v:%v)",
			pp, s.Token.Lsource, s.Token.Lline))
	}
	return cs
}
