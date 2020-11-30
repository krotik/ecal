/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package interpreter

import (
	"os"
	"path/filepath"

	"devt.de/krotik/common/timeutil"
	"devt.de/krotik/ecal/config"
	"devt.de/krotik/ecal/engine"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/util"
)

/*
ecalRuntimeNew is used to instantiate ECAL runtime components.
*/
type ecalRuntimeNew func(*ECALRuntimeProvider, *parser.ASTNode) parser.Runtime

/*
providerMap contains the mapping of AST nodes to runtime components for ECAL ASTs.
*/
var providerMap = map[string]ecalRuntimeNew{

	parser.NodeEOF: invalidRuntimeInst,

	parser.NodeSTRING:     stringValueRuntimeInst, // String constant
	parser.NodeNUMBER:     numberValueRuntimeInst, // Number constant
	parser.NodeIDENTIFIER: identifierRuntimeInst,  // Idendifier

	// Constructed tokens

	parser.NodeSTATEMENTS: statementsRuntimeInst, // List of statements
	parser.NodeFUNCCALL:   voidRuntimeInst,       // Function call
	parser.NodeCOMPACCESS: voidRuntimeInst,       // Composition structure access
	parser.NodeLIST:       listValueRuntimeInst,  // List value
	parser.NodeMAP:        mapValueRuntimeInst,   // Map value
	parser.NodePARAMS:     voidRuntimeInst,       // Function parameters
	parser.NodeGUARD:      guardRuntimeInst,      // Guard expressions for conditional statements

	// Condition operators

	parser.NodeGEQ: greaterequalOpRuntimeInst,
	parser.NodeLEQ: lessequalOpRuntimeInst,
	parser.NodeNEQ: notequalOpRuntimeInst,
	parser.NodeEQ:  equalOpRuntimeInst,
	parser.NodeGT:  greaterOpRuntimeInst,
	parser.NodeLT:  lessOpRuntimeInst,

	// Separators

	parser.NodeKVP:    voidRuntimeInst, // Key-value pair
	parser.NodePRESET: voidRuntimeInst, // Preset value

	// Arithmetic operators

	parser.NodePLUS: plusOpRuntimeInst,

	parser.NodeMINUS:  minusOpRuntimeInst,
	parser.NodeTIMES:  timesOpRuntimeInst,
	parser.NodeDIV:    divOpRuntimeInst,
	parser.NodeMODINT: modintOpRuntimeInst,
	parser.NodeDIVINT: divintOpRuntimeInst,

	// Assignment statement

	parser.NodeASSIGN: assignmentRuntimeInst,
	parser.NodeLET:    letRuntimeInst,

	// Import statement

	parser.NodeIMPORT: importRuntimeInst,
	parser.NodeAS:     voidRuntimeInst,

	// Sink definition

	parser.NodeSINK:       sinkRuntimeInst,
	parser.NodeKINDMATCH:  kindMatchRuntimeInst,
	parser.NodeSCOPEMATCH: scopeMatchRuntimeInst,
	parser.NodeSTATEMATCH: stateMatchRuntimeInst,
	parser.NodePRIORITY:   priorityRuntimeInst,
	parser.NodeSUPPRESSES: suppressesRuntimeInst,

	// Function definition

	parser.NodeFUNC:   funcRuntimeInst,
	parser.NodeRETURN: returnRuntimeInst,

	// Boolean operators

	parser.NodeOR:  orOpRuntimeInst,
	parser.NodeAND: andOpRuntimeInst,
	parser.NodeNOT: notOpRuntimeInst,

	// Condition operators

	parser.NodeLIKE:      likeOpRuntimeInst,
	parser.NodeIN:        inOpRuntimeInst,
	parser.NodeHASPREFIX: beginswithOpRuntimeInst,
	parser.NodeHASSUFFIX: endswithOpRuntimeInst,
	parser.NodeNOTIN:     notinOpRuntimeInst,

	// Constant terminals

	parser.NodeFALSE: falseRuntimeInst,
	parser.NodeTRUE:  trueRuntimeInst,
	parser.NodeNULL:  nullRuntimeInst,

	// Conditional statements

	parser.NodeIF: ifRuntimeInst,

	// Loop statements

	parser.NodeLOOP:     loopRuntimeInst,
	parser.NodeBREAK:    breakRuntimeInst,
	parser.NodeCONTINUE: continueRuntimeInst,

	// Try statement

	parser.NodeTRY:     tryRuntimeInst,
	parser.NodeEXCEPT:  voidRuntimeInst,
	parser.NodeFINALLY: voidRuntimeInst,
}

/*
ECALRuntimeProvider is the factory object producing runtime objects for ECAL ASTs.
*/
type ECALRuntimeProvider struct {
	Name          string                 // Name to identify the input
	ImportLocator util.ECALImportLocator // Locator object for imports
	Logger        util.Logger            // Logger object for log messages
	Processor     engine.Processor       // Processor of the ECA engine
	Cron          *timeutil.Cron         // Cron object for scheduled execution
	Debugger      util.ECALDebugger      // Optional: ECAL Debugger object
}

/*
NewECALRuntimeProvider returns a new instance of a ECAL runtime provider.
*/
func NewECALRuntimeProvider(name string, importLocator util.ECALImportLocator, logger util.Logger) *ECALRuntimeProvider {

	if importLocator == nil {

		// By default imports are located in the current directory

		importLocator = &util.FileImportLocator{Root: filepath.Dir(os.Args[0])}
	}

	if logger == nil {

		// By default we just have a memory logger

		logger = util.NewMemoryLogger(100)
	}

	proc := engine.NewProcessor(config.Int(config.WorkerCount))

	// By default ECAL should stop the triggering sequence of sinks after the
	// first sink that returns a sinkerror.

	proc.SetFailOnFirstErrorInTriggerSequence(true)

	cron := timeutil.NewCron()
	cron.Start()

	return &ECALRuntimeProvider{name, importLocator, logger, proc, cron, nil}
}

/*
Runtime returns a runtime component for a given ASTNode.
*/
func (erp *ECALRuntimeProvider) Runtime(node *parser.ASTNode) parser.Runtime {

	if instFunc, ok := providerMap[node.Name]; ok {
		return instFunc(erp, node)
	}

	return invalidRuntimeInst(erp, node)
}

/*
NewRuntimeError creates a new RuntimeError object.
*/
func (erp *ECALRuntimeProvider) NewRuntimeError(t error, d string, node *parser.ASTNode) error {
	return util.NewRuntimeError(erp.Name, t, d, node)
}

/*
NewThreadID creates a new thread ID unique to this runtime provider instance.
This ID can be safely used for the thread ID when calling Eval on a
parser.Runtime instance.
*/
func (erp *ECALRuntimeProvider) NewThreadID() uint64 {
	return erp.Processor.ThreadPool().NewThreadID()
}
