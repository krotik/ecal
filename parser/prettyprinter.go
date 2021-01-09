/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/common/stringutil"
)

/*
IndentationLevel is the level of indentation which the pretty printer should use
*/
const IndentationLevel = 4

/*
Map of AST nodes corresponding to lexer tokens
*/
var prettyPrinterMap map[string]*template.Template

/*
Map of nodes where the precedence might have changed because of parentheses
*/
var bracketPrecedenceMap map[string]bool

func init() {
	prettyPrinterMap = map[string]*template.Template{

		NodeSTRING: template.Must(template.New(NodeSTRING).Parse("{{.qval}}")),
		NodeNUMBER: template.Must(template.New(NodeNUMBER).Parse("{{.val}}")),
		// NodeIDENTIFIER - Special case (handled in code)

		// Constructed tokens

		// NodeSTATEMENTS - Special case (handled in code)
		// NodeFUNCCALL - Special case (handled in code)
		NodeCOMPACCESS + "_1": template.Must(template.New(NodeCOMPACCESS).Parse("[{{.c1}}]")),
		// TokenLIST - Special case (handled in code)
		// TokenMAP - Special case (handled in code)
		// TokenPARAMS - Special case (handled in code)
		NodeGUARD + "_1": template.Must(template.New(NodeGUARD).Parse("{{.c1}}")),

		// Condition operators

		NodeGEQ + "_2": template.Must(template.New(NodeGEQ).Parse("{{.c1}} >= {{.c2}}")),
		NodeLEQ + "_2": template.Must(template.New(NodeLEQ).Parse("{{.c1}} <= {{.c2}}")),
		NodeNEQ + "_2": template.Must(template.New(NodeNEQ).Parse("{{.c1}} != {{.c2}}")),
		NodeEQ + "_2":  template.Must(template.New(NodeEQ).Parse("{{.c1}} == {{.c2}}")),
		NodeGT + "_2":  template.Must(template.New(NodeGT).Parse("{{.c1}} > {{.c2}}")),
		NodeLT + "_2":  template.Must(template.New(NodeLT).Parse("{{.c1}} < {{.c2}}")),

		// Separators

		NodeKVP + "_2":    template.Must(template.New(NodeKVP).Parse("{{.c1}} : {{.c2}}")),
		NodePRESET + "_2": template.Must(template.New(NodePRESET).Parse("{{.c1}}={{.c2}}")),

		// Arithmetic operators

		NodePLUS + "_1":   template.Must(template.New(NodePLUS).Parse("+{{.c1}}")),
		NodePLUS + "_2":   template.Must(template.New(NodePLUS).Parse("{{.c1}} + {{.c2}}")),
		NodeMINUS + "_1":  template.Must(template.New(NodeMINUS).Parse("-{{.c1}}")),
		NodeMINUS + "_2":  template.Must(template.New(NodeMINUS).Parse("{{.c1}} - {{.c2}}")),
		NodeTIMES + "_2":  template.Must(template.New(NodeTIMES).Parse("{{.c1}} * {{.c2}}")),
		NodeDIV + "_2":    template.Must(template.New(NodeDIV).Parse("{{.c1}} / {{.c2}}")),
		NodeMODINT + "_2": template.Must(template.New(NodeMODINT).Parse("{{.c1}} % {{.c2}}")),
		NodeDIVINT + "_2": template.Must(template.New(NodeDIVINT).Parse("{{.c1}} // {{.c2}}")),

		// Assignment statement

		NodeASSIGN + "_2": template.Must(template.New(NodeASSIGN).Parse("{{.c1}} := {{.c2}}")),
		NodeLET + "_1":    template.Must(template.New(NodeASSIGN).Parse("let {{.c1}}")),

		// Import statement

		NodeIMPORT + "_2": template.Must(template.New(NodeIMPORT).Parse("import {{.c1}} as {{.c2}}")),
		NodeAS + "_1":     template.Must(template.New(NodeRETURN).Parse("as {{.c1}}")),

		// Sink definition

		// NodeSINK - Special case (handled in code)
		NodeKINDMATCH + "_1":  template.Must(template.New(NodeKINDMATCH).Parse("kindmatch {{.c1}}")),
		NodeSCOPEMATCH + "_1": template.Must(template.New(NodeSCOPEMATCH).Parse("scopematch {{.c1}}")),
		NodeSTATEMATCH + "_1": template.Must(template.New(NodeSTATEMATCH).Parse("statematch {{.c1}}")),
		NodePRIORITY + "_1":   template.Must(template.New(NodePRIORITY).Parse("priority {{.c1}}")),
		NodeSUPPRESSES + "_1": template.Must(template.New(NodeSUPPRESSES).Parse("suppresses {{.c1}}")),

		// Function definition

		NodeFUNC + "_2":   template.Must(template.New(NodeFUNC).Parse("func {{.c1}} {\n{{.c2}}}")),
		NodeFUNC + "_3":   template.Must(template.New(NodeFUNC).Parse("func {{.c1}}{{.c2}} {\n{{.c3}}}")),
		NodeRETURN:        template.Must(template.New(NodeRETURN).Parse("return")),
		NodeRETURN + "_1": template.Must(template.New(NodeRETURN).Parse("return {{.c1}}")),

		// Boolean operators

		NodeOR + "_2":  template.Must(template.New(NodeOR).Parse("{{.c1}} or {{.c2}}")),
		NodeAND + "_2": template.Must(template.New(NodeAND).Parse("{{.c1}} and {{.c2}}")),
		NodeNOT + "_1": template.Must(template.New(NodeNOT).Parse("not {{.c1}}")),

		// Condition operators

		NodeLIKE + "_2":      template.Must(template.New(NodeLIKE).Parse("{{.c1}} like {{.c2}}")),
		NodeIN + "_2":        template.Must(template.New(NodeIN).Parse("{{.c1}} in {{.c2}}")),
		NodeHASPREFIX + "_2": template.Must(template.New(NodeHASPREFIX).Parse("{{.c1}} hasprefix {{.c2}}")),
		NodeHASSUFFIX + "_2": template.Must(template.New(NodeHASSUFFIX).Parse("{{.c1}} hassuffix {{.c2}}")),
		NodeNOTIN + "_2":     template.Must(template.New(NodeNOTIN).Parse("{{.c1}} notin {{.c2}}")),

		// Constant terminals

		NodeTRUE:  template.Must(template.New(NodeTRUE).Parse("true")),
		NodeFALSE: template.Must(template.New(NodeFALSE).Parse("false")),
		NodeNULL:  template.Must(template.New(NodeNULL).Parse("null")),

		// Conditional statements

		// TokenIF - Special case (handled in code)
		// TokenELIF - Special case (handled in code)
		// TokenELSE - Special case (handled in code)

		// Loop statement

		NodeLOOP + "_2": template.Must(template.New(NodeLOOP).Parse("for {{.c1}} {\n{{.c2}}}")),
		NodeBREAK:       template.Must(template.New(NodeBREAK).Parse("break")),
		NodeCONTINUE:    template.Must(template.New(NodeCONTINUE).Parse("continue")),

		// Try statement

		// TokenTRY - Special case (handled in code)
		// TokenEXCEPT - Special case (handled in code)
		NodeFINALLY + "_1": template.Must(template.New(NodeFINALLY).Parse(" finally {\n{{.c1}}}\n")),

		// Mutex block

		NodeMUTEX + "_2": template.Must(template.New(NodeLOOP).Parse("mutex {{.c1}} {\n{{.c2}}}\n")),
	}

	bracketPrecedenceMap = map[string]bool{
		NodePLUS:  true,
		NodeMINUS: true,
		NodeAND:   true,
		NodeOR:    true,
	}
}

/*
PrettyPrint produces pretty printed code from a given AST.
*/
func PrettyPrint(ast *ASTNode) (string, error) {
	var visit func(ast *ASTNode, path []*ASTNode) (string, error)

	visit = func(ast *ASTNode, path []*ASTNode) (string, error) {
		var buf bytes.Buffer

		if ast == nil {
			return "", fmt.Errorf("Nil pointer in AST")
		}

		numChildren := len(ast.Children)

		tempKey := ast.Name
		tempParam := make(map[string]string)

		// First pretty print children

		if numChildren > 0 {
			for i, child := range ast.Children {
				res, err := visit(child, append(path, child))
				if err != nil {
					return "", err
				}

				if _, ok := bracketPrecedenceMap[child.Name]; ok && ast.binding > child.binding {

					// Put the expression in brackets iff (if and only if) the binding would
					// normally order things differently

					res = fmt.Sprintf("(%v)", res)
				}

				tempParam[fmt.Sprint("c", i+1)] = res
			}

			tempKey += fmt.Sprint("_", len(tempParam))
		}

		if res, ok := ppSpecialDefs(ast, path, tempParam, &buf); ok {
			return res, nil
		} else if res, ok := ppSpecialBlocks(ast, path, tempParam, &buf); ok {
			return res, nil
		} else if res, ok := ppContainerBlocks(ast, path, tempParam, &buf); ok {
			return res, nil
		} else if res, ok := ppSpecialStatements(ast, path, tempParam, &buf); ok {
			return res, nil
		}

		if ast.Token != nil {

			// Adding node value to template parameters

			tempParam["val"] = ast.Token.Val
			tempParam["qval"] = strconv.Quote(ast.Token.Val)
		}

		// Retrieve the template

		temp, ok := prettyPrinterMap[tempKey]
		errorutil.AssertTrue(ok,
			fmt.Sprintf("Could not find template for %v (tempkey: %v)",
				ast.Name, tempKey))

		// Use the children as parameters for template

		errorutil.AssertOk(temp.Execute(&buf, tempParam))

		return ppPostProcessing(ast, path, buf.String()), nil
	}

	res, err := visit(ast, []*ASTNode{ast})

	return strings.TrimSpace(res), err
}

/*
ppPostProcessing applies post processing rules.
*/
func ppPostProcessing(ast *ASTNode, path []*ASTNode, ppString string) string {
	ret := ppString

	// Add meta data

	if len(ast.Meta) > 0 {

		for _, meta := range ast.Meta {
			metaValue := meta.Value()
			if meta.Type() == MetaDataPreComment {
				var buf bytes.Buffer

				scanner := bufio.NewScanner(strings.NewReader(metaValue))
				for scanner.Scan() {
					buf.WriteString(fmt.Sprintf(" %v\n", strings.TrimSpace(scanner.Text())))
				}
				buf.Truncate(buf.Len() - 1) // Remove the last newline

				if strings.Index(buf.String(), "\n") == -1 {
					buf.WriteString(" ")
				}

				ret = fmt.Sprintf("/*%v*/\n%v", buf.String(), ret)

			} else if meta.Type() == MetaDataPostComment {
				metaValue = strings.TrimSpace(strings.ReplaceAll(metaValue, "\n", ""))
				ret = fmt.Sprintf("%v # %v", ret, metaValue)
			}
		}
	}

	// Apply indentation

	if len(path) > 1 {
		if stringutil.IndexOf(ast.Name, []string{
			NodeSTATEMENTS,
			NodeMAP,
			NodeLIST,
			NodeKINDMATCH,
			NodeSTATEMATCH,
			NodeSCOPEMATCH,
			NodePRIORITY,
			NodeSUPPRESSES,
		}) != -1 {
			parent := path[len(path)-2]

			indentSpaces := stringutil.GenerateRollingString(" ", IndentationLevel)
			ret = strings.ReplaceAll(ret, "\n", "\n"+indentSpaces)

			// Add initial indent only if we are inside a block statement

			if stringutil.IndexOf(parent.Name, []string{
				NodeASSIGN,
				NodePRESET,
				NodeKVP,
				NodeLIST,
				NodeFUNCCALL,
				NodeKINDMATCH,
				NodeSTATEMATCH,
				NodeSCOPEMATCH,
				NodePRIORITY,
				NodeSUPPRESSES,
			}) == -1 {
				ret = fmt.Sprintf("%v%v", indentSpaces, ret)
			}

			// Remove indentation from last line unless we have a special case

			if stringutil.IndexOf(parent.Name, []string{
				NodeSINK,
			}) == -1 || ast.Name == NodeSTATEMENTS {

				if idx := strings.LastIndex(ret, "\n"); idx != -1 {
					ret = ret[:idx+1] + ret[idx+IndentationLevel+1:]
				}
			}
		}
	}

	return ret
}

/*
ppSpecialDefs pretty prints special cases.
*/
func ppSpecialDefs(ast *ASTNode, path []*ASTNode, tempParam map[string]string, buf *bytes.Buffer) (string, bool) {
	numChildren := len(ast.Children)

	if ast.Name == NodeFUNCCALL {

		for i := 0; i < numChildren; i++ {
			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
			if i < numChildren-1 {
				buf.WriteString(", ")
			}
		}

		return ppPostProcessing(ast, path, buf.String()), true

	} else if ast.Name == NodeSINK {

		buf.WriteString("sink ")
		buf.WriteString(tempParam["c1"])
		buf.WriteString("\n")

		for i := 1; i < len(ast.Children)-1; i++ {
			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
			buf.WriteString("\n")
		}

		buf.WriteString("{\n")
		buf.WriteString(tempParam[fmt.Sprint("c", len(ast.Children))])
		buf.WriteString("}\n")

		return ppPostProcessing(ast, path, buf.String()), true
	}

	return "", false
}

/*
ppContainerBlocks pretty prints container structures.
*/
func ppContainerBlocks(ast *ASTNode, path []*ASTNode, tempParam map[string]string, buf *bytes.Buffer) (string, bool) {
	numChildren := len(ast.Children)

	if ast.Name == NodeLIST {
		multilineThreshold := 4
		buf.WriteString("[")

		if numChildren > multilineThreshold {
			buf.WriteString("\n")
		}

		for i := 0; i < numChildren; i++ {

			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])

			if i < numChildren-1 {
				if numChildren > multilineThreshold {
					buf.WriteString(",")
				} else {
					buf.WriteString(", ")
				}
			}
			if numChildren > multilineThreshold {
				buf.WriteString("\n")
			}
		}

		buf.WriteString("]")

		return ppPostProcessing(ast, path, buf.String()), true

	} else if ast.Name == NodeMAP {
		multilineThreshold := 2
		buf.WriteString("{")

		if numChildren > multilineThreshold {
			buf.WriteString("\n")
		}

		for i := 0; i < numChildren; i++ {

			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])

			if i < numChildren-1 {
				if numChildren > multilineThreshold {
					buf.WriteString(",")
				} else {
					buf.WriteString(", ")
				}
			}
			if numChildren > multilineThreshold {
				buf.WriteString("\n")
			}
		}

		buf.WriteString("}")

		return ppPostProcessing(ast, path, buf.String()), true

	}

	return "", false
}

/*
ppSpecialBlocks pretty prints special cases.
*/
func ppSpecialBlocks(ast *ASTNode, path []*ASTNode, tempParam map[string]string, buf *bytes.Buffer) (string, bool) {
	numChildren := len(ast.Children)

	// Handle special cases - children in tempParam have been resolved

	if ast.Name == NodeSTATEMENTS {

		// For statements just concat all children

		for i := 0; i < numChildren; i++ {
			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
			buf.WriteString("\n")
		}

		return ppPostProcessing(ast, path, buf.String()), true

	} else if ast.Name == NodeTRY {

		buf.WriteString("try {\n")
		buf.WriteString(tempParam[fmt.Sprint("c1")])

		buf.WriteString("}")

		for i := 1; i < len(ast.Children); i++ {
			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
		}

		return ppPostProcessing(ast, path, buf.String()), true

	} else if ast.Name == NodeEXCEPT {

		buf.WriteString(" except ")

		for i := 0; i < len(ast.Children)-1; i++ {
			buf.WriteString(tempParam[fmt.Sprint("c", i+1)])

			if ast.Children[i+1].Name != NodeAS && i < len(ast.Children)-2 {
				buf.WriteString(",")
			}
			buf.WriteString(" ")
		}

		buf.WriteString("{\n")
		buf.WriteString(tempParam[fmt.Sprint("c", len(ast.Children))])
		buf.WriteString("}")

		return ppPostProcessing(ast, path, buf.String()), true
	}

	return "", false
}

/*
ppSpecialStatements pretty prints special cases.
*/
func ppSpecialStatements(ast *ASTNode, path []*ASTNode, tempParam map[string]string, buf *bytes.Buffer) (string, bool) {
	numChildren := len(ast.Children)

	if ast.Name == NodeIDENTIFIER {

		buf.WriteString(ast.Token.Val)

		for i := 0; i < numChildren; i++ {
			if ast.Children[i].Name == NodeIDENTIFIER {
				buf.WriteString(".")
				buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
			} else if ast.Children[i].Name == NodeFUNCCALL {
				buf.WriteString("(")
				buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
				buf.WriteString(")")
			} else if ast.Children[i].Name == NodeCOMPACCESS {
				buf.WriteString(tempParam[fmt.Sprint("c", i+1)])
			}
		}

		return ppPostProcessing(ast, path, buf.String()), true

	} else if ast.Name == NodePARAMS {

		buf.WriteString("(")
		i := 1
		for ; i < numChildren; i++ {
			buf.WriteString(tempParam[fmt.Sprint("c", i)])
			buf.WriteString(", ")
		}
		buf.WriteString(tempParam[fmt.Sprint("c", i)])
		buf.WriteString(")")

		return ppPostProcessing(ast, path, buf.String()), true

	} else if ast.Name == NodeIF {

		writeGUARD := func(child int) {
			buf.WriteString(tempParam[fmt.Sprint("c", child)])
			buf.WriteString(" {\n")
			buf.WriteString(tempParam[fmt.Sprint("c", child+1)])
			buf.WriteString("}")
		}

		buf.WriteString("if ")

		writeGUARD(1)

		for i := 0; i < len(ast.Children); i += 2 {
			if i+2 == len(ast.Children) && ast.Children[i].Children[0].Name == NodeTRUE {
				buf.WriteString(" else {\n")
				buf.WriteString(tempParam[fmt.Sprint("c", i+2)])
				buf.WriteString("}")
			} else if i > 0 {
				buf.WriteString(" elif ")
				writeGUARD(i + 1)
			}
		}

		return ppPostProcessing(ast, path, buf.String()), true
	}

	return "", false
}
