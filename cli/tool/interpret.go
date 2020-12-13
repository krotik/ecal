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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"devt.de/krotik/common/fileutil"
	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/common/termutil"
	"devt.de/krotik/ecal/config"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/stdlib"
	"devt.de/krotik/ecal/util"
)

/*
CLICustomHandler is a handler for custom operations.
*/
type CLICustomHandler interface {
	CLIInputHandler

	/*
	   LoadInitialFile clears the global scope and reloads the initial file.
	*/
	LoadInitialFile(tid uint64) error
}

/*
CLIInterpreter is a commandline interpreter for ECAL.
*/
type CLIInterpreter struct {
	GlobalVS        parser.Scope                     // Global variable scope
	RuntimeProvider *interpreter.ECALRuntimeProvider // Runtime provider of the interpreter

	// Customizations of output and input handling

	CustomHandler        CLICustomHandler
	CustomWelcomeMessage string
	CustomHelpString     string

	EntryFile   string // Entry file for the program
	LoadPlugins bool   // Flag if stdlib plugins should be loaded

	// Parameter these can either be set programmatically or via CLI args

	Dir      *string // Root dir for interpreter
	LogFile  *string // Logfile (blank for stdout)
	LogLevel *string // Log level string (Debug, Info, Error)

	// User terminal

	Term termutil.ConsoleLineTerminal

	// Log output

	LogOut io.Writer
}

/*
NewCLIInterpreter creates a new commandline interpreter for ECAL.
*/
func NewCLIInterpreter() *CLIInterpreter {
	return &CLIInterpreter{scope.NewScope(scope.GlobalScope), nil, nil, "", "", "",
		true, nil, nil, nil, nil, os.Stdout}
}

/*
ParseArgs parses the command line arguments. Call this after adding custon flags.
Returns true if the program should exit.
*/
func (i *CLIInterpreter) ParseArgs() bool {

	if i.Dir != nil && i.LogFile != nil && i.LogLevel != nil {
		return false
	}

	wd, _ := os.Getwd()

	i.Dir = flag.String("dir", wd, "Root directory for ECAL interpreter")
	i.LogFile = flag.String("logfile", "", "Log to a file")
	i.LogLevel = flag.String("loglevel", "Info", "Logging level (Debug, Info, Error)")
	showHelp := flag.Bool("help", false, "Show this help message")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), fmt.Sprintf("Usage of %s run [options] [file]", osArgs[0]))
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output())
	}

	if len(osArgs) >= 2 {
		flag.CommandLine.Parse(osArgs[2:])

		if cargs := flag.Args(); len(cargs) > 0 {
			i.EntryFile = flag.Arg(0)
		}

		if *showHelp {
			flag.Usage()
		}
	}

	return *showHelp
}

/*
CreateRuntimeProvider creates the runtime provider of this interpreter. This function expects Dir,
LogFile and LogLevel to be set.
*/
func (i *CLIInterpreter) CreateRuntimeProvider(name string) error {
	var logger util.Logger
	var err error

	if i.RuntimeProvider != nil {
		return nil
	}

	// Check if we should log to a file

	if i.LogFile != nil && *i.LogFile != "" {
		var logWriter io.Writer

		logFileRollover := fileutil.SizeBasedRolloverCondition(1000000) // Each file can be up to a megabyte
		logWriter, err = fileutil.NewMultiFileBuffer(*i.LogFile, fileutil.ConsecutiveNumberIterator(10), logFileRollover)
		logger = util.NewBufferLogger(logWriter)

	} else {

		// Log to the console by default

		logger = util.NewStdOutLogger()
	}

	// Set the log level

	if err == nil {
		if i.LogLevel != nil && *i.LogLevel != "" {
			logger, err = util.NewLogLevelLogger(logger, *i.LogLevel)
		}

		if err == nil {
			// Get the import locator

			importLocator := &util.FileImportLocator{Root: *i.Dir}

			// Create interpreter

			i.RuntimeProvider = interpreter.NewECALRuntimeProvider(name, importLocator, logger)
		}
	}

	return err
}

/*
LoadInitialFile clears the global scope and reloads the initial file.
*/
func (i *CLIInterpreter) LoadInitialFile(tid uint64) error {
	var err error

	if i.CustomHandler != nil {
		i.CustomHandler.LoadInitialFile(tid)
	}

	i.GlobalVS.Clear()

	if i.EntryFile != "" {
		var ast *parser.ASTNode
		var initFile []byte

		initFile, err = ioutil.ReadFile(i.EntryFile)

		if err == nil {
			if ast, err = parser.ParseWithRuntime(i.EntryFile, string(initFile), i.RuntimeProvider); err == nil {
				if err = ast.Runtime.Validate(); err == nil {
					_, err = ast.Runtime.Eval(i.GlobalVS, make(map[string]interface{}), tid)
				}
				defer func() {
					if i.RuntimeProvider.Debugger != nil {
						i.RuntimeProvider.Debugger.RecordThreadFinished(tid)
					}
				}()
			}
		}
	}

	return err
}

/*
CreateTerm creates a new console terminal for stdout.
*/
func (i *CLIInterpreter) CreateTerm() error {
	var err error

	if i.Term == nil {
		i.Term, err = termutil.NewConsoleLineTerminal(os.Stdout)
	}

	return err
}

/*
Interpret starts the ECAL code interpreter. Starts an interactive console in
the current tty if the interactive flag is set.
*/
func (i *CLIInterpreter) Interpret(interactive bool) error {

	if i.ParseArgs() {
		return nil
	}

	err := i.LoadStdlibPlugins(interactive)

	if err == nil {
		err = i.CreateTerm()

		if interactive {
			fmt.Fprintln(i.LogOut, fmt.Sprintf("ECAL %v", config.ProductVersion))
		}

		// Create Runtime Provider

		if err == nil {

			if err = i.CreateRuntimeProvider("console"); err == nil {

				tid := i.RuntimeProvider.NewThreadID()

				if interactive {
					if lll, ok := i.RuntimeProvider.Logger.(*util.LogLevelLogger); ok {
						fmt.Fprint(i.LogOut, fmt.Sprintf("Log level: %v - ", lll.Level()))
					}

					fmt.Fprintln(i.LogOut, fmt.Sprintf("Root directory: %v", *i.Dir))

					if i.CustomWelcomeMessage != "" {
						fmt.Fprintln(i.LogOut, fmt.Sprintf(i.CustomWelcomeMessage))
					}
				}

				// Execute file if given

				if err = i.LoadInitialFile(tid); err == nil {

					// Drop into interactive shell

					if interactive {

						// Add history functionality without file persistence

						i.Term, err = termutil.AddHistoryMixin(i.Term, "",
							func(s string) bool {
								return i.isExitLine(s)
							})

						if err == nil {

							if err = i.Term.StartTerm(); err == nil {
								var line string

								defer i.Term.StopTerm()

								fmt.Fprintln(i.LogOut, "Type 'q' or 'quit' to exit the shell and '?' to get help")

								line, err = i.Term.NextLine()
								for err == nil && !i.isExitLine(line) {
									trimmedLine := strings.TrimSpace(line)

									i.HandleInput(i.Term, trimmedLine, tid)

									line, err = i.Term.NextLine()
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
LoadStdlibPlugins load plugins from .ecal.json.
*/
func (i *CLIInterpreter) LoadStdlibPlugins(interactive bool) error {
	var err error

	if i.LoadPlugins {
		confFile := filepath.Join(*i.Dir, ".ecal.json")
		if ok, _ := fileutil.PathExists(confFile); ok {

			if interactive {
				fmt.Fprintln(i.LogOut, fmt.Sprintf("Loading stdlib plugins from %v", confFile))
			}

			var content []byte
			if content, err = ioutil.ReadFile(confFile); err == nil {
				var conf map[string]interface{}
				if err = json.Unmarshal(content, &conf); err == nil {
					if stdlibPlugins, ok := conf["stdlibPlugins"]; ok {
						err = fmt.Errorf("Config stdlibPlugins should be a list")
						if plugins, ok := stdlibPlugins.([]interface{}); ok {
							err = nil
							if errs := stdlib.LoadStdlibPlugins(plugins); len(errs) > 0 {
								for _, e := range errs {
									fmt.Fprintln(i.LogOut, fmt.Sprintf("Error loading plugins: %v", e))
								}
								err = fmt.Errorf("Could not load plugins defined in .ecal.json")
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
isExitLine returns if a given input line should exit the interpreter.
*/
func (i *CLIInterpreter) isExitLine(s string) bool {
	return s == "exit" || s == "q" || s == "quit" || s == "bye" || s == "\x04"
}

/*
HandleInput handles input to this interpreter. It parses a given input line
and outputs on the given output terminal. Requires a thread ID of the executing
thread - use the RuntimeProvider to generate a unique one.
*/
func (i *CLIInterpreter) HandleInput(ot OutputTerminal, line string, tid uint64) {

	// Process the entered line

	if line == "?" {

		// Show help

		ot.WriteString(fmt.Sprintf("ECAL %v\n", config.ProductVersion))
		ot.WriteString(fmt.Sprint("\n"))
		ot.WriteString(fmt.Sprint("Console supports all normal ECAL statements and the following special commands:\n"))
		ot.WriteString(fmt.Sprint("\n"))
		ot.WriteString(fmt.Sprint("    @reload - Clear the interpreter and reload the initial file if it was given.\n"))
		ot.WriteString(fmt.Sprint("    @sym [glob] - List all available inbuild functions and available stdlib packages of ECAL.\n"))
		ot.WriteString(fmt.Sprint("    @std <package> [glob] - List all available constants and functions of a stdlib package.\n"))
		if i.CustomHelpString != "" {
			ot.WriteString(i.CustomHelpString)
		}
		ot.WriteString(fmt.Sprint("\n"))
		ot.WriteString(fmt.Sprint("Add an argument after a list command to do a full text search. The search string should be in glob format.\n"))

	} else if strings.HasPrefix(line, "@reload") {

		// Reload happens in a separate thread as it may be suspended on start

		go i.LoadInitialFile(i.RuntimeProvider.NewThreadID())
		ot.WriteString(fmt.Sprintln(fmt.Sprintln("Reloading interpreter state")))

	} else if strings.HasPrefix(line, "@sym") {
		i.displaySymbols(ot, strings.Split(line, " ")[1:])

	} else if strings.HasPrefix(line, "@std") {
		i.displayPackage(ot, strings.Split(line, " ")[1:])

	} else if i.CustomHandler != nil && i.CustomHandler.CanHandle(line) {
		i.CustomHandler.Handle(ot, line)

	} else {
		var ierr error
		var ast *parser.ASTNode
		var res interface{}

		if line != "" {
			if ast, ierr = parser.ParseWithRuntime("console input", line, i.RuntimeProvider); ierr == nil {

				if ierr = ast.Runtime.Validate(); ierr == nil {

					if res, ierr = ast.Runtime.Eval(i.GlobalVS, make(map[string]interface{}), tid); ierr == nil && res != nil {
						ot.WriteString(fmt.Sprintln(stringutil.ConvertToString(res)))
					}
					defer func() {
						if i.RuntimeProvider.Debugger != nil {
							i.RuntimeProvider.Debugger.RecordThreadFinished(tid)
						}
					}()
				}
			}

			if ierr != nil {
				ot.WriteString(fmt.Sprintln(ierr.Error()))
			}
		}
	}
}

/*
displaySymbols lists all available inbuild functions and available stdlib packages of ECAL.
*/
func (i *CLIInterpreter) displaySymbols(ot OutputTerminal, args []string) {

	tabData := []string{"Inbuild function", "Description"}

	for name, f := range interpreter.InbuildFuncMap {
		ds, _ := f.DocString()

		if len(args) > 0 && !matchesFulltextSearch(ot, fmt.Sprintf("%v %v", name, ds), args[0]) {
			continue
		}

		tabData = fillTableRow(tabData, name, ds)
	}

	if len(tabData) > 2 {
		ot.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}

	packageNames, _, _ := stdlib.GetStdlibSymbols()

	tabData = []string{"Package name", "Description"}

	for _, p := range packageNames {
		ps, _ := stdlib.GetPkgDocString(p)

		if len(args) > 0 && !matchesFulltextSearch(ot, fmt.Sprintf("%v %v", p, ps), args[0]) {
			continue
		}

		tabData = fillTableRow(tabData, p, ps)
	}

	if len(tabData) > 2 {
		ot.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}
}

/*
displayPackage list all available constants and functions of a stdlib package.
*/
func (i *CLIInterpreter) displayPackage(ot OutputTerminal, args []string) {

	_, constSymbols, funcSymbols := stdlib.GetStdlibSymbols()

	tabData := []string{"Constant", "Value"}

	for _, s := range constSymbols {

		if len(args) > 0 && !strings.HasPrefix(s, args[0]) {
			continue
		}

		val, _ := stdlib.GetStdlibConst(s)

		tabData = fillTableRow(tabData, s, fmt.Sprint(val))
	}

	if len(tabData) > 2 {
		ot.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}

	tabData = []string{"Function", "Description"}

	for _, f := range funcSymbols {
		if len(args) > 0 && !strings.HasPrefix(f, args[0]) {
			continue
		}

		fObj, _ := stdlib.GetStdlibFunc(f)
		fDoc, _ := fObj.DocString()

		fDoc = strings.Replace(fDoc, "\n", " ", -1)
		fDoc = strings.Replace(fDoc, "\t", " ", -1)

		if len(args) > 1 && !matchesFulltextSearch(ot, fmt.Sprintf("%v %v", f, fDoc), args[1]) {
			continue
		}

		tabData = fillTableRow(tabData, f, fDoc)
	}

	if len(tabData) > 2 {
		ot.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}
}
