/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package tool

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/krotik/common/stringutil"
)

/*
osArgs is a local copy of os.Args (used for unit tests)
*/
var osArgs = os.Args

/*
osStderr is a local copy of os.Stderr (used for unit tests)
*/
var osStderr io.Writer = os.Stderr

/*
osExit is a local variable pointing to os.Exit (used for unit tests)
*/
var osExit func(int) = os.Exit

/*
CLIInputHandler is a handler object for CLI input.
*/
type CLIInputHandler interface {

	/*
	   CanHandle checks if a given string can be handled by this handler.
	*/
	CanHandle(s string) bool

	/*
	   Handle handles a given input string.
	*/
	Handle(ot OutputTerminal, input string)
}

/*
OutputTerminal is a generic output terminal which can write strings.
*/
type OutputTerminal interface {

	/*
	   WriteString write a string on this terminal.
	*/
	WriteString(s string)
}

/*
matchesFulltextSearch checks if a given text matches a given glob expression. Returns
true if an error occurs.
*/
func matchesFulltextSearch(ot OutputTerminal, text string, glob string) bool {
	var res bool

	re, err := stringutil.GlobToRegex(glob)

	if err == nil {
		res, err = regexp.MatchString(re, text)
	}

	if err != nil {
		ot.WriteString(fmt.Sprintln("Invalid search expression:", err.Error()))
		res = true
	}

	return res
}

/*
fillTableRow fills a table row of a display table.
*/
func fillTableRow(tabData []string, key string, value string) []string {

	tabData = append(tabData, key)

	valSplit := stringutil.ChunkSplit(value, 80, true)
	tabData = append(tabData, strings.TrimSpace(valSplit[0]))
	for _, valPart := range valSplit[1:] {
		tabData = append(tabData, "")
		tabData = append(tabData, strings.TrimSpace(valPart))
	}

	// Insert empty rows to ease reading

	tabData = append(tabData, "")
	tabData = append(tabData, "")

	return tabData
}
