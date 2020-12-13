/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package stdlib

import (
	"fmt"
	"plugin"
	"reflect"
	"strings"

	"devt.de/krotik/ecal/util"
)

/*
internalStdlibFuncMap holds all registered functions
*/
var internalStdlibFuncMap = make(map[string]util.ECALFunction)

/*
internalStdlibDocMap holds the docstrings for all registered functions
*/
var internalStdlibDocMap = make(map[string]string)

/*
pluginLookup is an interface for required function of the plugin object - only used for unit testing.
*/
type pluginLookup interface {
	Lookup(symName string) (plugin.Symbol, error)
}

/*
pluginTestLookup override plugin object - only used for unit testing.
*/
var pluginTestLookup pluginLookup

/*
AddStdlibPkg adds a package to stdlib. A package needs to be added before functions
can be added.
*/
func AddStdlibPkg(pkg string, docstring string) error {
	_, ok1 := GetPkgDocString(pkg)
	_, ok2 := internalStdlibDocMap[pkg]

	if ok1 || ok2 {
		return fmt.Errorf("Package %v already exists", pkg)
	}

	internalStdlibDocMap[pkg] = docstring

	return nil
}

/*
AddStdlibFunc adds a function to stdlib.
*/
func AddStdlibFunc(pkg string, name string, funcObj util.ECALFunction) error {
	_, ok1 := GetPkgDocString(pkg)
	_, ok2 := internalStdlibDocMap[pkg]

	if !ok1 && !ok2 {
		return fmt.Errorf("Package %v does not exist", pkg)
	}

	internalStdlibFuncMap[fmt.Sprintf("%v.%v", pkg, name)] = funcObj

	return nil
}

/*
LoadStdlibPlugins attempts to load stdlib functions from a given list of definitions.
*/
func LoadStdlibPlugins(jsonObj []interface{}) []error {
	var errs []error

	for _, i := range jsonObj {
		if err := LoadStdlibPlugin(i.(map[string]interface{})); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

/*
LoadStdlibPlugin attempts to load a stdlib function from a given definition.
*/
func LoadStdlibPlugin(jsonObj map[string]interface{}) error {
	pkg := fmt.Sprint(jsonObj["package"])
	name := fmt.Sprint(jsonObj["name"])
	path := fmt.Sprint(jsonObj["path"])
	symName := fmt.Sprint(jsonObj["symbol"])

	return AddStdlibPluginFunc(pkg, name, path, symName)
}

/*
AddStdlibPluginFunc adds a function to stdlib via a loaded plugin.
The plugin needs to be build as a Go plugin (https://golang.org/pkg/plugin):

go build -buildmode=plugin -o myfunc.so myfunc.go

And have an exported variable (passed here as symName) which conforms
to util.ECALPluginFunction.
*/
func AddStdlibPluginFunc(pkg string, name string, path string, symName string) error {
	var err error
	var plug pluginLookup

	AddStdlibPkg(pkg, "Functions provided by plugins")

	if plug, err = plugin.Open(path); err == nil || pluginTestLookup != nil {
		var sym plugin.Symbol

		if pluginTestLookup != nil {
			plug = pluginTestLookup
		}

		if sym, err = plug.Lookup(symName); err == nil {

			if stdlibPluginFunc, ok := sym.(util.ECALPluginFunction); ok {

				adapterFunc := func(a ...interface{}) (interface{}, error) {
					return stdlibPluginFunc.Run(a)
				}

				err = AddStdlibFunc(pkg, name, &ECALFunctionAdapter{
					reflect.ValueOf(adapterFunc), stdlibPluginFunc.DocString()})

			} else {

				err = fmt.Errorf("Symbol %v is not a stdlib function", symName)
			}
		}
	}

	return err
}

/*
GetStdlibSymbols returns all available packages of stdlib and their constant
and function symbols.
*/
func GetStdlibSymbols() ([]string, []string, []string) {
	var constSymbols, funcSymbols []string
	var packageNames []string

	packageSet := make(map[string]bool)

	addSym := func(sym string, suffix string, symMap map[interface{}]interface{},
		ret []string) []string {

		if strings.HasSuffix(sym, suffix) {
			trimSym := strings.TrimSuffix(sym, suffix)
			packageSet[trimSym] = true
			for k := range symMap {
				ret = append(ret, fmt.Sprintf("%v.%v", trimSym, k))
			}
		}

		return ret
	}

	for k, v := range genStdlib {
		sym := fmt.Sprint(k)

		if symMap, ok := v.(map[interface{}]interface{}); ok {
			constSymbols = addSym(sym, "-const", symMap, constSymbols)
			funcSymbols = addSym(sym, "-func", symMap, funcSymbols)
		}
	}

	for k := range packageSet {
		packageNames = append(packageNames, k)
	}

	// Add internal stuff

	for k := range internalStdlibDocMap {
		packageNames = append(packageNames, k)
	}
	for k := range internalStdlibFuncMap {
		funcSymbols = append(funcSymbols, k)
	}

	return packageNames, constSymbols, funcSymbols
}

/*
GetStdlibConst looks up a constant from stdlib.
*/
func GetStdlibConst(name string) (interface{}, bool) {
	var res interface{}
	var resok bool

	if m, n := splitModuleAndName(name); n != "" {
		if cmap, ok := genStdlib[fmt.Sprintf("%v-const", m)]; ok {
			res, resok = cmap.(map[interface{}]interface{})[n]
		}
	}

	return res, resok
}

/*
GetStdlibFunc looks up a function from stdlib.
*/
func GetStdlibFunc(name string) (util.ECALFunction, bool) {
	var res util.ECALFunction
	var resok bool

	if m, n := splitModuleAndName(name); n != "" {
		if fmap, ok := genStdlib[fmt.Sprintf("%v-func", m)]; ok {
			if fn, ok := fmap.(map[interface{}]interface{})[n]; ok {
				res = fn.(util.ECALFunction)
				resok = true
			}
		}
	}

	if !resok {
		res, resok = internalStdlibFuncMap[name]
	}

	return res, resok
}

/*
GetPkgDocString returns the docstring of a stdlib package.
*/
func GetPkgDocString(name string) (string, bool) {
	var res string
	s, ok := genStdlib[fmt.Sprintf("%v-synopsis", name)]
	if ok {
		res = fmt.Sprint(s)
	} else {
		res, ok = internalStdlibDocMap[name]
	}

	return res, ok
}

/*
splitModuleAndName splits up a given full function name in module and function name part.
*/
func splitModuleAndName(fullname string) (string, string) {
	var module, name string

	ccSplit := strings.SplitN(fullname, ".", 2)

	if len(ccSplit) != 0 {
		module = ccSplit[0]
		name = strings.Join(ccSplit[1:], "")
	}

	return module, name
}
