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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/fileutil"
)

const importTestDir = "importtest"

func TestImportLocater(t *testing.T) {
	if res, _ := fileutil.PathExists(importTestDir); res {
		os.RemoveAll(importTestDir)
	}

	err := os.Mkdir(importTestDir, 0770)
	if err != nil {
		t.Error("Could not create test dir:", err)
		return
	}

	defer func() {

		// Teardown

		if err := os.RemoveAll(importTestDir); err != nil {
			t.Error("Could not create test dir:", err)
			return
		}
	}()

	err = os.Mkdir(filepath.Join(importTestDir, "test1"), 0770)
	if err != nil {
		t.Error("Could not create test dir:", err)
		return
	}

	codecontent := "\na := 1 + 1\n"

	ioutil.WriteFile(filepath.Join(importTestDir, "test1", "myfile.ecal"),
		[]byte(codecontent), 0770)

	fil := &FileImportLocator{importTestDir}

	res, err := fil.Resolve(filepath.Join("..", "t"))

	expectedError := fmt.Sprintf("Import path is outside of code root: ..%vt",
		string(os.PathSeparator))

	if res != "" || err.Error() != expectedError {
		t.Error("Unexpected result:", res, err)
		return
	}

	res, err = fil.Resolve(filepath.Join("..", importTestDir, "x"))

	if res != "" || !strings.HasPrefix(err.Error(), "Could not import path") {
		t.Error("Unexpected result:", res, err)
		return
	}

	res, err = fil.Resolve(filepath.Join("..", importTestDir, "x"))

	if res != "" || !strings.HasPrefix(err.Error(), "Could not import path") {
		t.Error("Unexpected result:", res, err)
		return
	}

	res, err = fil.Resolve(filepath.Join("test1", "myfile.ecal"))
	errorutil.AssertOk(err)

	if res != codecontent {
		t.Error("Unexpected result:", res, err)
		return
	}

	mil := &MemoryImportLocator{make(map[string]string)}

	mil.Files["foo"] = "bar"
	mil.Files["test"] = "test1"

	_, err = mil.Resolve("xxx")

	if err.Error() != "Could not find import path: xxx" {
		t.Error("Unexpected result:", res, err)
		return
	}

	res, err = mil.Resolve("foo")
	errorutil.AssertOk(err)

	if res != "bar" {
		t.Error("Unexpected result:", res, err)
		return
	}

	res, err = mil.Resolve("test")
	errorutil.AssertOk(err)

	if res != "test1" {
		t.Error("Unexpected result:", res, err)
		return
	}
}
