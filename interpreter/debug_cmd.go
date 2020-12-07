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
	"strconv"
	"strings"

	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/util"
)

/*
DebugCommandsMap contains the mapping of inbuild debug commands.
*/
var DebugCommandsMap = map[string]util.DebugCommand{
	"breakonstart": &breakOnStartCommand{&inbuildDebugCommand{}},
	"break":        &setBreakpointCommand{&inbuildDebugCommand{}},
	"rmbreak":      &rmBreakpointCommand{&inbuildDebugCommand{}},
	"disablebreak": &disableBreakpointCommand{&inbuildDebugCommand{}},
	"cont":         &contCommand{&inbuildDebugCommand{}},
	"describe":     &describeCommand{&inbuildDebugCommand{}},
	"status":       &statusCommand{&inbuildDebugCommand{}},
	"extract":      &extractCommand{&inbuildDebugCommand{}},
	"inject":       &injectCommand{&inbuildDebugCommand{}},
}

/*
inbuildDebugCommand is the base structure for inbuild debug commands providing some
utility functions.
*/
type inbuildDebugCommand struct {
}

/*
AssertNumParam converts a parameter into a number.
*/
func (ibf *inbuildDebugCommand) AssertNumParam(index int, val string) (uint64, error) {
	if resNum, err := strconv.ParseInt(fmt.Sprint(val), 10, 0); err == nil {
		return uint64(resNum), nil
	}
	return 0, fmt.Errorf("Parameter %v should be a number", index)
}

// break
// =====

/*
setBreakpointCommand sets a breakpoint
*/
type setBreakpointCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *setBreakpointCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Need a break target (<source>:<line>) as first parameter")
	}

	targetSplit := strings.Split(args[0], ":")

	if len(targetSplit) > 1 {
		if line, err := strconv.Atoi(targetSplit[1]); err == nil {

			debugger.SetBreakPoint(targetSplit[0], line)

			return nil, nil
		}
	}

	return nil, fmt.Errorf("Invalid break target - should be <source>:<line>")
}

/*
DocString returns a descriptive text about this command.
*/
func (c *setBreakpointCommand) DocString() string {
	return "Set a breakpoint specifying <source>:<line>"
}

// breakOnStartCommand
// ===================

/*
breakOnStartCommand breaks on the start of the next execution.
*/
type breakOnStartCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *breakOnStartCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	b := true
	if len(args) > 0 {
		b, _ = strconv.ParseBool(args[0])
	}
	debugger.BreakOnStart(b)
	return nil, nil
}

/*
DocString returns a descriptive text about this command.
*/
func (c *breakOnStartCommand) DocString() string {
	return "Break on the start of the next execution."
}

// rmbreak
// =======

/*
rmBreakpointCommand removes a breakpoint
*/
type rmBreakpointCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *rmBreakpointCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Need a break target (<source>[:<line>]) as first parameter")
	}

	targetSplit := strings.Split(args[0], ":")

	if len(targetSplit) > 1 {

		if line, err := strconv.Atoi(targetSplit[1]); err == nil {

			debugger.RemoveBreakPoint(targetSplit[0], line)

			return nil, nil
		}

	} else {

		debugger.RemoveBreakPoint(args[0], -1)
	}

	return nil, nil
}

/*
DocString returns a descriptive text about this command.
*/
func (c *rmBreakpointCommand) DocString() string {
	return "Remove a breakpoint specifying <source>:<line>"
}

// disablebreak
// ============

/*
disableBreakpointCommand temporarily disables a breakpoint
*/
type disableBreakpointCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *disableBreakpointCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Need a break target (<source>:<line>) as first parameter")
	}

	targetSplit := strings.Split(args[0], ":")

	if len(targetSplit) > 1 {

		if line, err := strconv.Atoi(targetSplit[1]); err == nil {

			debugger.DisableBreakPoint(targetSplit[0], line)

			return nil, nil
		}
	}

	return nil, fmt.Errorf("Invalid break target - should be <source>:<line>")
}

/*
DocString returns a descriptive text about this command.
*/
func (c *disableBreakpointCommand) DocString() string {
	return "Temporarily disable a breakpoint specifying <source>:<line>"
}

// cont
// ====

/*
contCommand continues a suspended thread
*/
type contCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *contCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	var cmd util.ContType

	if len(args) != 2 {
		return nil, fmt.Errorf("Need a thread ID and a command Resume, StepIn, StepOver or StepOut")
	}

	threadID, err := c.AssertNumParam(1, args[0])

	if err == nil {
		cmdString := strings.ToLower(args[1])
		switch cmdString {
		case "resume":
			cmd = util.Resume
		case "stepin":
			cmd = util.StepIn
		case "stepover":
			cmd = util.StepOver
		case "stepout":
			cmd = util.StepOut
		default:
			return nil, fmt.Errorf("Invalid command %v - must be resume, stepin, stepover or stepout", cmdString)
		}

		debugger.Continue(threadID, cmd)
	}

	return nil, err
}

/*
DocString returns a descriptive text about this command.
*/
func (c *contCommand) DocString() string {
	return "Continues a suspended thread. Specify <threadID> <Resume | StepIn | StepOver | StepOut>"
}

// describe
// ========

/*
describeCommand describes a suspended thread
*/
type describeCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *describeCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	var res interface{}

	if len(args) != 1 {
		return nil, fmt.Errorf("Need a thread ID")
	}

	threadID, err := c.AssertNumParam(1, args[0])

	if err == nil {

		res = debugger.Describe(threadID)
	}

	return res, err
}

/*
DocString returns a descriptive text about this command.
*/
func (c *describeCommand) DocString() string {
	return "Describes a suspended thread."
}

// status
// ======

/*
statusCommand shows breakpoints and suspended threads
*/
type statusCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *statusCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	return debugger.Status(), nil
}

/*
DocString returns a descriptive text about this command.
*/
func (c *statusCommand) DocString() string {
	return "Shows breakpoints and suspended threads."
}

// extract
// =======

/*
extractCommand copies a value from a suspended thread into the
global variable scope
*/
type extractCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *extractCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("Need a thread ID, a variable name and a destination variable name")
	}

	threadID, err := c.AssertNumParam(1, args[0])

	if err == nil {
		if !parser.NamePattern.MatchString(args[1]) || !parser.NamePattern.MatchString(args[2]) {
			err = fmt.Errorf("Variable names may only contain [a-zA-Z] and [a-zA-Z0-9] from the second character")
		}

		if err == nil {
			err = debugger.ExtractValue(threadID, args[1], args[2])
		}
	}

	return nil, err
}

/*
DocString returns a descriptive text about this command.
*/
func (c *extractCommand) DocString() string {
	return "Copies a value from a suspended thread into the global variable scope."
}

// inject
// =======

/*
injectCommand copies a value from the global variable scope into
a suspended thread
*/
type injectCommand struct {
	*inbuildDebugCommand
}

/*
Execute the debug command and return its result. It must be possible to
convert the output data into a JSON string.
*/
func (c *injectCommand) Run(debugger util.ECALDebugger, args []string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("Need a thread ID, a variable name and an expression")
	}

	threadID, err := c.AssertNumParam(1, args[0])

	if err == nil {
		varName := args[1]
		expression := strings.Join(args[2:], " ")

		err = debugger.InjectValue(threadID, varName, expression)
	}

	return nil, err
}

/*
DocString returns a descriptive text about this command.
*/
func (c *injectCommand) DocString() string {
	return "Copies a value from the global variable scope into a suspended thread."
}
