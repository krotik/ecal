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
	"fmt"
	"math"
	"strings"

	"devt.de/krotik/ecal/engine"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

/*
sinkRuntime is the runtime for sink declarations.
*/
type sinkRuntime struct {
	*baseRuntime
}

/*
sinkRuntimeInst returns a new runtime component instance.
*/
func sinkRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &sinkRuntime{newBaseRuntime(erp, node)}
}

/*
Validate this node and all its child nodes.
*/
func (rt *sinkRuntime) Validate() error {
	err := rt.baseRuntime.Validate()

	if err == nil {

		// Check that all children are valid

		for _, child := range rt.node.Children[1:] {
			switch child.Name {
			case parser.NodeKINDMATCH:
			case parser.NodeSCOPEMATCH:
			case parser.NodeSTATEMATCH:
			case parser.NodePRIORITY:
			case parser.NodeSUPPRESSES:
			case parser.NodeSTATEMENTS:
				continue
			default:
				err = rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
					fmt.Sprintf("Unknown expression in sink declaration %v", child.Token.Val),
					child)
			}

			if err != nil {
				break
			}
		}
	}

	return err
}

/*
Eval evaluate this runtime component.
*/
func (rt *sinkRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var kindMatch, scopeMatch, suppresses []string
	var stateMatch map[string]interface{}
	var priority int
	var statements *parser.ASTNode

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {
		// Create default scope

		scopeMatch = []string{}

		// Get the name of the sink

		name := rt.node.Children[0].Token.Val

		// Create helper function

		makeStringList := func(child *parser.ASTNode) ([]string, error) {
			var ret []string

			val, err := child.Runtime.Eval(vs, is, tid)

			if err == nil {
				for _, v := range val.([]interface{}) {
					ret = append(ret, fmt.Sprint(v))
				}
			}

			return ret, err
		}

		// Collect values from children

		for _, child := range rt.node.Children[1:] {

			switch child.Name {

			case parser.NodeKINDMATCH:
				kindMatch, err = makeStringList(child)
				break

			case parser.NodeSCOPEMATCH:
				scopeMatch, err = makeStringList(child)
				break

			case parser.NodeSTATEMATCH:
				var val interface{}
				stateMatch = make(map[string]interface{})

				if val, err = child.Runtime.Eval(vs, is, tid); err == nil {
					for k, v := range val.(map[interface{}]interface{}) {
						stateMatch[fmt.Sprint(k)] = v
					}
				}
				break

			case parser.NodePRIORITY:
				var val interface{}

				if val, err = child.Runtime.Eval(vs, is, tid); err == nil {
					priority = int(math.Floor(val.(float64)))
				}
				break

			case parser.NodeSUPPRESSES:
				suppresses, err = makeStringList(child)
				break

			case parser.NodeSTATEMENTS:
				statements = child
				break
			}

			if err != nil {
				break
			}
		}

		if err == nil && statements != nil {
			var desc string

			sinkName := fmt.Sprint(name)

			if len(rt.node.Meta) > 0 &&
				(rt.node.Meta[0].Type() == parser.MetaDataPreComment ||
					rt.node.Meta[0].Type() == parser.MetaDataPostComment) {
				desc = strings.TrimSpace(rt.node.Meta[0].Value())
			}

			rule := &engine.Rule{
				Name:            sinkName,   // Name
				Desc:            desc,       // Description
				KindMatch:       kindMatch,  // Kind match
				ScopeMatch:      scopeMatch, // Match on event cascade scope
				StateMatch:      stateMatch, // No state match
				Priority:        priority,   // Priority of the rule
				SuppressionList: suppresses, // List of suppressed rules by this rule
				Action: func(p engine.Processor, m engine.Monitor, e *engine.Event, tid uint64) error { // Action of the rule

					// Create a new root variable scope

					sinkVS := scope.NewScope(fmt.Sprintf("sink: %v", sinkName))

					// Create a new instance state with the monitor - everything called
					// by the rule will have access to the current monitor.

					sinkIs := map[string]interface{}{
						"monitor": m,
					}

					err = sinkVS.SetValue("event", map[interface{}]interface{}{
						"name":  e.Name(),
						"kind":  strings.Join(e.Kind(), engine.RuleKindSeparator),
						"state": e.State(),
					})

					if err == nil {
						scope.SetParentOfScope(sinkVS, vs)

						if _, err = statements.Runtime.Eval(sinkVS, sinkIs, tid); err != nil {

							if sre, ok := err.(*util.RuntimeErrorWithDetail); ok {
								sre.Environment = sinkVS

							} else {

								// Provide additional information for unexpected errors

								err = &util.RuntimeErrorWithDetail{
									RuntimeError: err.(*util.RuntimeError),
									Environment:  sinkVS,
									Data:         nil,
								}
							}
						}
					}

					return err
				},
			}

			if err = rt.erp.Processor.AddRule(rule); err != nil {
				err = rt.erp.NewRuntimeError(util.ErrInvalidState, err.Error(), rt.node)
			}
		}
	}

	return nil, err
}

// Sink child nodes
// ================

/*
sinkDetailRuntime is the runtime for sink detail declarations.
*/
type sinkDetailRuntime struct {
	*baseRuntime
	valType string
}

/*
Eval evaluate this runtime component.
*/
func (rt *sinkDetailRuntime) Eval(vs parser.Scope, is map[string]interface{}, tid uint64) (interface{}, error) {
	var ret interface{}

	_, err := rt.baseRuntime.Eval(vs, is, tid)

	if err == nil {

		if ret, err = rt.node.Children[0].Runtime.Eval(vs, is, tid); err == nil {

			// Check value is of expected type

			if rt.valType == "list" {
				if _, ok := ret.([]interface{}); !ok {
					return nil, rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
						fmt.Sprintf("Expected a list as value"),
						rt.node)
				}

			} else if rt.valType == "map" {

				if _, ok := ret.(map[interface{}]interface{}); !ok {
					return nil, rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
						fmt.Sprintf("Expected a map as value"),
						rt.node)
				}

			} else if rt.valType == "int" {

				if _, ok := ret.(float64); !ok {
					return nil, rt.erp.NewRuntimeError(util.ErrInvalidConstruct,
						fmt.Sprintf("Expected a number as value"),
						rt.node)
				}
			}
		}
	}

	return ret, err
}

/*
kindMatchRuntimeInst returns a new runtime component instance.
*/
func kindMatchRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &sinkDetailRuntime{newBaseRuntime(erp, node), "list"}
}

/*
scopeMatchRuntimeInst returns a new runtime component instance.
*/
func scopeMatchRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &sinkDetailRuntime{newBaseRuntime(erp, node), "list"}
}

/*
stateMatchRuntimeInst returns a new runtime component instance.
*/
func stateMatchRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &sinkDetailRuntime{newBaseRuntime(erp, node), "map"}
}

/*
priorityRuntimeInst returns a new runtime component instance.
*/
func priorityRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &sinkDetailRuntime{newBaseRuntime(erp, node), "int"}
}

/*
suppressesRuntimeInst returns a new runtime component instance.
*/
func suppressesRuntimeInst(erp *ECALRuntimeProvider, node *parser.ASTNode) parser.Runtime {
	return &sinkDetailRuntime{newBaseRuntime(erp, node), "list"}
}
