
package parser

import (
	"io"
	"os"
	"fmt"
	"strconv"
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
	LocalObject *ParseObject
	ObjectStack []string
	IndexNameStack []string
	IndexValueStack []int
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
	parser.LocalObject = parser.ObjectRoot
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
		// FIXME clarify using MatchToken?
		if t2.Type == TYPE_OTHER && t2.Text == ":" {

			// FIXME clarify with an IsDefined method?
			if parser.LocalObject.Values[sym.Text] != nil {
				return fmt.Errorf("value %v already defined", sym.Text)
			}
			if parser.LocalObject.Definitions[t2.Text] != nil {
				return fmt.Errorf("value %v already defined", sym.Text)
			}

			exp, err := parser.ParseExpression()
			if err != nil {
				return err
			}

			parser.forObjectStack(exp.AssignTo, sym)

		} else if t2.Type == TYPE_SYMBOL {

			// FIXME clarify with an IsDefined method?
			if parser.LocalObject.Values[t2.Text] != nil {
				return fmt.Errorf("value %v already defined", t2.Text)
			}
			if parser.LocalObject.Definitions[t2.Text] != nil {
				return fmt.Errorf("value %v already defined", t2.Text)
			}

			tp := ObjectArrayTemplate{}
			tp.ObjectType = sym.Text
			err := parser.ParseIndexDef(&tp)
			if err != nil {
				return err
			}

			if !parser.MatchToken("{") {
				return fmt.Errorf("Expected {")
			}

			// FIXME check for duplicates
			parser.forObjectStack(tp.CreateObjects, t2)
			parser.pushBlockContext(t2.Text, &tp)
			savedLocal := parser.LocalObject // FIXME make this cleaner?
			parser.LocalObject = &savedLocal.Definitions[t2.Text].Children[0]

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
			parser.LocalObject = savedLocal

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
		n := size.Number()
		if n != float32(int(n)) {
			return fmt.Errorf("Index size should be an integer")
		}
		if n < 1 {
			return fmt.Errorf("Index size should be positive")
		}

		tp.IndexSizes = append(tp.IndexSizes, int(n))
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

	// FIXME more math operators
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

type LocalObject struct {
	Op Evaluator
	Example ParseValue
}

func (parser *Parser) ParseDotList() (Evaluator, error) {

	lhs := &LocalObject{ nil, ObjectValue{ parser.ObjectRoot } }
	var err error

	for {
		lhs, err = parser.ParseName(lhs)
		if err != nil {
			return nil, err
		}

		if !parser.MatchToken(".") {
			return lhs.Op, nil
		}
	}
}

func (parser *Parser) ParseName(parent *LocalObject) (*LocalObject, error) {

	// FIXME handle builtin function calls like sqrt(x)

	tok := parser.GetToken()
	if tok.Type != TYPE_SYMBOL {
		return nil, fmt.Errorf("expected variable name")
	}

	dval := parent.Example.Object().LookupValue(tok.Text)
	if dval == nil {
		return nil, fmt.Errorf("undefined variable %v", tok.Text)
	}

	op := LookupOp { parent.Op, tok.Text, 0 }
	op.dtype = dval.Type()

	ret := LocalObject { op, dval }

	if parser.MatchToken("[") {

		if dval.Type() != TYPE_ARRAY {
			return nil, fmt.Errorf("indexing something that isn't an array")
		}
		// FIXME check number of array dimensions

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

		ret.Example = ObjectValue { &dval.Array().Children[0] }
		ret.Op = LookupIndexOp { op, idx }
	}

	return &ret, nil
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

	parser.ObjectRoot.Print(0)

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

