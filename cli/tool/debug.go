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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/util"
)

/*
CLIDebugInterpreter is a commandline interpreter with debug capabilities for ECAL.
*/
type CLIDebugInterpreter struct {
	*CLIInterpreter

	// Parameter these can either be set programmatically or via CLI args

	DebugServerAddr *string // Debug server address
	RunDebugServer  *bool   // Run a debug server
	EchoDebugServer *bool   // Echo all input and output of the debug server
	Interactive     *bool   // Flag if the interpreter should open a console in the current tty.
	BreakOnStart    *bool   // Flag if the debugger should stop the execution on start
	BreakOnError    *bool   // Flag if the debugger should stop when encountering an error

	// Log output

	LogOut io.Writer
}

/*
NewCLIDebugInterpreter wraps an existing CLIInterpreter object and adds capabilities.
*/
func NewCLIDebugInterpreter(i *CLIInterpreter) *CLIDebugInterpreter {
	return &CLIDebugInterpreter{i, nil, nil, nil, nil, nil, nil, os.Stdout}
}

/*
ParseArgs parses the command line arguments.
*/
func (i *CLIDebugInterpreter) ParseArgs() bool {

	if i.Interactive != nil {
		return false
	}

	i.DebugServerAddr = flag.String("serveraddr", "localhost:33274", "Debug server address") // Think BERTA
	i.RunDebugServer = flag.Bool("server", false, "Run a debug server")
	i.EchoDebugServer = flag.Bool("echo", false, "Echo all i/o of the debug server")
	i.Interactive = flag.Bool("interactive", true, "Run interactive console")
	i.BreakOnStart = flag.Bool("breakonstart", false, "Stop the execution on start")
	i.BreakOnError = flag.Bool("breakonerror", false, "Stop the execution when encountering an error")

	return i.CLIInterpreter.ParseArgs()
}

/*
Interpret starts the ECAL code interpreter with debug capabilities.
*/
func (i *CLIDebugInterpreter) Interpret() error {

	if i.ParseArgs() {
		return nil
	}

	err := i.CreateRuntimeProvider("debug console")

	if err == nil {

		// Set custom messages

		i.CLIInterpreter.CustomWelcomeMessage = "Running in debug mode - "
		if *i.RunDebugServer {
			i.CLIInterpreter.CustomWelcomeMessage += fmt.Sprintf("with debug server on %v - ", *i.DebugServerAddr)
		}
		i.CLIInterpreter.CustomWelcomeMessage += "prefix debug commands with ##"
		i.CustomHelpString = "    @dbg [glob] - List all available debug commands.\n"

		// Set debug object on the runtime provider

		i.RuntimeProvider.Debugger = interpreter.NewECALDebugger(i.GlobalVS)
		i.RuntimeProvider.Debugger.BreakOnStart(*i.BreakOnStart)
		i.RuntimeProvider.Debugger.BreakOnError(*i.BreakOnError)

		// Set this object as a custom handler to deal with input.

		i.CustomHandler = i

		if *i.RunDebugServer {

			// Start the debug server

			debugServer := &debugTelnetServer{*i.DebugServerAddr, "ECALDebugServer: ",
				nil, true, *i.EchoDebugServer, i, i.RuntimeProvider.Logger}

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go debugServer.Run(wg)
			wg.Wait()

			defer func() {
				if debugServer.listener != nil {
					debugServer.listen = false
					debugServer.listener.Close() // Attempt to cleanup
				}
			}()
		}

		err = i.CLIInterpreter.Interpret(*i.Interactive)
	}

	return err
}

/*
LoadInitialFile clears the global scope and reloads the initial file.
*/
func (i *CLIDebugInterpreter) LoadInitialFile(tid uint64) error {
	i.RuntimeProvider.Debugger.StopThreads(500 * time.Millisecond)
	i.RuntimeProvider.Debugger.BreakOnStart(*i.BreakOnStart)
	i.RuntimeProvider.Debugger.BreakOnError(*i.BreakOnError)
	return nil
}

/*
CanHandle checks if a given string can be handled by this handler.
*/
func (i *CLIDebugInterpreter) CanHandle(s string) bool {
	return strings.HasPrefix(s, "##") || strings.HasPrefix(s, "@dbg")
}

/*
Handle handles a given input string.
*/
func (i *CLIDebugInterpreter) Handle(ot OutputTerminal, line string) {

	if strings.HasPrefix(line, "@dbg") {

		args := strings.Fields(line)[1:]

		tabData := []string{"Debug command", "Description"}

		for name, f := range interpreter.DebugCommandsMap {
			ds := f.DocString()

			if len(args) > 0 && !matchesFulltextSearch(ot, fmt.Sprintf("%v %v", name, ds), args[0]) {
				continue
			}

			tabData = fillTableRow(tabData, name, ds)
		}

		if len(tabData) > 2 {
			ot.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
				stringutil.SingleDoubleLineTable))
		}
		ot.WriteString(fmt.Sprintln(fmt.Sprintln()))

	} else {
		res, err := i.RuntimeProvider.Debugger.HandleInput(strings.TrimSpace(line[2:]))

		if err == nil {
			var outBytes []byte
			outBytes, err = json.MarshalIndent(res, "", "  ")
			if err == nil {
				ot.WriteString(fmt.Sprintln(fmt.Sprintln(string(outBytes))))
			}
		}

		if err != nil {
			var outBytes []byte
			outBytes, err = json.MarshalIndent(map[string]interface{}{
				"DebuggerError": err.Error(),
			}, "", "  ")
			errorutil.AssertOk(err)
			ot.WriteString(fmt.Sprintln(fmt.Sprintln(string(outBytes))))
		}
	}
}

/*
debugTelnetServer is a simple telnet server to send and receive debug data.
*/
type debugTelnetServer struct {
	address     string
	logPrefix   string
	listener    *net.TCPListener
	listen      bool
	echo        bool
	interpreter *CLIDebugInterpreter
	logger      util.Logger
}

/*
Run runs the debug server.
*/
func (s *debugTelnetServer) Run(wg *sync.WaitGroup) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", s.address)

	if err == nil {

		s.listener, err = net.ListenTCP("tcp", tcpaddr)

		if err == nil {

			wg.Done()

			s.logger.LogInfo(s.logPrefix,
				"Running Debug Server on ", tcpaddr.String())

			for s.listen {
				var conn net.Conn

				if conn, err = s.listener.Accept(); err == nil {
					go s.HandleConnection(conn)

				} else if s.listen {
					s.logger.LogError(s.logPrefix, err)
					err = nil
				}
			}
		}
	}

	if s.listen && err != nil {
		s.logger.LogError(s.logPrefix, "Could not start debug server - ", err)
		wg.Done()
	}
}

/*
HandleConnection handles an incoming connection.
*/
func (s *debugTelnetServer) HandleConnection(conn net.Conn) {
	tid := s.interpreter.RuntimeProvider.NewThreadID()
	inputReader := bufio.NewReader(conn)
	outputTerminal := OutputTerminal(&bufioWriterShim{fmt.Sprint(conn.RemoteAddr()),
		bufio.NewWriter(conn), s.echo, s.interpreter.LogOut})

	line := ""

	s.logger.LogDebug(s.logPrefix, "Connect ", conn.RemoteAddr())
	if s.echo {
		fmt.Fprintln(s.interpreter.LogOut, fmt.Sprintf("%v : Connected", conn.RemoteAddr()))
	}

	for {
		var outBytes []byte
		var err error

		if line, err = inputReader.ReadString('\n'); err == nil {
			line = strings.TrimSpace(line)

			if s.echo {
				fmt.Fprintln(s.interpreter.LogOut, fmt.Sprintf("%v > %v", conn.RemoteAddr(), line))
			}

			if line == "exit" || line == "q" || line == "quit" || line == "bye" || line == "\x04" {
				break
			}

			isHelpTable := strings.HasPrefix(line, "@")

			if !s.interpreter.CanHandle(line) || isHelpTable {
				buffer := bytes.NewBuffer(nil)

				s.interpreter.HandleInput(&bufioWriterShim{"tmpbuffer",
					bufio.NewWriter(buffer), false, s.interpreter.LogOut}, line, tid)

				if isHelpTable {

					// Special case we have tables which should be transformed

					r := strings.NewReplacer("═", "*", "│", "*", "╪", "*", "╒", "*",
						"╕", "*", "╘", "*", "╛", "*", "╤", "*", "╞", "*", "╡", "*", "╧", "*")

					outBytes = []byte(r.Replace(buffer.String()))

				} else {

					outBytes = buffer.Bytes()
				}

				outBytes, err = json.MarshalIndent(map[string]interface{}{
					"EncodedOutput": base64.StdEncoding.EncodeToString(outBytes),
				}, "", "  ")
				errorutil.AssertOk(err)
				outputTerminal.WriteString(fmt.Sprintln(fmt.Sprintln(string(outBytes))))

			} else {

				s.interpreter.HandleInput(outputTerminal, line, tid)
			}
		}

		if err != nil {
			if s.echo {
				fmt.Fprintln(s.interpreter.LogOut, fmt.Sprintf("%v : Disconnected", conn.RemoteAddr()))
			}
			s.logger.LogDebug(s.logPrefix, "Disconnect ", conn.RemoteAddr(), " - ", err)
			break
		}
	}

	conn.Close()
}

/*
bufioWriterShim is a shim to allow a bufio.Writer to be used as an OutputTerminal.
*/
type bufioWriterShim struct {
	id     string
	writer *bufio.Writer
	echo   bool
	logOut io.Writer
}

/*
WriteString write a string to the writer.
*/
func (shim *bufioWriterShim) WriteString(s string) {
	if shim.echo {
		fmt.Fprintln(shim.logOut, fmt.Sprintf("%v < %v", shim.id, s))
	}
	shim.writer.WriteString(s)
	shim.writer.Flush()
}
