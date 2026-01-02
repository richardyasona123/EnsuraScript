package lexer

import (
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes EnsuraScript source code.
type Lexer struct {
	input    string
	filename string
	pos      int  // current position in input (points to current char)
	readPos  int  // current reading position in input (after current char)
	ch       rune // current character under examination
	line     int
	column   int
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return NewWithFilename(input, "")
}

// NewWithFilename creates a new Lexer with a filename for error messages.
func NewWithFilename(input, filename string) *Lexer {
	l := &Lexer{
		input:    input,
		filename: filename,
		line:     1,
		column:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	l.pos = l.readPos
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPos:])
		l.ch = r
		l.readPos += size
	}
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *Lexer) currentPos() Position {
	return Position{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.column,
		Offset:   l.pos,
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	pos := l.currentPos()

	var tok Token
	tok.Pos = pos

	switch l.ch {
	case '{':
		tok = l.newToken(LBRACE, string(l.ch))
	case '}':
		tok = l.newToken(RBRACE, string(l.ch))
	case '(':
		tok = l.newToken(LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(RPAREN, string(l.ch))
	case ',':
		tok = l.newToken(COMMA, string(l.ch))
	case ':':
		tok = l.newToken(COLON, string(l.ch))
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(EQUALS, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(NOTEQUALS, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	case '#':
		tok.Type = COMMENT
		tok.Literal = l.readComment()
		tok.Pos = pos
		return tok
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		tok.Pos = pos
		return tok
	case 0:
		tok.Type = EOF
		tok.Literal = ""
		tok.Pos = pos
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			tok.Pos = pos
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = NUMBER
			tok.Pos = pos
			return tok
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Pos:     l.currentPos(),
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readNumber() string {
	start := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening quote
	start := l.pos
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip escape char
		}
		l.readChar()
	}
	str := l.input[start:l.pos]
	if l.ch == '"' {
		l.readChar() // skip closing quote
	}
	return str
}

func (l *Lexer) readComment() string {
	l.readChar() // skip #
	start := l.pos
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// Tokenize returns all tokens from the input.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		if tok.Type == COMMENT {
			continue // skip comments
		}
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}

// TokenizeAll returns all tokens including comments.
func (l *Lexer) TokenizeAll() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}
