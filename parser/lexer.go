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
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var NamePattern = regexp.MustCompile("^[A-Za-z][A-Za-z0-9]*$")
var NumberPattern = regexp.MustCompile("^[0-9].*$")

/*
LexToken represents a token which is returned by the lexer.
*/
type LexToken struct {
	ID           LexTokenID // Token kind
	Pos          int        // Starting position (in bytes)
	Val          string     // Token value
	Identifier   bool       // Flag if the value is an identifier (not quoted and not a number)
	AllowEscapes bool       // Flag if the value did interpret escape charaters
	Lsource      string     // Input source label (e.g. filename)
	Lline        int        // Line in the input this token appears
	Lpos         int        // Position in the input line this token appears
}

/*
NewLexTokenInstance creates a new LexToken object instance from given LexToken values.
*/
func NewLexTokenInstance(t LexToken) *LexToken {
	return &LexToken{
		t.ID,
		t.Pos,
		t.Val,
		t.Identifier,
		t.AllowEscapes,
		t.Lsource,
		t.Lline,
		t.Lpos,
	}
}

/*
Equals checks if this LexToken equals another LexToken. Returns also a message describing
what is the found difference.
*/
func (t LexToken) Equals(other LexToken, ignorePosition bool) (bool, string) {
	var res = true
	var msg = ""

	if t.ID != other.ID {
		res = false
		msg += fmt.Sprintf("ID is different %v vs %v\n", t.ID, other.ID)
	}

	if !ignorePosition && t.Pos != other.Pos {
		res = false
		msg += fmt.Sprintf("Pos is different %v vs %v\n", t.Pos, other.Pos)
	}

	if t.Val != other.Val {
		res = false
		msg += fmt.Sprintf("Val is different %v vs %v\n", t.Val, other.Val)
	}

	if t.Identifier != other.Identifier {
		res = false
		msg += fmt.Sprintf("Identifier is different %v vs %v\n", t.Identifier, other.Identifier)
	}

	if !ignorePosition && t.Lline != other.Lline {
		res = false
		msg += fmt.Sprintf("Lline is different %v vs %v\n", t.Lline, other.Lline)
	}

	if !ignorePosition && t.Lpos != other.Lpos {
		res = false
		msg += fmt.Sprintf("Lpos is different %v vs %v\n", t.Lpos, other.Lpos)
	}

	if msg != "" {
		var buf bytes.Buffer
		out, _ := json.MarshalIndent(t, "", "  ")
		buf.WriteString(string(out))
		buf.WriteString("\nvs\n")
		out, _ = json.MarshalIndent(other, "", "  ")
		buf.WriteString(string(out))
		msg = fmt.Sprintf("%v%v", msg, buf.String())
	}

	return res, msg
}

/*
PosString returns the position of this token in the origianl input as a string.
*/
func (t LexToken) PosString() string {
	return fmt.Sprintf("Line %v, Pos %v", t.Lline, t.Lpos)
}

/*
String returns a string representation of a token.
*/
func (t LexToken) String() string {

	prefix := ""

	if !t.Identifier {
		prefix = "v:" // Value is not an identifier
	}

	switch {

	case t.ID == TokenEOF:
		return "EOF"

	case t.ID == TokenError:
		return fmt.Sprintf("Error: %s (%s)", t.Val, t.PosString())

	case t.ID == TokenPRECOMMENT:
		return fmt.Sprintf("/* %s */", t.Val)

	case t.ID == TokenPOSTCOMMENT:
		return fmt.Sprintf("# %s", t.Val)

	case t.ID > TOKENodeSYMBOLS && t.ID < TOKENodeKEYWORDS:
		return fmt.Sprintf("%s", strings.ToUpper(t.Val))

	case t.ID > TOKENodeKEYWORDS:
		return fmt.Sprintf("<%s>", strings.ToUpper(t.Val))

	case len(t.Val) > 20:

		// Special case for very long values

		return fmt.Sprintf("%s%.10q...", prefix, t.Val)
	}

	return fmt.Sprintf("%s%q", prefix, t.Val)
}

// Meta data interface

/*
Type returns the meta data type.
*/
func (t LexToken) Type() string {
	if t.ID == TokenPRECOMMENT {
		return MetaDataPreComment
	} else if t.ID == TokenPOSTCOMMENT {
		return MetaDataPostComment
	}
	return MetaDataGeneral
}

/*
Value returns the meta data value.
*/
func (t LexToken) Value() string {
	return t.Val
}

/*
KeywordMap is a map of keywords - these require spaces between them
*/
var KeywordMap = map[string]LexTokenID{

	// Assign statement

	"let": TokenLET,

	// Import statement

	"import": TokenIMPORT,
	"as":     TokenAS,

	// Sink definition

	"sink":       TokenSINK,
	"kindmatch":  TokenKINDMATCH,
	"scopematch": TokenSCOPEMATCH,
	"statematch": TokenSTATEMATCH,
	"priority":   TokenPRIORITY,
	"suppresses": TokenSUPPRESSES,

	// Function definition

	"func":   TokenFUNC,
	"return": TokenRETURN,

	// Boolean operators

	"and": TokenAND,
	"or":  TokenOR,
	"not": TokenNOT,

	// String operators

	"like":      TokenLIKE,
	"hasprefix": TokenHASPREFIX,
	"hassuffix": TokenHASSUFFIX,

	// List operators

	"in":    TokenIN,
	"notin": TokenNOTIN,

	// Constant terminals

	"false": TokenFALSE,
	"true":  TokenTRUE,
	"null":  TokenNULL,

	// Conditional statements

	"if":   TokenIF,
	"elif": TokenELIF,
	"else": TokenELSE,

	// Loop statements

	"for":      TokenFOR,
	"break":    TokenBREAK,
	"continue": TokenCONTINUE,

	// Try block

	"try":     TokenTRY,
	"except":  TokenEXCEPT,
	"finally": TokenFINALLY,
}

/*
SymbolMap is a map of special symbols which will always be unique - these will separate unquoted strings
Symbols can be maximal 2 characters long.
*/
var SymbolMap = map[string]LexTokenID{

	// Condition operators

	">=": TokenGEQ,
	"<=": TokenLEQ,
	"!=": TokenNEQ,
	"==": TokenEQ,
	">":  TokenGT,
	"<":  TokenLT,

	// Grouping symbols

	"(": TokenLPAREN,
	")": TokenRPAREN,
	"[": TokenLBRACK,
	"]": TokenRBRACK,
	"{": TokenLBRACE,
	"}": TokenRBRACE,

	// Separators

	".": TokenDOT,
	",": TokenCOMMA,
	";": TokenSEMICOLON,

	// Grouping

	":": TokenCOLON,
	"=": TokenEQUAL,

	// Arithmetic operators

	"+":  TokenPLUS,
	"-":  TokenMINUS,
	"*":  TokenTIMES,
	"/":  TokenDIV,
	"//": TokenDIVINT,
	"%":  TokenMODINT,

	// Assignment statement

	":=": TokenASSIGN,
}

// Lexer
// =====

/*
RuneEOF is a special rune which represents the end of the input
*/
const RuneEOF = -1

/*
Function which represents the current state of the lexer and returns the next state
*/
type lexFunc func(*lexer) lexFunc

/*
Lexer data structure
*/
type lexer struct {
	name   string        // Name to identify the input
	input  string        // Input string of the lexer
	pos    int           // Current rune pointer
	line   int           // Current line pointer
	lastnl int           // Last newline position
	width  int           // Width of last rune
	start  int           // Start position of the current red token
	tokens chan LexToken // Channel for lexer output
}

/*
Lex lexes a given input. Returns a channel which contains tokens.
*/
func Lex(name string, input string) chan LexToken {
	l := &lexer{name, input, 0, 0, 0, 0, 0, make(chan LexToken)}
	go l.run()
	return l.tokens
}

/*
LexToList lexes a given input. Returns a list of tokens.
*/
func LexToList(name string, input string) []LexToken {
	var tokens []LexToken

	for t := range Lex(name, input) {
		tokens = append(tokens, t)
	}

	return tokens
}

/*
Main loop of the lexer.
*/
func (l *lexer) run() {

	if skipWhiteSpace(l) {
		for state := lexToken; state != nil; {
			state = state(l)

			if !skipWhiteSpace(l) {
				break
			}
		}
	}

	close(l.tokens)
}

/*
next returns the next rune in the input and advances the current rune pointer
if peek is 0. If peek is >0 then the nth character is returned without advancing
the rune pointer.
*/
func (l *lexer) next(peek int) rune {

	// Check if we reached the end

	if int(l.pos) >= len(l.input) {
		return RuneEOF
	}

	// Decode the next rune

	pos := l.pos
	if peek > 0 {
		pos += peek - 1
	}

	r, w := utf8.DecodeRuneInString(l.input[pos:])

	if peek == 0 {
		l.width = w
		l.pos += l.width
	}

	return r
}

/*
backup sets the pointer one rune back. Can only be called once per next call.
*/
func (l *lexer) backup(width int) {
	if width == 0 {
		width = l.width
	}
	l.pos -= width
}

/*
startNew starts a new token.
*/
func (l *lexer) startNew() {
	l.start = l.pos
}

/*
emitToken passes a token back to the client.
*/
func (l *lexer) emitToken(t LexTokenID) {
	if t == TokenEOF {
		l.emitTokenAndValue(t, "", false, false)
		return
	}

	if l.tokens != nil {
		l.tokens <- LexToken{t, l.start, l.input[l.start:l.pos], false, false, l.name,
			l.line + 1, l.start - l.lastnl + 1}
	}
}

/*
emitTokenAndValue passes a token with a given value back to the client.
*/
func (l *lexer) emitTokenAndValue(t LexTokenID, val string, identifier bool, allowEscapes bool) {
	if l.tokens != nil {
		l.tokens <- LexToken{t, l.start, val, identifier, allowEscapes, l.name, l.line + 1, l.start - l.lastnl + 1}
	}
}

/*
emitError passes an error token back to the client.
*/
func (l *lexer) emitError(msg string) {
	if l.tokens != nil {
		l.tokens <- LexToken{TokenError, l.start, msg, false, false, l.name, l.line + 1, l.start - l.lastnl + 1}
	}
}

// Helper functions
// ================

/*
skipWhiteSpace skips any number of whitespace characters. Returns false if the parser
reaches EOF while skipping whitespaces.
*/
func skipWhiteSpace(l *lexer) bool {
	r := l.next(0)

	for unicode.IsSpace(r) || unicode.IsControl(r) || r == RuneEOF {
		if r == '\n' {
			l.line++
			l.lastnl = l.pos
		}
		r = l.next(0)

		if r == RuneEOF {
			l.emitToken(TokenEOF)
			return false
		}
	}

	l.backup(0)
	return true
}

/*
lexTextBlock lexes a block of text without whitespaces. Interprets
optionally all one or two letter tokens.
*/
func lexTextBlock(l *lexer, interpretToken bool) {

	r := l.next(0)

	if interpretToken {

		// Check if we start with a known symbol

		nr := l.next(1)
		if _, ok := SymbolMap[strings.ToLower(string(r)+string(nr))]; ok {
			l.next(0)
			return
		}

		if _, ok := SymbolMap[strings.ToLower(string(r))]; ok {
			return
		}
	}

	for !unicode.IsSpace(r) && !unicode.IsControl(r) && r != RuneEOF {

		if interpretToken {

			// Check if we find a token in the block

			if _, ok := SymbolMap[strings.ToLower(string(r))]; ok {
				l.backup(0)
				return
			}

			nr := l.next(1)
			if _, ok := SymbolMap[strings.ToLower(string(r)+string(nr))]; ok {
				l.backup(0)
				return
			}
		}

		r = l.next(0)
	}

	if r != RuneEOF {
		l.backup(0)
	}
}

/*
lexNumberBlock lexes a block potentially containing a number.
*/
func lexNumberBlock(l *lexer) {

	r := l.next(0)

	for !unicode.IsSpace(r) && !unicode.IsControl(r) && r != RuneEOF {

		if !unicode.IsNumber(r) && r != '.' {
			if r == 'e' {

				l1 := l.next(1)
				l2 := l.next(2)
				if l1 != '+' || !unicode.IsNumber(l2) {
					break
				}
				l.next(0)
				l.next(0)
			} else {
				break
			}
		}
		r = l.next(0)
	}

	if r != RuneEOF {
		l.backup(0)
	}
}

// State functions
// ===============

/*
lexToken is the main entry function for the lexer.
*/
func lexToken(l *lexer) lexFunc {

	// Check if we got a quoted value or a comment

	n1 := l.next(1)
	n2 := l.next(2)

	// Parse comments

	if (n1 == '/' && n2 == '*') || n1 == '#' {
		return lexComment
	}

	// Parse strings

	if (n1 == '"' || n1 == '\'') || (n1 == 'r' && (n2 == '"' || n2 == '\'')) {
		return lexValue
	}

	// Lex a block of text and emit any found tokens

	l.startNew()

	// First try to parse a number

	lexNumberBlock(l)
	identifierCandidate := l.input[l.start:l.pos]
	keywordCandidate := strings.ToLower(identifierCandidate)

	// Check for number

	if NumberPattern.MatchString(keywordCandidate) {
		_, err := strconv.ParseFloat(keywordCandidate, 64)

		if err == nil {
			l.emitTokenAndValue(TokenNUMBER, keywordCandidate, false, false)
			return lexToken
		}
	}

	if len(keywordCandidate) > 0 {
		l.backup(l.pos - l.start)
	}
	lexTextBlock(l, true)
	identifierCandidate = l.input[l.start:l.pos]
	keywordCandidate = strings.ToLower(identifierCandidate)

	// Check for keyword

	token, ok := KeywordMap[keywordCandidate]

	if !ok {

		// Check for symbol

		token, ok = SymbolMap[keywordCandidate]
	}

	if ok {

		// A known token was found

		l.emitToken(token)

	} else {

		if !NamePattern.MatchString(keywordCandidate) {
			l.emitError(fmt.Sprintf("Cannot parse identifier '%v'. Identifies may only contain [a-zA-Z] and [a-zA-Z0-9] from the second character", keywordCandidate))
			return nil
		}

		// An identifier was found

		l.emitTokenAndValue(TokenIDENTIFIER, identifierCandidate, true, false)
	}

	return lexToken
}

/*
lexValue lexes a string value.

Values can be declared in different ways:

' ... ' or " ... "
Characters are parsed between quotes (escape sequences are interpreted)

r' ... ' or r" ... "
Characters are parsed plain between quote
*/
func lexValue(l *lexer) lexFunc {
	var endToken rune

	l.startNew()

	allowEscapes := false

	r := l.next(0)

	// Check if we have a raw quoted string

	if q := l.next(1); r == 'r' && (q == '"' || q == '\'') {
		endToken = q
		l.next(0)
	} else {
		allowEscapes = true
		endToken = r
	}

	r = l.next(0)
	rprev := ' '
	lLine := l.line
	lLastnl := l.lastnl

	for (!allowEscapes && r != endToken) ||
		(allowEscapes && (r != endToken || rprev == '\\')) {

		if r == '\n' {
			lLine++
			lLastnl = l.pos
		}
		rprev = r
		r = l.next(0)

		if r == RuneEOF {
			l.emitError("Unexpected end while reading string value (unclosed quotes)")
			return nil
		}
	}

	if allowEscapes {
		val := l.input[l.start+1 : l.pos-1]

		// Interpret escape sequences right away

		if endToken == '\'' {

			// Escape double quotes in a single quoted string

			val = strings.Replace(val, "\"", "\\\"", -1)
		}

		s, err := strconv.Unquote("\"" + val + "\"")
		if err != nil {
			l.emitError(err.Error() + " while parsing string")
			return nil
		}

		l.emitTokenAndValue(TokenSTRING, s, false, true)

	} else {

		l.emitTokenAndValue(TokenSTRING, l.input[l.start+2:l.pos-1], false, false)
	}

	//  Set newline

	l.line = lLine
	l.lastnl = lLastnl

	return lexToken
}

/*
lexComment lexes comments.
*/
func lexComment(l *lexer) lexFunc {

	// Consume initial /*

	r := l.next(0)

	if r == '#' {

		l.startNew()

		for r != '\n' && r != RuneEOF {
			r = l.next(0)
		}

		l.emitTokenAndValue(TokenPOSTCOMMENT, l.input[l.start:l.pos], false, false)

		if r == RuneEOF {
			return nil
		}

		l.line++

	} else {

		l.next(0)

		lLine := l.line
		lLastnl := l.lastnl

		l.startNew()

		r = l.next(0)

		for r != '*' || l.next(1) != '/' {

			if r == '\n' {
				lLine++
				lLastnl = l.pos
			}
			r = l.next(0)

			if r == RuneEOF {
				l.emitError("Unexpected end while reading comment")
				return nil
			}
		}

		l.emitTokenAndValue(TokenPRECOMMENT, l.input[l.start:l.pos-1], false, false)

		// Consume final /

		l.next(0)

		//  Set newline

		l.line = lLine
		l.lastnl = lLastnl

	}

	return lexToken
}
