/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/importer"
	goparser "go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"unicode"

	"github.com/krotik/common/errorutil"
	"github.com/krotik/common/stringutil"
)

//go:generate echo Generating ECAL stdlib from Go functions ...
//go:generate go run github.com/krotik/ecal/stdlib/generate $PWD/stdlib

/*
Stdlib candidates modules:

go list std | grep -v internal | grep -v '\.' | grep -v unsafe | grep -v syscall
*/

// List underneath the Go symbols which should be available in ECAL's stdlib
// e.g. 	var pkgNames = map[string][]string{ "fmt":  {"Println", "Sprint"} }

// Turn the generateDoc switch on or off to extract documentation from the
// Go source.

// =============EDIT HERE START=============

var pkgNames = map[string][]string{
	//	"fmt":  {"Println", "Sprint"},
}

var generateDoc = false

// ==============EDIT HERE END==============

var filename = filepath.Join(os.Args[1], "stdlib_gen.go")

var stderrPrint = fmt.Println
var stdoutPrint = fmt.Println

func main() {
	var err error
	var outbuf bytes.Buffer

	synopsis := make(map[string]string)
	pkgDocs := make(map[string]*doc.Package)

	flag.Parse()

	// Make sure we have at least an empty pkgName

	if len(pkgNames) == 0 {
		pkgNames["math"] = []string{
			"E",
			"Pi",
			"Phi",
			"Sqrt2",
			"SqrtE",
			"SqrtPi",
			"SqrtPhi",
			"Ln2",
			"Log2E",
			"Ln10",
			"Log10E",

			"Abs",
			"Acos",
			"Acosh",
			"Asin",
			"Asinh",
			"Atan",
			"Atan2",
			"Atanh",
			"Cbrt",
			"Ceil",
			"Copysign",
			"Cos",
			"Cosh",
			"Dim",
			"Erf",
			"Erfc",
			"Erfcinv",
			"Erfinv",
			"Exp",
			"Exp2",
			"Expm1",
			"Floor",
			"Frexp",
			"Gamma",
			"Hypot",
			"Ilogb",
			"Inf",
			"IsInf",
			"IsNaN",
			"J0",
			"J1",
			"Jn",
			"Ldexp",
			"Lgamma",
			"Log",
			"Log10",
			"Log1p",
			"Log2",
			"Logb",
			"Max",
			"Min",
			"Mod",
			"Modf",
			"NaN",
			"Nextafter",
			"Nextafter32",
			"Pow",
			"Pow10",
			"Remainder",
			"Round",
			"RoundToEven",
			"Signbit",
			"Sin",
			"Sincos",
			"Sinh",
			"Sqrt",
			"Tan",
			"Tanh",
			"Trunc",
			"Y0",
			"Y1",
			"Yn",
		}
	}

	// Make sure pkgNames is sorted

	var importList []string
	for pkgName, names := range pkgNames {
		sort.Strings(names)
		importList = append(importList, pkgName)
		synopsis["math"] = "Mathematics-related constants and functions"
	}
	sort.Strings(importList)

	outbuf.WriteString(`
// Code generated by ecal/stdlib/generate; DO NOT EDIT.

package stdlib

`)

	outbuf.WriteString("import (\n")
	for _, pkgName := range importList {

		if generateDoc {
			syn, pkgDoc, err := getPackageDocs(pkgName)
			errorutil.AssertOk(err) // If this throws try not generating the docs!
			synopsis[pkgName] = syn
			pkgDocs[pkgName] = pkgDoc
		} else if _, ok := synopsis[pkgName]; !ok {
			synopsis[pkgName] = fmt.Sprintf("Package %v", pkgName)
		}

		outbuf.WriteString(fmt.Sprintf("\t\"%v\"\n", pkgName))
	}

	if stringutil.IndexOf("fmt", importList) == -1 {
		outbuf.WriteString(`	"fmt"
`)
	}

	if stringutil.IndexOf("reflect", importList) == -1 {
		outbuf.WriteString(`	"reflect"
)

`)
	}

	outbuf.WriteString(`/*
genStdlib contains all generated stdlib constructs.
*/
`)
	outbuf.WriteString("var genStdlib = map[interface{}]interface{}{\n")
	for _, pkgName := range importList {
		if s, ok := synopsis[pkgName]; ok {
			outbuf.WriteString(fmt.Sprintf("\t\"%v-synopsis\" : %#v,\n", pkgName, s))
		}
		outbuf.WriteString(fmt.Sprintf("\t\"%v-const\" : %vConstMap,\n", pkgName, pkgName))
		outbuf.WriteString(fmt.Sprintf("\t\"%v-func\" : %vFuncMap,\n", pkgName, pkgName))
		outbuf.WriteString(fmt.Sprintf("\t\"%v-func-doc\" : %vFuncDocMap,\n", pkgName, pkgName))
	}
	outbuf.WriteString("}\n\n")

	for _, pkgName := range importList {
		var pkg *types.Package

		pkgSymbols := pkgNames[pkgName]

		if err == nil {

			pkg, err = importer.ForCompiler(fset, "source", nil).Import(pkgName)

			if err == nil {
				stdoutPrint("Generating adapter functions for", pkg)

				scope := pkg.Scope()

				// Write constants

				writeConstants(&outbuf, pkgName, pkgSymbols, scope)

				// Write function documentation

				writeFuncDoc(&outbuf, pkgName, pkgDocs, pkgSymbols, scope)

				// Write functions

				writeFuncs(&outbuf, pkgName, pkgSymbols, scope)
			}
		}
	}

	// Write dummy statement
	outbuf.WriteString("// Dummy statement to prevent declared and not used errors\n")
	outbuf.WriteString("var Dummy = fmt.Sprint(reflect.ValueOf(fmt.Sprint))\n\n")

	if err == nil {
		err = ioutil.WriteFile(filename, outbuf.Bytes(), 0644)
	}

	if err != nil {
		stderrPrint("Error:", err)
	}
}

var (
	fset = token.NewFileSet()
	ctx  = &build.Default
)

/*
writeConstants writes out all stdlib constant definitions.
*/
func writeConstants(outbuf *bytes.Buffer, pkgName string, pkgSymbols []string, scope *types.Scope) {

	outbuf.WriteString(fmt.Sprintf(`/*
%vConstMap contains the mapping of stdlib %v constants.
*/
var %vConstMap = map[interface{}]interface{}{
`, pkgName, pkgName, pkgName))

	for _, name := range scope.Names() {

		if !containsSymbol(pkgSymbols, name) {
			continue
		}

		switch obj := scope.Lookup(name).(type) {
		case *types.Const:

			if unicode.IsUpper([]rune(name)[0]) {

				line := fmt.Sprintf(`	"%v": %v.%v,
`, name, pkgName, obj.Name())

				if basicType, ok := obj.Type().(*types.Basic); ok {

					// Convert number constants so they can be used in calculations

					switch basicType.Kind() {
					case types.Int,
						types.Int8,
						types.Int16,
						types.Int32,
						types.Int64,
						types.Uint,
						types.Uint8,
						types.Uint16,
						types.Uint32,
						types.Uint64,
						types.Uintptr,
						types.Float32,
						types.UntypedInt,
						types.UntypedFloat:

						line = fmt.Sprintf(`	"%v": float64(%v.%v),
`, name, pkgName, obj.Name())
					}
				}

				outbuf.WriteString(line)
			}
		}
	}

	outbuf.WriteString("}\n\n")
}

/*
writeFuncDoc writes out all stdlib function documentation.
*/
func writeFuncDoc(outbuf *bytes.Buffer, pkgName string, pkgDocs map[string]*doc.Package,
	pkgSymbols []string, scope *types.Scope) {

	outbuf.WriteString(fmt.Sprintf(`/*
%vFuncDocMap contains the documentation of stdlib %v functions.
*/
var %vFuncDocMap = map[interface{}]interface{}{
`, pkgName, pkgName, pkgName))

	if pkgDoc, ok := pkgDocs[pkgName]; ok {

		for _, name := range scope.Names() {

			if !containsSymbol(pkgSymbols, name) {
				continue
			}

			for _, f := range pkgDoc.Funcs {
				if f.Name == name {
					outbuf.WriteString(
						fmt.Sprintf(`	"%v": %#v,
`, name, f.Doc))
				}
			}
		}

	} else {

		for _, name := range pkgSymbols {
			switch scope.Lookup(name).(type) {
			case *types.Func:
				outbuf.WriteString(
					fmt.Sprintf(`	"%v": "Function: %v",
`, lcFirst(name), lcFirst(name)))
			}
		}
	}

	outbuf.WriteString("}\n\n")
}

/*
writeFuncs writes out all stdlib function definitions.
*/
func writeFuncs(outbuf *bytes.Buffer, pkgName string, pkgSymbols []string, scope *types.Scope) {
	outbuf.WriteString(fmt.Sprintf(`/*
%vFuncMap contains the mapping of stdlib %v functions.
*/
var %vFuncMap = map[interface{}]interface{}{
`, pkgName, pkgName, pkgName))

	for _, name := range scope.Names() {

		if !containsSymbol(pkgSymbols, name) {
			continue
		}

		switch obj := scope.Lookup(name).(type) {
		case *types.Func:
			if unicode.IsUpper([]rune(name)[0]) {
				outbuf.WriteString(
					fmt.Sprintf(`	%#v: &ECALFunctionAdapter{reflect.ValueOf(%v), fmt.Sprint(%vFuncDocMap[%#v])},
`, lcFirst(name), obj.FullName(), pkgName, lcFirst(name)))
			}
		}
	}

	outbuf.WriteString("}\n\n")
}

/*
getPackageDocs returns the source code documentation of as given Go package.
Returns a short synopsis and a documentation object.
*/
func getPackageDocs(pkgName string) (string, *doc.Package, error) {
	var synopsis string
	var pkgDoc *doc.Package
	var filenames []string

	bp, err := ctx.Import(pkgName, ".", 0)

	if err == nil {

		synopsis = bp.Doc

		// Get all go files of the package

		filenames = append(filenames, bp.GoFiles...)
		filenames = append(filenames, bp.CgoFiles...)

		// Build the ast package from Go source

		astPkg := &ast.Package{
			Name:  bp.Name,
			Files: make(map[string]*ast.File),
		}

		for _, filename := range filenames {
			filepath := filepath.Join(bp.Dir, filename)
			astFile, _ := goparser.ParseFile(fset, filepath, nil, goparser.ParseComments)
			astPkg.Files[filepath] = astFile
		}

		// Build the package doc object

		pkgDoc = doc.New(astPkg, bp.Dir, doc.AllDecls)
	}

	return synopsis, pkgDoc, err
}

/*
containsSymbol checks if a list of strings contains a given item.
*/
func containsSymbol(symbols []string, item string) bool {
	i := sort.SearchStrings(symbols, item)
	return i < len(symbols) && symbols[i] == item
}

/*
lcFirst lower cases the first rune of a given string
*/
func lcFirst(s string) string {
	var ret = ""
	for i, v := range s {
		ret = string(unicode.ToLower(v)) + s[i+len(string(v)):]
		break
	}
	return ret
}
