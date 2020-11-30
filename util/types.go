/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package util

import (
	"time"

	"devt.de/krotik/ecal/parser"
)

/*
Processor models a top level execution instance for ECAL.
*/
type Processor interface {
}

/*
ECALImportLocator is used to resolve imports.
*/
type ECALImportLocator interface {

	/*
		Resolve a given import path and parse the imported file into an AST.
	*/
	Resolve(path string) (string, error)
}

/*
ECALFunction models a callable function in ECAL.
*/
type ECALFunction interface {

	/*
		Run executes this function. The envirnment provides a unique instanceID for
		every code location in the running code, the variable scope of the function,
		an instance state which can be used in combinartion with the instanceID
		to store instance specific state (e.g. for iterator functions) and a list
		of argument values which were passed to the function by the calling code.
	*/
	Run(instanceID string, vs parser.Scope, is map[string]interface{}, tid uint64, args []interface{}) (interface{}, error)

	/*
	   DocString returns a descriptive text about this function.
	*/
	DocString() (string, error)
}

/*
Logger is required external object to which the interpreter releases its log messages.
*/
type Logger interface {

	/*
	   LogError adds a new error log message.
	*/
	LogError(v ...interface{})

	/*
	   LogInfo adds a new info log message.
	*/
	LogInfo(v ...interface{})

	/*
	   LogDebug adds a new debug log message.
	*/
	LogDebug(v ...interface{})
}

/*
ContType represents a way how to resume code execution of a suspended thread.
*/
type ContType int

/*
Available lexer token types
*/
const (
	Resume   ContType = iota // Resume code execution until the next breakpoint or the end
	StepIn                   // Step into a function call or over the next non-function call
	StepOver                 // Step over the current statement onto the next line
	StepOut                  // Step out of the current function call
)

/*
ECALDebugger is a debugging object which can be used to inspect and modify a running
ECAL environment.
*/
type ECALDebugger interface {

	/*
		HandleInput handles a given debug instruction. It must be possible to
		convert the output data into a JSON string.
	*/
	HandleInput(input string) (interface{}, error)

	/*
	   StopThreads will continue all suspended threads and set them to be killed.
	   Returns true if a waiting thread was resumed. Can wait for threads to end
	   by ensuring that for at least d time no state change occured.
	*/
	StopThreads(d time.Duration) bool

	/*
	   BreakOnStart breaks on the start of the next execution.
	*/
	BreakOnStart(flag bool)

	/*
	   BreakOnError breaks if an error occurs.
	*/
	BreakOnError(flag bool)

	/*
	   VisitState is called for every state during the execution of a program.
	*/
	VisitState(node *parser.ASTNode, vs parser.Scope, tid uint64) TraceableRuntimeError

	/*
	   VisitStepInState is called before entering a function call.
	*/
	VisitStepInState(node *parser.ASTNode, vs parser.Scope, tid uint64) TraceableRuntimeError

	/*
	   VisitStepOutState is called after returning from a function call.
	*/
	VisitStepOutState(node *parser.ASTNode, vs parser.Scope, tid uint64, soErr error) TraceableRuntimeError

	/*
	   RecordThreadFinished lets the debugger know that a thread has finished.
	*/
	RecordThreadFinished(tid uint64)

	/*
	   SetBreakPoint sets a break point.
	*/
	SetBreakPoint(source string, line int)

	/*
	   DisableBreakPoint disables a break point but keeps the code reference.
	*/
	DisableBreakPoint(source string, line int)

	/*
	   RemoveBreakPoint removes a break point.
	*/
	RemoveBreakPoint(source string, line int)

	/*
		ExtractValue copies a value from a suspended thread into the
		global variable scope.
	*/
	ExtractValue(threadId uint64, varName string, destVarName string) error

	/*
		InjectValue copies a value from an expression (using the global
		variable scope) into a suspended thread.
	*/
	InjectValue(threadId uint64, varName string, expression string) error

	/*
	   Continue will continue a suspended thread.
	*/
	Continue(threadId uint64, contType ContType)

	/*
		Status returns the current status of the debugger.
	*/
	Status() interface{}

	/*
	   Describe decribes a thread currently observed by the debugger.
	*/
	Describe(threadId uint64) interface{}
}

/*
DebugCommand is command which can modify and interrogate the debugger.
*/
type DebugCommand interface {

	/*
		Execute the debug command and return its result. It must be possible to
		convert the output data into a JSON string.
	*/
	Run(debugger ECALDebugger, args []string) (interface{}, error)

	/*
	   DocString returns a descriptive text about this command.
	*/
	DocString() string
}
