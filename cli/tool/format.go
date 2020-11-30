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

	"devt.de/krotik/ecal/parser"
)

func Format() error {
	var err error

	wd, _ := os.Getwd()

	dir := flag.String("dir", wd, "Root directory for ECAL files")
	ext := flag.String("ext", ".ecal", "Extension for ECAL files")
	showHelp := flag.Bool("help", false, "Show this help message")

	flag.Usage = func() {
		fmt.Println()
		fmt.Println(fmt.Sprintf("Usage of %s format [options]", os.Args[0]))
		fmt.Println()
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("This tool will format all ECAL files in a directory structure.")
		fmt.Println()
	}

	if len(os.Args) >= 2 {
		flag.CommandLine.Parse(os.Args[2:])

		if *showHelp {
			flag.Usage()
			return nil
		}
	}

	fmt.Println(fmt.Sprintf("Formatting all %v files in %v", *ext, *dir))

	err = filepath.Walk(".",
		func(path string, i os.FileInfo, err error) error {
			if err == nil && !i.IsDir() {
				var data []byte
				var ast *parser.ASTNode
				var srcFormatted string

				if data, err = ioutil.ReadFile(path); err == nil {
					var ferr error

					if ast, ferr = parser.Parse(path, string(data)); ferr == nil {
						if srcFormatted, ferr = parser.PrettyPrint(ast); ferr == nil {
							ioutil.WriteFile(path, []byte(srcFormatted), i.Mode())
						}
					}

					if ferr != nil {
						fmt.Fprintln(os.Stderr, fmt.Sprintf("Could not format %v: %v", path, ferr))
					}
				}
			}
			return err
		})

	return err
}
