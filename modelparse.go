
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

	fmt.Printf("FIXME create %v objects %v in context %v\n", n, sym.Text, parser.ObjectStack)
	fmt.Println(tp)
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

		fmt.Printf("FIXME running forObjectStackRec on %v in context %v %d\n", sym.Text, parser.ObjectStack, depth)

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

	for i:=0; i<array.Template.IndexSizes[dim]; i++ {
		// FIXME add index to symbol table
		seq = parser.forObjectStackDim(actor, sym, array, depth, dim+1, seq)
		// FIXME remove index from symbol table
	}
	return seq
}


func (parser *Parser) pushBlockContext(name string) {
	parser.ObjectStack = append(parser.ObjectStack, name)
	fmt.Println(parser.ObjectStack)
}

func (parser *Parser) popBlockContext() {
	parser.ObjectStack = parser.ObjectStack[0:len(parser.ObjectStack)-1]
	fmt.Println(parser.ObjectStack)
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

			parser.forObjectStack(tp.CreateObjects, sym)
			parser.pushBlockContext(sym.Text)

			err = parser.ParseAssignmentList()
			if err != nil {
				return err
			}

			if !parser.MatchToken("}") {
				return fmt.Errorf("Expected }")
			}
			parser.popBlockContext()

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

	if tok.Type == TYPE_SYMBOL {
		n := SymbolValue{tok.Text} // FIXME fake
		return n, nil
	}

	return nil, fmt.Errorf("Unexpected token '%v'\n", tok.Text)
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

