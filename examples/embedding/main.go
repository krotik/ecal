/*
 * ECAL Embedding Example
 */

package main

import (
	"fmt"
	"log"

	"devt.de/krotik/ecal/engine"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/stdlib"
	"devt.de/krotik/ecal/util"
)

func main() {

	// The code to execute

	code := `
sink mysink
    kindmatch [ "foo.*" ],
    {
        log("Handling: ", event)
        log("Result: ", event.state.op1 + event.state.op2)
    }

sink mysink2
    kindmatch [ "foo.*" ],
    {
        raise("Some error")
    }

func compute(x) {
	let result := x + 1
	return result
}

mystuff.add(compute(5), 1)
`

	// Add a stdlib function

	stdlib.AddStdlibPkg("mystuff", "My special functions")

	// A single instance if the ECALFunction struct will be used for all function calls across all threads

	stdlib.AddStdlibFunc("mystuff", "add", &AddFunc{})

	// Logger for log() statements in the code

	logger := util.NewMemoryLogger(100)

	// Import locator when using import statements in the code

	importLocator := &util.MemoryImportLocator{Files: make(map[string]string)}

	// Runtime provider which contains all objects needed by the interpreter

	rtp := interpreter.NewECALRuntimeProvider("Embedded Example", importLocator, logger)

	// First we need to parse the code into an Abstract Syntax Tree

	ast, err := parser.ParseWithRuntime("code1", code, rtp)
	if err != nil {
		log.Fatal(err)
	}

	// Then we need to validate the code - this prepares certain runtime bits
	// of the AST for execution.

	if err = ast.Runtime.Validate(); err != nil {
		log.Fatal(err)
	}

	// We need a global variable scope which contains all declared variables - use
	// this object to inject initialization values into the ECAL program.

	vs := scope.NewScope(scope.GlobalScope)

	// Each thread which evaluates the Runtime of an AST should get a unique thread ID

	var threadId uint64 = 1

	// Evaluate the Runtime of an AST with a variable scope

	res, err := ast.Runtime.Eval(vs, make(map[string]interface{}), threadId)
	if err != nil {
		log.Fatal(err)
	}

	// The executed code returns the value of the last statement

	fmt.Println("Computation result:", res)

	// We can also react to events

	rtp.Processor.Start()
	monitor, err := rtp.Processor.AddEventAndWait(engine.NewEvent("MyEvent", []string{"foo", "bar"}, map[interface{}]interface{}{
		"op1": float64(5.2),
		"op2": float64(5.3),
	}), nil)

	if err != nil {
		log.Fatal(err)
	}

	// All errors can be found on the returned monitor object

	fmt.Println("Event result:", monitor.RootMonitor().AllErrors())

	// The log messages of a program can be collected

	fmt.Println("Log:", logger.String())
}

/*
AddFunc is a simple add function which calculates the sum of two numbers.
*/
type AddFunc struct {
}

func (f *AddFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, tid uint64, args []interface{}) (interface{}, error) {

	// This should have some proper error checking

	// Arguments are either of type string, float64, map[interface{}]interface{}
	// or []interface{}

	return args[0].(float64) + args[1].(float64), nil
}

func (f *AddFunc) DocString() (string, error) {
	return "Sum up two numbers", nil
}
