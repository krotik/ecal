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
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"devt.de/krotik/common/stringutil"
)

/*
EventTracer is a debugging interface to the engine
*/
var EventTracer = &eventTrace{lock: &sync.Mutex{}, out: os.Stdout}

/*
eventTrace handles low-level event tracing for debugging purposes
*/
type eventTrace struct {
	lock            *sync.Mutex
	eventTraceKind  []string
	eventTraceState []map[interface{}]interface{}
	out             io.Writer
}

/*
MonitorEvent adds a request to monitor certain events. The events to monitor
should match the given kind and have the given state values (nil values match
only on the key).
*/
func (et *eventTrace) MonitorEvent(kind string, state map[interface{}]interface{}) {
	et.lock.Lock()
	defer et.lock.Unlock()

	et.eventTraceKind = append(et.eventTraceKind, kind)
	et.eventTraceState = append(et.eventTraceState, state)
}

/*
Reset removes all added monitoring requests.
*/
func (et *eventTrace) Reset() {
	et.lock.Lock()
	defer et.lock.Unlock()

	et.eventTraceKind = nil
	et.eventTraceState = nil
}

/*
record records an event action.
*/
func (et *eventTrace) record(which *Event, where string, what ...interface{}) {
	et.lock.Lock()
	defer et.lock.Unlock()

	if et.eventTraceKind == nil {

		// Return in the normal case

		return
	}

	whichKind := strings.Join(which.Kind(), ".")

	// Check if the event matches

	for i, tkind := range et.eventTraceKind {
		tstate := et.eventTraceState[i]

		regexMatch, _ := regexp.MatchString(tkind, whichKind)

		if whichKind == tkind || regexMatch {

			if tstate == nil || stateMatch(tstate, which.State()) {

				fmt.Fprintln(et.out, fmt.Sprintf("%v %v", tkind, where))

				for _, w := range what {
					fmt.Fprintln(et.out, fmt.Sprintf("    %v",
						stringutil.ConvertToString(w)))
				}

				fmt.Fprintln(et.out, fmt.Sprintf("    %v", which))
			}
		}
	}
}

// Helper functions
// ================

/*
stateMatch checks if a given template matches a given event state.
*/
func stateMatch(template, state map[interface{}]interface{}) bool {

	for k, v := range template {
		if sv, ok := state[k]; !ok {
			return false
		} else if v != nil {
			regexMatch, _ := regexp.MatchString(fmt.Sprint(v), fmt.Sprint(sv))

			if v != sv && !regexMatch {
				return false
			}
		}
	}

	return true
}
