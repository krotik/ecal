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
)

// ImportLocator implementations
// =============================

/*
MemoryImportLocator holds a given set of code in memory and can provide it as imports.
*/
type MemoryImportLocator struct {
	Files map[string]string
}

/*
Resolve a given import path and parse the imported file into an AST.
*/
func (il *MemoryImportLocator) Resolve(path string) (string, error) {

	res, ok := il.Files[path]

	if !ok {
		return "", fmt.Errorf("Could not find import path: %v", path)
	}

	return res, nil
}

/*
FileImportLocator tries to locate files on disk relative to a root directory and provide them as imports.
*/
type FileImportLocator struct {
	Root string // Relative root path
}

/*
Resolve a given import path and parse the imported file into an AST.
*/
func (il *FileImportLocator) Resolve(path string) (string, error) {
	var res string

	importPath := filepath.Clean(filepath.Join(il.Root, path))

	ok, err := isSubpath(il.Root, importPath)

	if err == nil && !ok {
		err = fmt.Errorf("Import path is outside of code root: %v", path)
	}

	if err == nil {
		var b []byte
		if b, err = ioutil.ReadFile(importPath); err != nil {
			err = fmt.Errorf("Could not import path %v: %v", path, err)
		} else {
			res = string(b)
		}
	}

	return res, err
}

/*
isSubpath checks if the given sub path is a child path of root.
*/
func isSubpath(root, sub string) (bool, error) {
	rel, err := filepath.Rel(root, sub)
	return err == nil &&
		!strings.HasPrefix(rel, fmt.Sprintf("..%v", string(os.PathSeparator))) &&
		rel != "..", err
}
