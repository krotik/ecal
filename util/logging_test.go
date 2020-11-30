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
	"bytes"
	"fmt"
	"testing"
)

func TestLogging(t *testing.T) {

	ml := NewMemoryLogger(5)

	ml.LogDebug("test")
	ml.LogInfo("test")

	if ml.String() != `debug: test
test` {
		t.Error("Unexpected result:", ml.String())
		return
	}

	if res := fmt.Sprint(ml.Slice()); res != "[debug: test test]" {
		t.Error("Unexpected result:", res)
		return
	}

	ml.Reset()

	ml.LogError("test1")

	if res := fmt.Sprint(ml.Slice()); res != "[error: test1]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := ml.Size(); res != 1 {
		t.Error("Unexpected result:", res)
		return
	}

	// Test that the functions can be called

	nl := NewNullLogger()
	nl.LogDebug(nil, "test")
	nl.LogInfo(nil, "test")
	nl.LogError(nil, "test")

	sol := NewStdOutLogger()
	sol.stdlog = func(v ...interface{}) {}
	sol.LogDebug(nil, "test")
	sol.LogInfo(nil, "test")
	sol.LogError(nil, "test")

	ml.Reset()

	if _, err := NewLogLevelLogger(ml, "test"); err == nil || err.Error() != "Invalid log level: test" {
		t.Error("Unexpected result:", err)
		return
	}

	ml.Reset()
	ll, _ := NewLogLevelLogger(ml, "debug")
	ll.LogDebug("l", "test1")
	ll.LogInfo(nil, "test2")
	ll.LogError("l", "test3")

	if ml.String() != `debug: ltest1
<nil>test2
error: ltest3` {
		t.Error("Unexpected result:", ml.String())
		return
	}

	ml.Reset()
	ll, _ = NewLogLevelLogger(ml, "info")
	ll.LogDebug("l", "test1")
	ll.LogInfo(nil, "test2")
	ll.LogError("l", "test3")

	if ml.String() != `<nil>test2
error: ltest3` {
		t.Error("Unexpected result:", ml.String())
		return
	}

	ml.Reset()
	ll, _ = NewLogLevelLogger(ml, "error")

	if ll.Level() != "error" {
		t.Error("Unexpected level:", ll.Level())
		return
	}

	ll.LogDebug("l", "test1")
	ll.LogInfo(nil, "test2")
	ll.LogError("l", "test3")

	if ml.String() != `error: ltest3` {
		t.Error("Unexpected result:", ml.String())
		return
	}

	buf := bytes.NewBuffer(nil)
	bl := NewBufferLogger(buf)
	bl.LogDebug("l", "test1")
	bl.LogInfo(nil, "test2")
	bl.LogError("l", "test3")

	if buf.String() != `debug: ltest1
<nil>test2
error: ltest3
` {
		t.Error("Unexpected result:", buf.String())
		return
	}
}
