
package parser

const (
	TYPE_EOF = 0
	TYPE_SYMBOL = 1
	TYPE_NUMBER = 2
	TYPE_OTHER = 3
	TYPE_OBJECT = 4
	TYPE_ARRAY = 5
)

type Token struct {
	Text string
	Type int
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

