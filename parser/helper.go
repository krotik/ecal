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
	"fmt"
	"strconv"

	"devt.de/krotik/common/datautil"
	"devt.de/krotik/common/stringutil"
)

// AST Nodes
// =========

/*
MetaData is auxiliary data which can be attached to ASTs.
*/
type MetaData interface {

	/*
		Type returns the type of the meta data.
	*/
	Type() string

	/*
		Value returns the value of the meta data.
	*/
	Value() string
}

/*
metaData is a minimal MetaData implementation.
*/
type metaData struct {
	metatype  string
	metavalue string
}

/*
Type returns the type of the meta data.
*/
func (m *metaData) Type() string {
	return m.metatype
}

/*
Value returns the value of the meta data.
*/
func (m *metaData) Value() string {
	return m.metavalue
}

/*
ASTNode models a node in the AST
*/
type ASTNode struct {
	Name     string     // Name of the node
	Token    *LexToken  // Lexer token of this ASTNode
	Meta     []MetaData // Meta data for this ASTNode (e.g. comments)
	Children []*ASTNode // Child nodes
	Runtime  Runtime    // Runtime component for this ASTNode

	binding        int                                                             // Binding power of this node
	nullDenotation func(p *parser, self *ASTNode) (*ASTNode, error)                // Configure token as beginning node
	leftDenotation func(p *parser, self *ASTNode, left *ASTNode) (*ASTNode, error) // Configure token as left node
}

/*
Create a new instance of this ASTNode which is connected to a concrete lexer token.
*/
func (n *ASTNode) instance(p *parser, t *LexToken) *ASTNode {

	ret := &ASTNode{n.Name, t, nil, make([]*ASTNode, 0, 2), nil, n.binding, n.nullDenotation, n.leftDenotation}

	if p.rp != nil {
		ret.Runtime = p.rp.Runtime(ret)
	}

	return ret
}

/*
Equals checks if this AST data equals another AST data. Returns also a message describing
what is the found difference.
*/
func (n *ASTNode) Equals(other *ASTNode, ignoreTokenPosition bool) (bool, string) {
	return n.equalsPath(n.Name, other, ignoreTokenPosition)
}

/*
equalsPath checks if this AST data equals another AST data while preserving the search path.
Returns also a message describing what is the found difference.
*/
func (n *ASTNode) equalsPath(path string, other *ASTNode, ignoreTokenPosition bool) (bool, string) {
	var res = true
	var msg = ""

	if n.Name != other.Name {
		res = false
		msg = fmt.Sprintf("Name is different %v vs %v\n", n.Name, other.Name)
	}

	if n.Token != nil && other.Token != nil {
		if ok, tokenMSG := n.Token.Equals(*other.Token, ignoreTokenPosition); !ok {
			res = false
			msg += fmt.Sprintf("Token is different:\n%v\n", tokenMSG)
		}
	}

	if len(n.Meta) != len(other.Meta) {
		res = false
		msg = fmt.Sprintf("Number of meta data entries is different %v vs %v\n",
			len(n.Meta), len(other.Meta))
	} else {
		for i, meta := range n.Meta {

			// Check for different in meta entries

			if meta.Type() != other.Meta[i].Type() {
				res = false
				msg += fmt.Sprintf("Meta data type is different %v vs %v\n", meta.Type(), other.Meta[i].Type())
			} else if meta.Value() != other.Meta[i].Value() {
				res = false
				msg += fmt.Sprintf("Meta data value is different %v vs %v\n", meta.Value(), other.Meta[i].Value())
			}
		}
	}

	if len(n.Children) != len(other.Children) {
		res = false
		msg = fmt.Sprintf("Number of children is different %v vs %v\n",
			len(n.Children), len(other.Children))
	} else {
		for i, child := range n.Children {

			// Check for different in children

			if ok, childMSG := child.equalsPath(fmt.Sprintf("%v > %v", path, child.Name),
				other.Children[i], ignoreTokenPosition); !ok {
				return ok, childMSG
			}
		}
	}

	if msg != "" {
		var buf bytes.Buffer
		buf.WriteString("AST Nodes:\n")
		n.levelString(0, &buf, 1)
		buf.WriteString("vs\n")
		other.levelString(0, &buf, 1)
		msg = fmt.Sprintf("Path to difference: %v\n\n%v\n%v", path, msg, buf.String())
	}

	return res, msg
}

/*
String returns a string representation of this token.
*/
func (n *ASTNode) String() string {
	var buf bytes.Buffer
	n.levelString(0, &buf, -1)
	return buf.String()
}

/*
levelString function to recursively print the tree.
*/
func (n *ASTNode) levelString(indent int, buf *bytes.Buffer, printChildren int) {

	// Print current level

	buf.WriteString(stringutil.GenerateRollingString(" ", indent*2))

	if n.Name == NodeSTRING {
		buf.WriteString(fmt.Sprintf("%v: '%v'", n.Name, n.Token.Val))
	} else if n.Name == NodeNUMBER {
		buf.WriteString(fmt.Sprintf("%v: %v", n.Name, n.Token.Val))
	} else if n.Name == NodeIDENTIFIER {
		buf.WriteString(fmt.Sprintf("%v: %v", n.Name, n.Token.Val))
	} else {
		buf.WriteString(n.Name)
	}

	if len(n.Meta) > 0 {
		buf.WriteString(" # ")
		for i, c := range n.Meta {
			buf.WriteString(c.Value())
			if i < len(n.Meta)-1 {
				buf.WriteString(" ")
			}
		}
	}

	buf.WriteString("\n")

	if printChildren == -1 || printChildren > 0 {

		if printChildren != -1 {
			printChildren--
		}

		// Print children

		for _, child := range n.Children {
			child.levelString(indent+1, buf, printChildren)
		}
	}
}

/*
ToJSONObject returns this ASTNode and all its children as a JSON object.
*/
func (n *ASTNode) ToJSONObject() map[string]interface{} {
	ret := make(map[string]interface{})

	ret["name"] = n.Name

	lenMeta := len(n.Meta)

	if lenMeta > 0 {
		meta := make([]map[string]interface{}, lenMeta)
		for i, metaChild := range n.Meta {
			meta[i] = map[string]interface{}{
				"type":  metaChild.Type(),
				"value": metaChild.Value(),
			}
		}

		ret["meta"] = meta
	}

	lenChildren := len(n.Children)

	if lenChildren > 0 {
		children := make([]map[string]interface{}, lenChildren)
		for i, child := range n.Children {
			children[i] = child.ToJSONObject()
		}

		ret["children"] = children
	}

	// The value is what the lexer found in the source

	if n.Token != nil {
		ret["id"] = n.Token.ID
		if n.Token.Val != "" {
			ret["value"] = n.Token.Val
		}
		ret["identifier"] = n.Token.Identifier
		ret["allowescapes"] = n.Token.AllowEscapes
		ret["pos"] = n.Token.Pos
		ret["source"] = n.Token.Lsource
		ret["line"] = n.Token.Lline
		ret["linepos"] = n.Token.Lpos
	}

	return ret
}

/*
ASTFromJSONObject creates an AST from a JSON Object.
The following nested map structure is expected:

	{
		name     : <name of node>

		// Optional node information
		value    : <value of node>
		children : [ <child nodes> ]

		// Optional token information
		id       : <token id>
	}
*/
func ASTFromJSONObject(jsonAST map[string]interface{}) (*ASTNode, error) {
	var astMeta []MetaData
	var astChildren []*ASTNode

	nodeID := TokenANY

	name, ok := jsonAST["name"]
	if !ok {
		return nil, fmt.Errorf("Found json ast node without a name: %v", jsonAST)
	}

	if nodeIDString, ok := jsonAST["id"]; ok {
		if nodeIDInt, err := strconv.Atoi(fmt.Sprint(nodeIDString)); err == nil && IsValidTokenID(nodeIDInt) {
			nodeID = LexTokenID(nodeIDInt)
		}
	}

	getVal := func(k string, d interface{}) (interface{}, int) {
		value, ok := jsonAST[k]
		if !ok {
			value = d
		}
		numVal, _ := strconv.Atoi(fmt.Sprint(value))
		return value, numVal
	}

	value, _ := getVal("value", "")
	identifier, _ := getVal("identifier", false)
	allowescapes, _ := getVal("allowescapes", false)
	_, prefixnl := getVal("prefixnewlines", "")
	_, pos := getVal("pos", "")
	_, line := getVal("line", "")
	_, linepos := getVal("linepos", "")
	source, _ := getVal("source", "")

	// Create meta data

	if meta, ok := jsonAST["meta"]; ok {

		if ic, ok := meta.([]interface{}); ok {

			// Do a list conversion if necessary - this is necessary when we parse
			// JSON with map[string]interface{}

			metaList := make([]map[string]interface{}, len(ic))
			for i := range ic {
				metaList[i] = ic[i].(map[string]interface{})
			}

			meta = metaList
		}

		for _, metaChild := range meta.([]map[string]interface{}) {
			astMeta = append(astMeta, &metaData{
				fmt.Sprint(metaChild["type"]), fmt.Sprint(metaChild["value"])})
		}
	}

	// Create children

	if children, ok := jsonAST["children"]; ok {

		if ic, ok := children.([]interface{}); ok {

			// Do a list conversion if necessary - this is necessary when we parse
			// JSON with map[string]interface{}

			childrenList := make([]map[string]interface{}, len(ic))
			for i := range ic {
				childrenList[i] = ic[i].(map[string]interface{})
			}

			children = childrenList
		}

		for _, child := range children.([]map[string]interface{}) {

			astChild, err := ASTFromJSONObject(child)
			if err != nil {
				return nil, err
			}

			astChildren = append(astChildren, astChild)
		}
	}

	token := &LexToken{
		nodeID,               // ID
		pos,                  // Pos
		fmt.Sprint(value),    // Val
		identifier == true,   // Identifier
		allowescapes == true, // AllowEscapes
		prefixnl,             // PrefixNewlines
		fmt.Sprint(source),   // Lsource
		line,                 // Lline
		linepos,              // Lpos
	}

	return &ASTNode{fmt.Sprint(name), token, astMeta, astChildren, nil, 0, nil, nil}, nil
}

// Look ahead buffer
// =================

/*
LABuffer models a look-ahead buffer.
*/
type LABuffer struct {
	tokens chan LexToken
	buffer *datautil.RingBuffer
}

/*
NewLABuffer creates a new NewLABuffer instance.
*/
func NewLABuffer(c chan LexToken, size int) *LABuffer {

	if size < 1 {
		size = 1
	}

	ret := &LABuffer{c, datautil.NewRingBuffer(size)}

	v, more := <-ret.tokens
	ret.buffer.Add(v)

	for ret.buffer.Size() < size && more && v.ID != TokenEOF {
		v, more = <-ret.tokens
		ret.buffer.Add(v)
	}

	return ret
}

/*
Next returns the next item.
*/
func (b *LABuffer) Next() (LexToken, bool) {

	ret := b.buffer.Poll()

	if v, more := <-b.tokens; more {
		b.buffer.Add(v)
	}

	if ret == nil {
		return LexToken{ID: TokenEOF}, false
	}

	return ret.(LexToken), true
}

/*
Peek looks inside the buffer starting with 0 as the next item.
*/
func (b *LABuffer) Peek(pos int) (LexToken, bool) {

	if pos >= b.buffer.Size() {
		return LexToken{ID: TokenEOF}, false
	}

	return b.buffer.Get(pos).(LexToken), true
}
