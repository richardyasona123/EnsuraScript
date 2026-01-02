// Package lexer provides tokenization for EnsuraScript source files.
package lexer

import "fmt"

// TokenType represents the type of a token.
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT  // identifiers like file, exists, etc.
	STRING // "string literal"
	NUMBER // 123, 0600

	// Delimiters
	LBRACE    // {
	RBRACE    // }
	LPAREN    // (
	RPAREN    // )
	COMMA     // ,
	COLON     // :
	NEWLINE   // \n (significant in some contexts)
	EQUALS    // ==
	NOTEQUALS // !=

	// Keywords
	RESOURCE
	ENSURE
	ON
	WITH
	REQUIRES
	AFTER
	BEFORE
	POLICY
	APPLY
	ON_VIOLATION
	RETRY
	NOTIFY
	ASSUME
	WHEN
	FOR
	EACH
	IN
	INVARIANT
	AS
	KEY
	MODE
	DIRECTORY
	FILE
	HTTP
	DATABASE
	SERVICE
	PROCESS
	CRON
	ENVIRONMENT
	PARALLEL
)

var tokenNames = map[TokenType]string{
	ILLEGAL:      "ILLEGAL",
	EOF:          "EOF",
	COMMENT:      "COMMENT",
	IDENT:        "IDENT",
	STRING:       "STRING",
	NUMBER:       "NUMBER",
	LBRACE:       "LBRACE",
	RBRACE:       "RBRACE",
	LPAREN:       "LPAREN",
	RPAREN:       "RPAREN",
	COMMA:        "COMMA",
	COLON:        "COLON",
	NEWLINE:      "NEWLINE",
	EQUALS:       "EQUALS",
	NOTEQUALS:    "NOTEQUALS",
	RESOURCE:     "RESOURCE",
	ENSURE:       "ENSURE",
	ON:           "ON",
	WITH:         "WITH",
	REQUIRES:     "REQUIRES",
	AFTER:        "AFTER",
	BEFORE:       "BEFORE",
	POLICY:       "POLICY",
	APPLY:        "APPLY",
	ON_VIOLATION: "ON_VIOLATION",
	RETRY:        "RETRY",
	NOTIFY:       "NOTIFY",
	ASSUME:       "ASSUME",
	WHEN:         "WHEN",
	FOR:          "FOR",
	EACH:         "EACH",
	IN:           "IN",
	INVARIANT:    "INVARIANT",
	AS:           "AS",
	KEY:          "KEY",
	MODE:         "MODE",
	DIRECTORY:    "DIRECTORY",
	FILE:         "FILE",
	HTTP:         "HTTP",
	DATABASE:     "DATABASE",
	SERVICE:      "SERVICE",
	PROCESS:      "PROCESS",
	CRON:         "CRON",
	ENVIRONMENT:  "ENVIRONMENT",
	PARALLEL:     "PARALLEL",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

var keywords = map[string]TokenType{
	"resource":    RESOURCE,
	"ensure":      ENSURE,
	"on":          ON,
	"with":        WITH,
	"requires":    REQUIRES,
	"after":       AFTER,
	"before":      BEFORE,
	"policy":      POLICY,
	"apply":       APPLY,
	"violation":   ON_VIOLATION, // used after "on"
	"retry":       RETRY,
	"notify":      NOTIFY,
	"assume":      ASSUME,
	"when":        WHEN,
	"for":         FOR,
	"each":        EACH,
	"in":          IN,
	"invariant":   INVARIANT,
	"as":          AS,
	"key":         KEY,
	"mode":        MODE,
	"directory":   DIRECTORY,
	"file":        FILE,
	"http":        HTTP,
	"database":    DATABASE,
	"service":     SERVICE,
	"process":     PROCESS,
	"cron":        CRON,
	"environment": ENVIRONMENT,
	"parallel":    PARALLEL,
}

// LookupIdent returns the token type for an identifier.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// Position represents a source position.
type Position struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

func (p Position) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %s", t.Type, t.Literal, t.Pos)
}
