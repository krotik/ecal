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
	"strings"

	"devt.de/krotik/common/stringutil"
)

/*
Event data structure
*/
type Event struct {
	name  string                      // Name of the event
	kind  []string                    // Kind of the event (dot notation expressed as array)
	state map[interface{}]interface{} // Event state
}

/*
NewEvent returns a new event object.
*/
func NewEvent(name string, kind []string, state map[interface{}]interface{}) *Event {
	return &Event{name, kind, state}
}

/*
Name returns the event name.
*/
func (e *Event) Name() string {
	return e.name
}

/*
Kind returns the event kind.
*/
func (e *Event) Kind() []string {
	return e.kind
}

/*
State returns the event state.
*/
func (e *Event) State() map[interface{}]interface{} {
	return e.state
}

func (e *Event) String() string {
	return fmt.Sprintf("Event: %v %v %v", e.name, strings.Join(e.kind, "."),
		stringutil.ConvertToString(e.state))
}
