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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"devt.de/krotik/ecal/parser"
)

/*
Format formats a given set of ECAL files.
*/
func Format() error {
	wd, _ := os.Getwd()

	dir := flag.String("dir", wd, "Root directory for ECAL files")
	ext := flag.String("ext", ".ecal", "Extension for ECAL files")
	showHelp := flag.Bool("help", false, "Show this help message")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), fmt.Sprintf("Usage of %s format [options]", os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), "This tool will format all ECAL files in a directory structure.")
		fmt.Fprintln(flag.CommandLine.Output())
	}

	if len(os.Args) >= 2 {
		flag.CommandLine.Parse(osArgs[2:])

		if *showHelp {
			flag.Usage()
			return nil
		}
	}

	fmt.Fprintln(flag.CommandLine.Output(), fmt.Sprintf("Formatting all %v files in %v", *ext, *dir))

	return FormatFiles(*dir, *ext)
}

/*
FormatFiles formats all ECAL files in a given directory with a given ending.
*/
func FormatFiles(dir string, ext string) error {
	var err error

	// Try to resolve symbolic links

	scanDir, lerr := os.Readlink(dir)
	if lerr != nil {
		scanDir = dir
	}

	if err == nil {
		err = filepath.Walk(scanDir,
			func(path string, i os.FileInfo, err error) error {
				if err == nil && !i.IsDir() {
					var data []byte
					var ast *parser.ASTNode
					var srcFormatted string

					if strings.HasSuffix(path, ext) {
						if data, err = ioutil.ReadFile(path); err == nil {
							var ferr error

							if ast, ferr = parser.Parse(path, string(data)); ferr == nil {
								if srcFormatted, ferr = parser.PrettyPrint(ast); ferr == nil {
									ioutil.WriteFile(path, []byte(fmt.Sprintln(srcFormatted)), i.Mode())
								}
							}

							if ferr != nil {
								fmt.Fprintln(flag.CommandLine.Output(), fmt.Sprintf("Could not format %v: %v", path, ferr))
							}
						}
					}
				}
				return err
			})
	}

	return err
}
