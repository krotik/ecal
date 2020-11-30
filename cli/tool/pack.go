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
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/fileutil"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

/*
CLIPacker is a commandline packing tool for ECAL. This tool can build a self
contained executable.
*/
type CLIPacker struct {
	EntryFile string // Entry file for the program

	// Parameter these can either be set programmatically or via CLI args

	Dir          *string // Root dir for interpreter (all files will be collected)
	SourceBinary *string // Binary which is used by the packer
	TargetBinary *string // Binary which will be build by the packer

	// Log output

	LogOut io.Writer
}

/*
NewCLIPacker creates a new commandline packer.
*/
func NewCLIPacker() *CLIPacker {
	return &CLIPacker{"", nil, nil, nil, os.Stdout}
}

/*
ParseArgs parses the command line arguments. Returns true if the program should exit.
*/
func (p *CLIPacker) ParseArgs() bool {

	if p.Dir != nil && p.TargetBinary != nil && p.EntryFile != "" {
		return false
	}

	binname, err := filepath.Abs(os.Args[0])
	errorutil.AssertOk(err)

	wd, _ := os.Getwd()

	p.Dir = flag.String("dir", wd, "Root directory for ECAL interpreter")
	p.SourceBinary = flag.String("source", binname, "Filename for source binary")
	p.TargetBinary = flag.String("target", "out.bin", "Filename for target binary")
	showHelp := flag.Bool("help", false, "Show this help message")

	flag.Usage = func() {
		fmt.Println()
		fmt.Println(fmt.Sprintf("Usage of %s pack [options] [entry file]", os.Args[0]))
		fmt.Println()
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("This tool will collect all files in the root directory and " +
			"build a standalone executable from the given source binary and the collected files.")
		fmt.Println()
	}

	if len(os.Args) >= 2 {
		flag.CommandLine.Parse(os.Args[2:])

		if cargs := flag.Args(); len(cargs) > 0 {
			p.EntryFile = flag.Arg(0)
		} else {
			*showHelp = true
		}

		if *showHelp {
			flag.Usage()
		}
	}

	return *showHelp
}

/*
Pack builds a standalone executable from a given source binary and collected files.
*/
func (p *CLIPacker) Pack() error {
	if p.ParseArgs() {
		return nil
	}

	fmt.Fprintln(p.LogOut, fmt.Sprintf("Packing %v -> %v from %v with entry: %v", *p.Dir,
		*p.TargetBinary, *p.SourceBinary, p.EntryFile))

	source, err := os.Open(*p.SourceBinary)
	if err == nil {
		var dest *os.File
		defer source.Close()

		if dest, err = os.Create(*p.TargetBinary); err == nil {
			var bytes int64

			defer dest.Close()

			// First copy the binary

			if bytes, err = io.Copy(dest, source); err == nil {
				fmt.Fprintln(p.LogOut, fmt.Sprintf("Copied %v bytes for interpreter.", bytes))
				var bytes int

				end := "####"
				marker := fmt.Sprintf("\n%v%v%v\n", end, "ECALSRC", end)

				if bytes, err = dest.WriteString(marker); err == nil {
					var data []byte
					fmt.Fprintln(p.LogOut, fmt.Sprintf("Writing marker %v bytes for source archive.", bytes))

					// Create a new zip archive.

					w := zip.NewWriter(dest)

					if data, err = ioutil.ReadFile(p.EntryFile); err == nil {
						var f io.Writer
						if f, err = w.Create(".ecalsrc-entry"); err == nil {
							if bytes, err = f.Write(data); err == nil {
								fmt.Fprintln(p.LogOut, fmt.Sprintf("Writing %v bytes for intro", bytes))

								// Add files to the archive

								if err = p.packFiles(w, *p.Dir, ""); err == nil {
									err = w.Close()

									os.Chmod(*p.TargetBinary, 0775) // Try a chmod but don't care about any errors
								}
							}
						}
					}
				}
			}
		}
	}

	return err
}

/*
packFiles walk through a given file structure and copies all files into a given zip writer.
*/
func (p *CLIPacker) packFiles(w *zip.Writer, filePath string, zipPath string) error {
	var bytes int
	files, err := ioutil.ReadDir(filePath)

	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				var data []byte
				if data, err = ioutil.ReadFile(filepath.Join(filePath, file.Name())); err == nil {
					var f io.Writer
					if f, err = w.Create(filepath.Join(zipPath, file.Name())); err == nil {
						if bytes, err = f.Write(data); err == nil {
							fmt.Fprintln(p.LogOut, fmt.Sprintf("Writing %v bytes for %v",
								bytes, filepath.Join(filePath, file.Name())))
						}
					}
				}
			} else if file.IsDir() {
				p.packFiles(w, filepath.Join(filePath, file.Name()),
					filepath.Join(zipPath, file.Name()))
			}
		}
	}

	return err
}

/*
RunPackedBinary runs ECAL code is it has been attached to the currently running binary.
Exits if attached ECAL code has been executed.
*/
func RunPackedBinary() {
	var retCode = 0
	var result bool

	exename, err := filepath.Abs(os.Args[0])
	errorutil.AssertOk(err)

	end := "####"
	marker := fmt.Sprintf("\n%v%v%v\n", end, "ECALSRC", end)

	if ok, _ := fileutil.PathExists(exename); !ok {

		// Try an optional .exe suffix which might work on Windows

		exename += ".exe"
	}

	stat, err := os.Stat(exename)
	if err == nil {
		var f *os.File

		if f, err = os.Open(exename); err == nil {
			var pos int64

			defer f.Close()

			found := false
			buf := make([]byte, 4096)
			buf2 := make([]byte, len(marker)+11)

			// Look for the marker which marks the beginning of the attached zip file

			for i, err := f.Read(buf); err == nil; i, err = f.Read(buf) {

				// Check if the marker could be in the read string

				if strings.Contains(string(buf), "#") {

					// Marker was found - read a bit more to ensure we got the full marker

					if i2, err := f.Read(buf2); err == nil || err == io.EOF {
						candidateString := string(append(buf, buf2...))

						// Now determine the position if the zip file

						markerIndex := strings.Index(candidateString, marker)

						if found = markerIndex >= 0; found {
							start := int64(markerIndex + len(marker))
							for unicode.IsSpace(rune(candidateString[start])) || unicode.IsControl(rune(candidateString[start])) {
								start++ // Skip final control characters \n or \r\n
							}
							pos += start
							break
						}

						pos += int64(i2)
					}
				}

				pos += int64(i)
			}

			if err == nil && found {

				// Extract the zip

				if _, err = f.Seek(pos, 0); err == nil {
					var ret interface{}

					zipLen := stat.Size() - pos

					ret, err = runInterpreter(io.NewSectionReader(f, pos, zipLen), zipLen)

					if retNum, ok := ret.(float64); ok {
						retCode = int(retNum)
					}

					result = err == nil
				}
			}
		}
	}

	errorutil.AssertOk(err)

	if result {
		os.Exit(retCode)
	}
}

func runInterpreter(reader io.ReaderAt, size int64) (interface{}, error) {
	var res interface{}
	var rc io.ReadCloser

	il := &util.MemoryImportLocator{Files: make(map[string]string)}

	r, err := zip.NewReader(reader, size)

	if err == nil {

		for _, f := range r.File {

			if rc, err = f.Open(); err == nil {
				var data []byte

				defer rc.Close()

				if data, err = ioutil.ReadAll(rc); err == nil {
					il.Files[f.Name] = string(data)
				}
			}

			if err != nil {
				break
			}
		}
	}

	if err == nil {
		var ast *parser.ASTNode

		erp := interpreter.NewECALRuntimeProvider(os.Args[0], il, util.NewStdOutLogger())

		if ast, err = parser.ParseWithRuntime(os.Args[0], il.Files[".ecalsrc-entry"], erp); err == nil {
			if err = ast.Runtime.Validate(); err == nil {
				var osArgs []interface{}

				vs := scope.NewScope(scope.GlobalScope)
				for _, arg := range os.Args {
					osArgs = append(osArgs, arg)
				}
				vs.SetValue("osArgs", osArgs)

				res, err = ast.Runtime.Eval(vs, make(map[string]interface{}), erp.NewThreadID())

				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())

					if terr, ok := err.(util.TraceableRuntimeError); ok {
						fmt.Fprintln(os.Stderr, fmt.Sprint("  ", strings.Join(terr.GetTraceString(), fmt.Sprint(fmt.Sprintln(), "  "))))
					}

					err = nil
				}
			}
		}
	}

	return res, err
}
