
package parser

import (
	"fmt"
)

type ParseValue interface {
	Type() int
	Number() float32
	Symbol() string
	Object() *ParseObject
	Array() *ParseObjectArray
	ToString() string
}

type Evaluator interface {
	Type() int
	Evaluate(parser *Parser) ParseValue
	AssignTo(parser *Parser, object *ParseObject, sym Token)
}

type NumberValue struct {
	x float32
}

func (v NumberValue) Type() int { return TYPE_NUMBER }
func (v NumberValue) Number() float32 { return v.x }
func (v NumberValue) Symbol() string { return "" }
func (v NumberValue) Object() *ParseObject { return nil }
func (v NumberValue) Array() *ParseObjectArray { return nil }
func (v NumberValue) ToString() string { return fmt.Sprintf("%v", v.x) }

func (v NumberValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v NumberValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
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
func (v SymbolValue) ToString() string { return v.x }

func (v SymbolValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v SymbolValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
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

func (v ObjectValue) ToString() string {
	ret := ""
	if v.x.Parent != nil {
		ret = fmt.Sprintf("%v#%v->", v.x.Parent.Name, v.x.Parent.Index)
	}
	ret +=fmt.Sprintf("%v#%v", v.x.Name, v.x.Index)
	return ret
}

func (v ObjectValue) Evaluate(parser *Parser) ParseValue {
	return v
}
func (v ObjectValue) AssignTo(parser *Parser, object *ParseObject, sym Token) {
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
func (v ArrayValue) ToString() string { return fmt.Sprintf("Array%v", v.x) }

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

	n := int(rhs) % len(lhs.Children)
	return ObjectValue{ &lhs.Children[n] }
}

func (v LookupIndexOp) AssignTo(parser *Parser, object *ParseObject, sym Token) {
	object.Values[sym.Text] = v.Evaluate(parser)
}

