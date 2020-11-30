/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package config

import (
	"testing"
)

func TestConfig(t *testing.T) {

	if res := Str(WorkerCount); res != "1" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := Bool(WorkerCount); !res {
		t.Error("Unexpected result:", res)
		return
	}

	if res := Int(WorkerCount); res != 1 {
		t.Error("Unexpected result:", res)
		return
	}
}
