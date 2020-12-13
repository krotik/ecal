/*
 * Public Domain Software
 *
 * I (Matthias Ladkau) am the author of the source code in this file.
 * I have placed the source code in this file in the public domain.
 *
 * For further information see: http://creativecommons.org/publicdomain/zero/1.0/
 */

/*
Example ECAL stdlib function plugin.

The plugins must be valid Go plugins: https://golang.org/pkg/plugin

ECAL usable functions imported via a plugin must conform to the following interface:

type ECALPluginFunction interface {
	Run(args []interface{}) (interface{}, error) // Function execution with given arguments
	DocString() string // Returns some function description
}
*/

package main

import "fmt"

func init() {

	// Here goes some initialisation code

	Greeting = "Hello"
	ECALmyfunc = myfunc{"World"}
}

/*
Greeting is first word in the output
*/
var Greeting string

type myfunc struct {
	place string
}

func (f *myfunc) Run(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("Need a name to greet as argument")
	}

	return fmt.Sprintf("%v %v for %v", Greeting, f.place, args[0]), nil
}

func (f *myfunc) DocString() string {
	return "Myfunc is an example function"
}

// Exported bits

/*
ECALmyfunc is the exported function which can be used by ECAL
*/
var ECALmyfunc myfunc
