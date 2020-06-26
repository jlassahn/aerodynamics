
package main

/*

tokens:
 whitespace is spaces, tabs, linesfeeds, comments
 comment is # ... EOL
 Alpha followed by any alphanumeric or _
 digit followed by digits and .
 any single nonalphanumeric


Grammar

File:
	*Assignment  # one or more

Assignment:
	Value
	Definition

Value:
	TOKEN : Expression

Definition:
	DefType DefName  { *Assignment }   # zero or more

DefType:
	TOKEN  # Tube, Sheet, Cap, Mount, ...

DefName:
	Name
	Name [ DefIndexList ]

Name:
	TOKEN

DefIndexList:
	DefIndex
	DefIndex , DefIndexList

DefIndex:
	TOKEN ~ Expression

Expression:
	Expression1
	Expression + Expression1
	Expression - Expression1

Expresion1:
	Expression2
	Expression1 * Expression2
	Expression1 / Expression2

Expression2:
	Expression3
	 - Expression2

Expression3:
	NUMBER
	NameRef
	( Expression )

NameRef:
	TOKEN
	TOKEN ( ArgList )  # built-in function
	TOKEN [ ArgList ]
	TOKEN . NameRef
	TOKEN [ ArgList ] . NameRef

ArgList:
	Expression
	Expression , ArgList
*/

import (
	"io"
	"os"
	"fmt"
	"strconv"
)

const (
	TYPE_EOF = 0
	TYPE_SYMBOL = 1
	TYPE_NUMBER = 2
	TYPE_OTHER = 3
	TYPE_OBJECT = 4
	TYPE_ARRAY = 5
)

type Parser struct {
	fp io.Reader
	charBuf []byte
	charIndex int
	charLength int
	isEOF bool
	lineCount int
	tokenPushback Token

	ObjectRoot *ParseObject
	ObjectStack []string
	IndexNameStack []string
	IndexValueStack []int
}

type Token struct {
	Text string
	Type int
}

type ParseObject struct {
	ObjectType string
	Values map[string]ParseValue
	Definitions map[string]*ParseObjectArray
}

func (obj *ParseObject) LookupValue(name string) ParseValue {
	v := obj.Values[name]
	if v != nil {
		return v
	}

	d := obj.Definitions[name]
	if d == nil {
		fmt.Printf("FIXME no value found for %v\n", name)
		fmt.Println(obj.Definitions)
		return nil
	}

	if d.Template.IndexNames[0] == "" {
		return ObjectValue { &d.Children[0] }
	} else {
		return ArrayValue { d }
	}
}

type ParseObjectArray struct {
	Children []ParseObject
	Template ObjectArrayTemplate
}

type ObjectArrayTemplate struct {
	ObjectType string
	IndexSizes []int
	IndexNames []string
}

func (tp *ObjectArrayTemplate) CreateObjects(parser *Parser, object *ParseObject, sym Token) {

	// FIXME check that index symbols aren't duplicates or in the symbol table
	n := 1
	for _,x := range tp.IndexSizes {
		n = n*x
	}

	array := ParseObjectArray{}
	array.Template = *tp
	array.Children = make([]ParseObject, n)
	for i:=0; i<n; i++ {
		array.Children[i].ObjectType = tp.ObjectType
		array.Children[i].Values = map[string]ParseValue{}
		array.Children[i].Definitions = map[string]*ParseObjectArray{}

	}
	object.Definitions[sym.Text] = &array
}

type ParseValue interface {
	Type() int
	Number() float32
	Symbol() string
	Object() *ParseObject
	Array() *ParseObjectArray
}

type Evaluator interface {
	Type() int
	Evaluate(parser *Parser) ParseValue
	AssignTo(parser *Parser, object *ParseObject, sym Token)
}

func NewParser(name string) (*Parser, error) {

	fp,err := os.Open(name)
	if err != nil {
		return nil, err
	}

	parser := Parser{}
	parser.fp = fp
	parser.charBuf = make([]byte, 8)
	parser.charIndex = 0
	parser.charLength = 0
	parser.isEOF = false
	parser.lineCount = 1

	return &parser, nil
}

func (parser *Parser) getChar() byte {

	if parser.charIndex >= parser.charLength && !parser.isEOF {
		n, err := parser.fp.Read(parser.charBuf)
		parser.charLength = n
		parser.charIndex = 0
		if err != nil {
			parser.isEOF = true
		}
	}

	if parser.charIndex < parser.charLength {
		ret := parser.charBuf[parser.charIndex]
		parser.charIndex ++
		if ret == '\n' {
			parser.lineCount ++
		}
		return ret
	} else {
		return 0
	}

}

func (parser *Parser) ungetChar(c byte) {
	if parser.charIndex == 0 {
		panic("invalid ungetChar")
	}
	parser.charIndex --
	parser.charBuf[parser.charIndex] = c
	if c == '\n' {
		parser.lineCount --
	}
}

func (parser *Parser) skipSpace() {

	for {
		c := parser.getChar()
		if c == '#' {
			for {
				c = parser.getChar()
				if c == 0 {
					return
				}
				if c == '\n' {
					break
				}
			}
		} else if (c==' ') || (c=='\t') || (c=='\n') {
		} else if c==0 {
			return
		} else {
			parser.ungetChar(c)
			return
		}
	}
}

func (parser *Parser) UngetToken(t Token) {

	if parser.tokenPushback.Type != TYPE_EOF {
		panic("invalid UngetToken")
	}

	parser.tokenPushback = t
}

func (parser *Parser) MatchToken(txt string) bool {
	t := parser.GetToken()
	if t.Text != txt {
		parser.UngetToken(t)
		return false
	}
	return true
}

func (parser *Parser) GetToken() Token {

	if parser.tokenPushback.Type != TYPE_EOF {
		ret := parser.tokenPushback
		parser.tokenPushback.Type = TYPE_EOF
		return ret
	}

	parser.skipSpace()

	c := parser.getChar()
	if c==0 {
		return Token{ "", TYPE_EOF}
	}

	if isDigit(c) {
		data := []byte{c}
		for {
			c := parser.getChar()
			if isDigit(c) || c == '.' {
				data = append(data, c)
			} else {
				parser.ungetChar(c)
				return Token{
					string(data),
					TYPE_NUMBER,
				}
			}
		}
	}

	if isAlpha(c) {
		data := []byte{c}
		for {
			c := parser.getChar()
			if isDigit(c) || isAlpha(c) {
				data = append(data, c)
			} else {
				parser.ungetChar(c)
				return Token{
					string(data),
					TYPE_SYMBOL,
				}
			}
		}
	}

	return Token{
		string([]byte{c}),
		TYPE_OTHER,
	}
}

func isDigit(c byte) bool {
	return (c>='0') && (c<='9')
}

func isAlpha(c byte) bool {
	return ((c>='a') && (c<='z')) || ((c>='A') && (c<='Z')) || (c=='_')
}

func (parser *Parser) forObjectStack(
	actor func(*Parser, *ParseObject, Token),
	sym Token) {

	parser.forObjectStackRec(actor, sym, parser.ObjectRoot, 0)


}

func (parser *Parser) forObjectStackRec(
	actor func(*Parser, *ParseObject, Token),
	sym Token,
	obj *ParseObject, depth int) {

	if depth == len(parser.ObjectStack) {
		actor(parser, obj, sym)
		return
	}

	array := obj.Definitions[parser.ObjectStack[depth]]
	parser.forObjectStackDim(actor, sym, array, depth+1, 0, 0)
}

func (parser *Parser) forObjectStackDim(
	actor func(*Parser, *ParseObject, Token),
	sym Token,
	array *ParseObjectArray,
	depth int, dim int, seq int) int {

	if dim == len(array.Template.IndexSizes) {

		parser.forObjectStackRec(actor, sym, &array.Children[seq], depth)
		return seq + 1
	}

	isNamed := array.Template.IndexNames[dim] != ""
	stackDepth := len(parser.IndexValueStack)

	if isNamed {
		parser.IndexValueStack = append(parser.IndexValueStack, 0)
	}

	for i:=0; i<array.Template.IndexSizes[dim]; i++ {
		seq = parser.forObjectStackDim(actor, sym, array, depth, dim+1, seq)
		if isNamed {
			parser.IndexValueStack[stackDepth]++
		}
	}

	if isNamed {
		parser.IndexValueStack = parser.IndexValueStack[0:stackDepth]
	}

	return seq
}


func (parser *Parser) pushBlockContext(name string, tp *ObjectArrayTemplate) {
	parser.ObjectStack = append(parser.ObjectStack, name)
	if tp.IndexNames[0] != "" {
		for _,txt := range tp.IndexNames {
			parser.IndexNameStack = append(parser.IndexNameStack, txt)
		}
	}
}

func (parser *Parser) popBlockContext(tp *ObjectArrayTemplate) {
	parser.ObjectStack = parser.ObjectStack[0:len(parser.ObjectStack)-1]
	if tp.IndexNames[0] != "" {
		n := len(parser.IndexNameStack) - len(tp.IndexNames)
		parser.IndexNameStack = parser.IndexNameStack[0:n]
	}
}


type NumberValue struct {
	x float32
}

func (v NumberValue) Type() int { return TYPE_NUMBER }
func (v NumberValue) Number() float32 { return v.x }
func (v NumberValue) Symbol() string { return "" }
func (v NumberValue) Object() *ParseObject { return nil }
func (v NumberValue) Array() *ParseObjectArray { return nil }

func (v NumberValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v NumberValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	fmt.Printf("Assign number %v to %v\n", v.x, sym.Text)
	object.Values[sym.Text] = v
}

type SymbolValue struct {
	x string
}

func (v SymbolValue) Type() int { return TYPE_SYMBOL }
func (v SymbolValue) Number() float32 { return 0 }
func (v SymbolValue) Symbol() string { return v.x }
func (v SymbolValue) Object() *ParseObject { return nil }
func (v SymbolValue) Array() *ParseObjectArray { return nil }

func (v SymbolValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v SymbolValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	fmt.Printf("Assign Symbol '%v' to %v\n", v.x, sym.Text)
	object.Values[sym.Text] = v
}

type ObjectValue struct {
	x *ParseObject
}

func (v ObjectValue) Type() int { return TYPE_OBJECT }
func (v ObjectValue) Number() float32 { return 0 }
func (v ObjectValue) Symbol() string { return "" }
func (v ObjectValue) Object() *ParseObject { return v.x }
func (v ObjectValue) Array() *ParseObjectArray { return nil }

func (v ObjectValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v ObjectValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	fmt.Printf("Assign Object %v to %v\n", v.x.ObjectType, sym.Text)
	object.Values[sym.Text] = v
}

type ArrayValue struct {
	x *ParseObjectArray
}

func (v ArrayValue) Type() int { return TYPE_ARRAY }
func (v ArrayValue) Number() float32 { return 0 }
func (v ArrayValue) Symbol() string { return "" }
func (v ArrayValue) Object() *ParseObject { return nil }
func (v ArrayValue) Array() *ParseObjectArray { return v.x }

func (v ArrayValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v ArrayValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	object.Values[sym.Text] = v // FIXME should setting variables to arrays be forbidden?
}

type MathOp struct {
	lhs Evaluator
	rhs Evaluator
	op string
}

func (v MathOp) Type() int { return TYPE_NUMBER }

func (v MathOp) Evaluate(parser *Parser) ParseValue {

	a := v.lhs.Evaluate(parser)
	b := v.rhs.Evaluate(parser)

	switch v.op {
	case "+": return NumberValue{a.Number() + b.Number()}
	case "-": return NumberValue{a.Number() - b.Number()}
	case "*": return NumberValue{a.Number() * b.Number()}
	case "/": return NumberValue{a.Number() / b.Number()}
	}

	panic("Invalid Operator")
}

func (v MathOp) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	fmt.Printf("MathOp Assign %v\n", v.Evaluate(parser))
	object.Values[sym.Text] = v.Evaluate(parser)
}

type IndexOp struct {
	i int
}

func (v IndexOp) Type() int { return TYPE_NUMBER }

func (v IndexOp) Evaluate(parser *Parser) ParseValue {

	if v.i < len(parser.IndexValueStack) {
		return NumberValue { float32(parser.IndexValueStack[v.i]) }
	} else {
		return NumberValue { 0 } // FIXME maybe create an error value?
	}
}

func (v IndexOp) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	fmt.Printf("IndexOp Assign %v\n", v.Evaluate(parser))
	object.Values[sym.Text] = v.Evaluate(parser)
}

type LookupOp struct {
	lhs Evaluator
	rhs string
	dtype int
}

func (v LookupOp) Type() int { return v.dtype }

func (v LookupOp) Evaluate(parser *Parser) ParseValue {

	var lhs *ParseObject
	if v.lhs == nil {
		lhs = parser.ObjectRoot
	} else {
		lhs = v.lhs.Evaluate(parser).Object()
	}
	if lhs == nil {
		return nil
	}
	return lhs.LookupValue(v.rhs)
}

func (v LookupOp) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	fmt.Printf("LookupOp Assign %v\n", v.Evaluate(parser))
	object.Values[sym.Text] = v.Evaluate(parser)
}

type LookupIndexOp struct {
	lhs Evaluator
	rhs Evaluator
}

func (v LookupIndexOp) Type() int { return TYPE_OBJECT }

func (v LookupIndexOp) Evaluate(parser *Parser) ParseValue {

	lhs := v.lhs.Evaluate(parser).Array()
	rhs := v.rhs.Evaluate(parser).Number()
	return ObjectValue{ &lhs.Children[int(rhs)] }
}

func (v LookupIndexOp) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	object.Values[sym.Text] = v.Evaluate(parser)
}


func (parser *Parser) LookupIndexName(name string) Evaluator {

	for i:=0; i<len(parser.IndexNameStack); i++ {
		if parser.IndexNameStack[i] == name {
			return IndexOp{ i }
		}
	}
	return nil
}


func (parser *Parser) ParseFile() error {

	parser.ObjectRoot = &ParseObject{}
	parser.ObjectRoot.Values = map[string]ParseValue{}
	parser.ObjectRoot.Definitions = map[string]*ParseObjectArray{}

	err := parser.ParseAssignmentList()
	if err != nil {
		return err
	}

	t := parser.GetToken()
	if t.Type != TYPE_EOF {
		return fmt.Errorf("unexpected token '%v'", t.Text)
	}

	return nil
}

func (parser *Parser) ParseAssignmentList() error {

	for {
		sym := parser.GetToken()
		if sym.Type == TYPE_EOF {
			return nil
		}
		// FIXME make a method to check for duplicate symbols

		if sym.Type != TYPE_SYMBOL {
			parser.UngetToken(sym)
			return nil
		}

		t2 := parser.GetToken()
		if t2.Type == TYPE_OTHER && t2.Text == ":" {
			exp, err := parser.ParseExpression()
			if err != nil {
				return err
			}

			parser.forObjectStack(exp.AssignTo, sym)

		} else if t2.Type == TYPE_SYMBOL {

			tp := ObjectArrayTemplate{}
			tp.ObjectType = t2.Text
			err := parser.ParseIndexDef(&tp)
			if err != nil {
				return err
			}

			if !parser.MatchToken("{") {
				return fmt.Errorf("Expected {")
			}

			parser.forObjectStack(tp.CreateObjects, t2)
			parser.pushBlockContext(t2.Text, &tp)

			err = parser.ParseAssignmentList()
			if err != nil {
				return err
			}

			if !parser.MatchToken("}") {
				return fmt.Errorf("Expected }")
			}

			err = parser.InsertDefaultValues(sym.Text)
			if err != nil {
				return err
			}

			parser.popBlockContext(&tp)

		} else {
			return fmt.Errorf("Unexpected token '%v'\n", sym.Text)
		}
	}
}

func (parser *Parser) ParseIndexDef(tp *ObjectArrayTemplate) error {

	if !parser.MatchToken("[") {
		tp.IndexSizes = []int{1}
		tp.IndexNames = []string{""}
		return nil
	}

	for {
		name := parser.GetToken()
		if name.Type != TYPE_SYMBOL {
			return fmt.Errorf("Expected a variable name for index")
		}

		if !parser.MatchToken("~") {
			return fmt.Errorf("Expected a ~")
		}

		sizeExp, err := parser.ParseExpression()
		if err != nil {
			return err
		}
		size := sizeExp.Evaluate(parser)
		if size.Type() != TYPE_NUMBER {
			return fmt.Errorf("Expected a constant number for index size")
		}

		tp.IndexSizes = append(tp.IndexSizes, int(size.Number()))
		tp.IndexNames = append(tp.IndexNames, name.Text)

		if parser.MatchToken("]") {
			return nil
		}

		if !parser.MatchToken(",") {
			return fmt.Errorf("expected , or ] in index list")
		}
	}
}

func (parser *Parser) ParseExpression() (Evaluator, error) {

	lhs, err := parser.ParseExpression1()
	if err != nil {
		return nil, err
	}

	return lhs, nil
}

func (parser *Parser) ParseExpression1() (Evaluator, error) {

	lhs, err := parser.ParseExpression2()
	if err != nil {
		return nil, err
	}

	for {
		if parser.MatchToken("*") {
			rhs, err := parser.ParseExpression2()
			if err != nil {
				return nil, err
			}
			if lhs.Type() != TYPE_NUMBER || rhs.Type() != TYPE_NUMBER {
				return nil, fmt.Errorf("Expected a number")
			}
			lhs = MathOp {
				lhs: lhs,
				rhs: rhs,
				op: "*",
			}

		} else if parser.MatchToken("/") {
			rhs, err := parser.ParseExpression2()
			if err != nil {
				return nil, err
			}
			if lhs.Type() != TYPE_NUMBER || rhs.Type() != TYPE_NUMBER {
				return nil, fmt.Errorf("Expected a number")
			}
			lhs = MathOp {
				lhs: lhs,
				rhs: rhs,
				op: "/",
			}
		} else {
			break
		}
	}

	return lhs, nil
}

func (parser *Parser) ParseExpression2() (Evaluator, error) {

	lhs, err := parser.ParseExpression3()
	if err != nil {
		return nil, err
	}

	return lhs, nil
}

func (parser *Parser) ParseExpression3() (Evaluator, error) {

	tok := parser.GetToken()
	if tok.Type == TYPE_NUMBER {
		x,_ := strconv.ParseFloat(tok.Text, 32)
		n := NumberValue{float32(x)}
		return n, nil
	}

	if tok.Text == "(" {
		ret, err := parser.ParseExpression()
		if err != nil {
			return nil, err
		}
		if !parser.MatchToken(")") {
			return nil, fmt.Errorf("Expected )")
		}
		return ret, nil
	}

	if tok.Type == TYPE_SYMBOL {
		ret := parser.LookupIndexName(tok.Text)
		if ret != nil {
			return ret, nil
		}

		ret = BuiltinValues[tok.Text]
		if ret != nil {
			return ret, nil
		}

		parser.UngetToken(tok)
		return parser.ParseDotList()
	}

	return nil, fmt.Errorf("Unexpected token '%v'\n", tok.Text)
}

func (parser *Parser) ParseDotList() (Evaluator, error) {

	var lhs Evaluator
	var err error

	for {
		lhs, err = parser.ParseName(lhs)
		if err != nil {
			return nil, err
		}

		if !parser.MatchToken(".") {
			return lhs, nil
		}
	}
}

func (parser *Parser) ParseName(parent Evaluator) (Evaluator, error) {

	tok := parser.GetToken()
	if tok.Type != TYPE_SYMBOL {
		return nil, fmt.Errorf("expected variable name")
	}

	// FIXME need to handle both the current object and the root object
	//       as base cases.

	// FIXME evaluating LookupOp to check validity and find type
	//       this is inefficient, but probably correct.
	op := LookupOp { parent, tok.Text, 0 }
	dval := op.Evaluate(parser)
	if dval == nil {
		return nil, fmt.Errorf("undefined variable %v", tok.Text)
	}
	op.dtype = dval.Type()

	if parser.MatchToken("[") {
		// FAKE should be an index list
		idx,err := parser.ParseExpression()
		if err != nil {
			return nil, err
		}
		if idx.Type() != TYPE_NUMBER {
			return nil, fmt.Errorf("expected a number as the index")
		}

		if !parser.MatchToken("]") {
			return nil, fmt.Errorf("expected a ]")
		}

		return LookupIndexOp { op, idx }, nil
	}

	return op, nil
}

func (parser *Parser) InsertDefaultValues(objType string) error {

	objList := DefaultObjects[objType]
	for _,obj := range objList {

		tok := Token{ obj[1], TYPE_SYMBOL }
		tp := &ObjectArrayTemplate {
			ObjectType:obj[0],
			IndexSizes: []int{1},
			IndexNames: []string{""},
		}
		parser.forObjectStack(tp.CreateObjects, tok)
	}
	return nil
}

var DefaultObjects = map[string] [][2]string {
	"Sheet": {
		{ "DefaultMount", "Tip" },
		{ "DefaultMount", "Root" },
		{ "DefaultMount", "LeadingEdge" },
		{ "DefaultMount", "TrailingEdge" },
	},
	"Tube": {
		{ "DefaultMount", "Top" },
		{ "DefaultMount", "Bottom" },
	},
	"Transition": {
		{ "DefaultMount", "Top" },
		{ "DefaultMount", "Bottom" },
	},
}


var BuiltinValues = map[string]Evaluator {
	"mm": SymbolValue { "mm" },
	"Mount": SymbolValue { "Mount" },
	"InMount": SymbolValue { "InMount" },
	"Ragged": SymbolValue { "Ragged" },
	"Sharp": SymbolValue { "Sharp" },
	"Flat": SymbolValue { "Flat" },
	"Round": SymbolValue { "Round" },
	"Ogive": SymbolValue { "Ogive" },
	"EngineFlat": SymbolValue { "EngineFlat" },
}

func ParseTest() {
	parser,_ := NewParser("aerodynamics/testmodel.txt")

	fmt.Println(parser.ParseFile())
	fmt.Printf("line count = %v\n", parser.lineCount)

	/*
	for {
		t := parser.GetToken()
		if t.Type == TYPE_EOF {
			break
		}
		fmt.Printf("%d: %s\n", t.Type, t.Text)
	}
	*/
}

