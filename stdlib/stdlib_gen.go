// Code generated by ecal/stdlib/generate; DO NOT EDIT.

package stdlib

import (
	"fmt"
	"math"
	"reflect"
)

/*
genStdlib contains all generated stdlib constructs.
*/
var genStdlib = map[interface{}]interface{}{
	"math-synopsis": "Mathematics-related constants and functions",
	"math-const":    mathConstMap,
	"math-func":     mathFuncMap,
	"math-func-doc": mathFuncDocMap,
}

/*
mathConstMap contains the mapping of stdlib math constants.
*/
var mathConstMap = map[interface{}]interface{}{
	"E":       float64(math.E),
	"Ln10":    float64(math.Ln10),
	"Ln2":     float64(math.Ln2),
	"Log10E":  float64(math.Log10E),
	"Log2E":   float64(math.Log2E),
	"Phi":     float64(math.Phi),
	"Pi":      float64(math.Pi),
	"Sqrt2":   float64(math.Sqrt2),
	"SqrtE":   float64(math.SqrtE),
	"SqrtPhi": float64(math.SqrtPhi),
	"SqrtPi":  float64(math.SqrtPi),
}

/*
mathFuncDocMap contains the documentation of stdlib math functions.
*/
var mathFuncDocMap = map[interface{}]interface{}{
	"abs":         "Function: abs",
	"acos":        "Function: acos",
	"acosh":       "Function: acosh",
	"asin":        "Function: asin",
	"asinh":       "Function: asinh",
	"atan":        "Function: atan",
	"atan2":       "Function: atan2",
	"atanh":       "Function: atanh",
	"cbrt":        "Function: cbrt",
	"ceil":        "Function: ceil",
	"copysign":    "Function: copysign",
	"cos":         "Function: cos",
	"cosh":        "Function: cosh",
	"dim":         "Function: dim",
	"erf":         "Function: erf",
	"erfc":        "Function: erfc",
	"erfcinv":     "Function: erfcinv",
	"erfinv":      "Function: erfinv",
	"exp":         "Function: exp",
	"exp2":        "Function: exp2",
	"expm1":       "Function: expm1",
	"floor":       "Function: floor",
	"frexp":       "Function: frexp",
	"gamma":       "Function: gamma",
	"hypot":       "Function: hypot",
	"ilogb":       "Function: ilogb",
	"inf":         "Function: inf",
	"isInf":       "Function: isInf",
	"isNaN":       "Function: isNaN",
	"j0":          "Function: j0",
	"j1":          "Function: j1",
	"jn":          "Function: jn",
	"ldexp":       "Function: ldexp",
	"lgamma":      "Function: lgamma",
	"log":         "Function: log",
	"log10":       "Function: log10",
	"log1p":       "Function: log1p",
	"log2":        "Function: log2",
	"logb":        "Function: logb",
	"max":         "Function: max",
	"min":         "Function: min",
	"mod":         "Function: mod",
	"modf":        "Function: modf",
	"naN":         "Function: naN",
	"nextafter":   "Function: nextafter",
	"nextafter32": "Function: nextafter32",
	"pow":         "Function: pow",
	"pow10":       "Function: pow10",
	"remainder":   "Function: remainder",
	"round":       "Function: round",
	"roundToEven": "Function: roundToEven",
	"signbit":     "Function: signbit",
	"sin":         "Function: sin",
	"sincos":      "Function: sincos",
	"sinh":        "Function: sinh",
	"sqrt":        "Function: sqrt",
	"tan":         "Function: tan",
	"tanh":        "Function: tanh",
	"trunc":       "Function: trunc",
	"y0":          "Function: y0",
	"y1":          "Function: y1",
	"yn":          "Function: yn",
}

/*
mathFuncMap contains the mapping of stdlib math functions.
*/
var mathFuncMap = map[interface{}]interface{}{
	"abs":         &ECALFunctionAdapter{reflect.ValueOf(math.Abs), fmt.Sprint(mathFuncDocMap["abs"])},
	"acos":        &ECALFunctionAdapter{reflect.ValueOf(math.Acos), fmt.Sprint(mathFuncDocMap["acos"])},
	"acosh":       &ECALFunctionAdapter{reflect.ValueOf(math.Acosh), fmt.Sprint(mathFuncDocMap["acosh"])},
	"asin":        &ECALFunctionAdapter{reflect.ValueOf(math.Asin), fmt.Sprint(mathFuncDocMap["asin"])},
	"asinh":       &ECALFunctionAdapter{reflect.ValueOf(math.Asinh), fmt.Sprint(mathFuncDocMap["asinh"])},
	"atan":        &ECALFunctionAdapter{reflect.ValueOf(math.Atan), fmt.Sprint(mathFuncDocMap["atan"])},
	"atan2":       &ECALFunctionAdapter{reflect.ValueOf(math.Atan2), fmt.Sprint(mathFuncDocMap["atan2"])},
	"atanh":       &ECALFunctionAdapter{reflect.ValueOf(math.Atanh), fmt.Sprint(mathFuncDocMap["atanh"])},
	"cbrt":        &ECALFunctionAdapter{reflect.ValueOf(math.Cbrt), fmt.Sprint(mathFuncDocMap["cbrt"])},
	"ceil":        &ECALFunctionAdapter{reflect.ValueOf(math.Ceil), fmt.Sprint(mathFuncDocMap["ceil"])},
	"copysign":    &ECALFunctionAdapter{reflect.ValueOf(math.Copysign), fmt.Sprint(mathFuncDocMap["copysign"])},
	"cos":         &ECALFunctionAdapter{reflect.ValueOf(math.Cos), fmt.Sprint(mathFuncDocMap["cos"])},
	"cosh":        &ECALFunctionAdapter{reflect.ValueOf(math.Cosh), fmt.Sprint(mathFuncDocMap["cosh"])},
	"dim":         &ECALFunctionAdapter{reflect.ValueOf(math.Dim), fmt.Sprint(mathFuncDocMap["dim"])},
	"erf":         &ECALFunctionAdapter{reflect.ValueOf(math.Erf), fmt.Sprint(mathFuncDocMap["erf"])},
	"erfc":        &ECALFunctionAdapter{reflect.ValueOf(math.Erfc), fmt.Sprint(mathFuncDocMap["erfc"])},
	"erfcinv":     &ECALFunctionAdapter{reflect.ValueOf(math.Erfcinv), fmt.Sprint(mathFuncDocMap["erfcinv"])},
	"erfinv":      &ECALFunctionAdapter{reflect.ValueOf(math.Erfinv), fmt.Sprint(mathFuncDocMap["erfinv"])},
	"exp":         &ECALFunctionAdapter{reflect.ValueOf(math.Exp), fmt.Sprint(mathFuncDocMap["exp"])},
	"exp2":        &ECALFunctionAdapter{reflect.ValueOf(math.Exp2), fmt.Sprint(mathFuncDocMap["exp2"])},
	"expm1":       &ECALFunctionAdapter{reflect.ValueOf(math.Expm1), fmt.Sprint(mathFuncDocMap["expm1"])},
	"floor":       &ECALFunctionAdapter{reflect.ValueOf(math.Floor), fmt.Sprint(mathFuncDocMap["floor"])},
	"frexp":       &ECALFunctionAdapter{reflect.ValueOf(math.Frexp), fmt.Sprint(mathFuncDocMap["frexp"])},
	"gamma":       &ECALFunctionAdapter{reflect.ValueOf(math.Gamma), fmt.Sprint(mathFuncDocMap["gamma"])},
	"hypot":       &ECALFunctionAdapter{reflect.ValueOf(math.Hypot), fmt.Sprint(mathFuncDocMap["hypot"])},
	"ilogb":       &ECALFunctionAdapter{reflect.ValueOf(math.Ilogb), fmt.Sprint(mathFuncDocMap["ilogb"])},
	"inf":         &ECALFunctionAdapter{reflect.ValueOf(math.Inf), fmt.Sprint(mathFuncDocMap["inf"])},
	"isInf":       &ECALFunctionAdapter{reflect.ValueOf(math.IsInf), fmt.Sprint(mathFuncDocMap["isInf"])},
	"isNaN":       &ECALFunctionAdapter{reflect.ValueOf(math.IsNaN), fmt.Sprint(mathFuncDocMap["isNaN"])},
	"j0":          &ECALFunctionAdapter{reflect.ValueOf(math.J0), fmt.Sprint(mathFuncDocMap["j0"])},
	"j1":          &ECALFunctionAdapter{reflect.ValueOf(math.J1), fmt.Sprint(mathFuncDocMap["j1"])},
	"jn":          &ECALFunctionAdapter{reflect.ValueOf(math.Jn), fmt.Sprint(mathFuncDocMap["jn"])},
	"ldexp":       &ECALFunctionAdapter{reflect.ValueOf(math.Ldexp), fmt.Sprint(mathFuncDocMap["ldexp"])},
	"lgamma":      &ECALFunctionAdapter{reflect.ValueOf(math.Lgamma), fmt.Sprint(mathFuncDocMap["lgamma"])},
	"log":         &ECALFunctionAdapter{reflect.ValueOf(math.Log), fmt.Sprint(mathFuncDocMap["log"])},
	"log10":       &ECALFunctionAdapter{reflect.ValueOf(math.Log10), fmt.Sprint(mathFuncDocMap["log10"])},
	"log1p":       &ECALFunctionAdapter{reflect.ValueOf(math.Log1p), fmt.Sprint(mathFuncDocMap["log1p"])},
	"log2":        &ECALFunctionAdapter{reflect.ValueOf(math.Log2), fmt.Sprint(mathFuncDocMap["log2"])},
	"logb":        &ECALFunctionAdapter{reflect.ValueOf(math.Logb), fmt.Sprint(mathFuncDocMap["logb"])},
	"max":         &ECALFunctionAdapter{reflect.ValueOf(math.Max), fmt.Sprint(mathFuncDocMap["max"])},
	"min":         &ECALFunctionAdapter{reflect.ValueOf(math.Min), fmt.Sprint(mathFuncDocMap["min"])},
	"mod":         &ECALFunctionAdapter{reflect.ValueOf(math.Mod), fmt.Sprint(mathFuncDocMap["mod"])},
	"modf":        &ECALFunctionAdapter{reflect.ValueOf(math.Modf), fmt.Sprint(mathFuncDocMap["modf"])},
	"naN":         &ECALFunctionAdapter{reflect.ValueOf(math.NaN), fmt.Sprint(mathFuncDocMap["naN"])},
	"nextafter":   &ECALFunctionAdapter{reflect.ValueOf(math.Nextafter), fmt.Sprint(mathFuncDocMap["nextafter"])},
	"nextafter32": &ECALFunctionAdapter{reflect.ValueOf(math.Nextafter32), fmt.Sprint(mathFuncDocMap["nextafter32"])},
	"pow":         &ECALFunctionAdapter{reflect.ValueOf(math.Pow), fmt.Sprint(mathFuncDocMap["pow"])},
	"pow10":       &ECALFunctionAdapter{reflect.ValueOf(math.Pow10), fmt.Sprint(mathFuncDocMap["pow10"])},
	"remainder":   &ECALFunctionAdapter{reflect.ValueOf(math.Remainder), fmt.Sprint(mathFuncDocMap["remainder"])},
	"round":       &ECALFunctionAdapter{reflect.ValueOf(math.Round), fmt.Sprint(mathFuncDocMap["round"])},
	"roundToEven": &ECALFunctionAdapter{reflect.ValueOf(math.RoundToEven), fmt.Sprint(mathFuncDocMap["roundToEven"])},
	"signbit":     &ECALFunctionAdapter{reflect.ValueOf(math.Signbit), fmt.Sprint(mathFuncDocMap["signbit"])},
	"sin":         &ECALFunctionAdapter{reflect.ValueOf(math.Sin), fmt.Sprint(mathFuncDocMap["sin"])},
	"sincos":      &ECALFunctionAdapter{reflect.ValueOf(math.Sincos), fmt.Sprint(mathFuncDocMap["sincos"])},
	"sinh":        &ECALFunctionAdapter{reflect.ValueOf(math.Sinh), fmt.Sprint(mathFuncDocMap["sinh"])},
	"sqrt":        &ECALFunctionAdapter{reflect.ValueOf(math.Sqrt), fmt.Sprint(mathFuncDocMap["sqrt"])},
	"tan":         &ECALFunctionAdapter{reflect.ValueOf(math.Tan), fmt.Sprint(mathFuncDocMap["tan"])},
	"tanh":        &ECALFunctionAdapter{reflect.ValueOf(math.Tanh), fmt.Sprint(mathFuncDocMap["tanh"])},
	"trunc":       &ECALFunctionAdapter{reflect.ValueOf(math.Trunc), fmt.Sprint(mathFuncDocMap["trunc"])},
	"y0":          &ECALFunctionAdapter{reflect.ValueOf(math.Y0), fmt.Sprint(mathFuncDocMap["y0"])},
	"y1":          &ECALFunctionAdapter{reflect.ValueOf(math.Y1), fmt.Sprint(mathFuncDocMap["y1"])},
	"yn":          &ECALFunctionAdapter{reflect.ValueOf(math.Yn), fmt.Sprint(mathFuncDocMap["yn"])},
}

// Dummy statement to prevent declared and not used errors
var Dummy = fmt.Sprint(reflect.ValueOf(fmt.Sprint))
