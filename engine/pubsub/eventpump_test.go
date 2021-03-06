/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package pubsub

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"testing"
)

var res []string
var source1 = &bytes.Buffer{}
var errSource2 error = fmt.Errorf("TEST")
var ep = NewEventPump()

func addObservers2(t *testing.T) {

	// Add observer 4

	ep.AddObserver("", source1, func(event string, eventSource interface{}) {
		if eventSource != source1 {
			t.Error("Unexpected event source:", eventSource)
			return
		}
		res = append(res, "4")
		sort.Strings(res)
	})

	// Add observer 5

	ep.AddObserver("", nil, func(event string, eventSource interface{}) {
		res = append(res, "5")
		sort.Strings(res)
	})

	// Add observer 6

	ep.AddObserver("", errSource2, func(event string, eventSource interface{}) {
		if eventSource != errSource2 {
			t.Error("Unexpected event source:", eventSource)
			return
		}
		res = append(res, "6")
		sort.Strings(res)
	})
}

func TestEventPump(t *testing.T) {

	addObservers1(t)

	// Run the tests

	// Test 1 straight forward case

	ep.PostEvent("event1", source1)

	if fmt.Sprint(res) != "[1]" {
		t.Error("Unexpected result:", res)
		return
	}

	res = make([]string, 0) // Reset res

	ep.PostEvent("event2", errSource2)

	if fmt.Sprint(res) != "[2 3]" {
		t.Error("Unexpected result:", res)
		return
	}

	res = make([]string, 0) // Reset res

	ep.PostEvent("event1", errSource2)

	if fmt.Sprint(res) != "[]" {
		t.Error("Unexpected result:", res)
		return
	}

	addObservers2(t)

	res = make([]string, 0) // Reset res

	ep.PostEvent("event1", errSource2)

	if fmt.Sprint(res) != "[5 6]" {
		t.Error("Unexpected result:", res)
		return
	}

	res = make([]string, 0) // Reset res

	ep.PostEvent("event3", errSource2)

	if fmt.Sprint(res) != "[5 6]" {
		t.Error("Unexpected result:", res)
		return
	}

	res = make([]string, 0) // Reset res

	ep.PostEvent("event3", source1)

	if fmt.Sprint(res) != "[4 5]" {
		t.Error("Unexpected result:", res)
		return
	}

	res = make([]string, 0) // Reset res

	ep.PostEvent("event3", errors.New("test"))

	if fmt.Sprint(res) != "[5]" {
		t.Error("Unexpected result:", res)
		return
	}

	// Remove observers

	res = make([]string, 0) // Reset res

	ep.PostEvent("event2", errSource2)

	if fmt.Sprint(res) != "[2 3 5 6]" {
		t.Error("Unexpected result:", res)
		return
	}
	ep.RemoveObservers("event2", errSource2)

	res = make([]string, 0) // Reset res

	ep.PostEvent("event2", errSource2)

	if fmt.Sprint(res) != "[5 6]" {
		t.Error("Unexpected result:", res)
		return
	}

	ep.RemoveObservers("", errSource2) // Remove all handlers specific to source 2

	res = make([]string, 0) // Reset res

	ep.PostEvent("event2", errSource2)

	if fmt.Sprint(res) != "[5]" {
		t.Error("Unexpected result:", res)
		return
	}

	ep.PostEvent("event1", source1)

	if fmt.Sprint(res) != "[1 4 5 5]" {
		t.Error("Unexpected result:", res)
		return
	}

	ep.RemoveObservers("event1", nil) // Remove all handlers specific to source 2

	res = make([]string, 0) // Reset res

	ep.PostEvent("event2", errSource2)

	if fmt.Sprint(res) != "[5]" {
		t.Error("Unexpected result:", res)
		return
	}

	ep.RemoveObservers("", nil) // Remove all handlers

	res = make([]string, 0) // Reset res

	ep.PostEvent("event2", errSource2)

	if fmt.Sprint(res) != "[]" {
		t.Error("Unexpected result:", res)
		return
	}

	// This call should be ignored

	ep.AddObserver("event1", source1, nil)

	if fmt.Sprint(ep.eventsObservers) != "map[]" {
		t.Error("Event map should be empty at this point:", ep.eventsObservers)
		return
	}
}

func addObservers1(t *testing.T) {

	// Add observer 1

	ep.AddObserver("event1", source1, func(event string, eventSource interface{}) {
		if eventSource != source1 {
			t.Error("Unexpected event source:", eventSource)
			return
		}
		res = append(res, "1")
		sort.Strings(res)

	})

	// Add observer 2

	ep.AddObserver("event2", errSource2, func(event string, eventSource interface{}) {
		if eventSource != errSource2 {
			t.Error("Unexpected event source:", eventSource)
			return
		}
		res = append(res, "2")
		sort.Strings(res)

	})

	// Add observer 3

	ep.AddObserver("event2", errSource2, func(event string, eventSource interface{}) {
		if eventSource != errSource2 {
			t.Error("Unexpected event source:", eventSource)
			return
		}
		res = append(res, "3")
		sort.Strings(res)

	})
}

func TestWrongPostEvent(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Posting events with empty values shouldn't work.")
		}
	}()

	ep := NewEventPump()
	ep.PostEvent("", nil)
}
