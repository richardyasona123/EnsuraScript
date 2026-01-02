// Package parser implements a recursive descent parser for EnsuraScript.
package parser

import (
	"fmt"
	"strconv"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/lexer"
)

// Parser parses EnsuraScript source code into an AST.
type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
}

// New creates a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Read two tokens so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

// ParseString parses source code from a string.
func ParseString(input string) (*ast.Program, []string) {
	l := lexer.New(input)
	p := New(l)
	return p.Parse(), p.Errors()
}

// ParseFile parses source code from a file.
func ParseFile(input, filename string) (*ast.Program, []string) {
	l := lexer.NewWithFilename(input, filename)
	p := New(l)
	return p.Parse(), p.Errors()
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	// Skip comments
	for p.peekToken.Type == lexer.COMMENT {
		p.peekToken = p.l.NextToken()
	}
}

// Errors returns all parser errors.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("%s: %s", p.curToken.Pos, msg))
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected %s, got %s", t, p.peekToken.Type))
	return false
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// Parse parses the input and returns the AST.
func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}
	if p.curToken.Type != lexer.EOF {
		program.Position = p.curToken.Pos
	}

	for p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.RESOURCE:
		return p.parseResourceDecl()
	case lexer.ENSURE:
		return p.parseEnsureStmt()
	case lexer.ON:
		return p.parseOnBlock()
	case lexer.POLICY:
		return p.parsePolicyDecl()
	case lexer.APPLY:
		return p.parseApplyStmt()
	case lexer.FOR:
		return p.parseForEachStmt()
	case lexer.INVARIANT:
		return p.parseInvariantBlock()
	case lexer.ASSUME:
		return p.parseAssumeStmt()
	case lexer.PARALLEL:
		return p.parseParallelBlock()
	case lexer.COMMENT:
		return nil
	default:
		p.addError(fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseResourceDecl() *ast.ResourceDecl {
	decl := &ast.ResourceDecl{Position: p.curToken.Pos}

	// resource <type> "<path>" [as <alias>]
	// type can be a keyword (file, http, etc.) or identifier
	if !p.expectResourceTypeOrIdent() {
		return nil
	}
	decl.ResourceType = p.curToken.Literal

	if !p.expectPeek(lexer.STRING) {
		return nil
	}
	decl.Path = p.curToken.Literal

	// Optional: as <alias>
	if p.peekTokenIs(lexer.AS) {
		p.nextToken()
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Alias = p.curToken.Literal
	}

	return decl
}

func (p *Parser) expectResourceType() bool {
	switch p.peekToken.Type {
	case lexer.FILE, lexer.DIRECTORY, lexer.HTTP, lexer.DATABASE, lexer.SERVICE, lexer.PROCESS, lexer.CRON:
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) expectResourceTypeOrIdent() bool {
	switch p.peekToken.Type {
	case lexer.FILE, lexer.DIRECTORY, lexer.HTTP, lexer.DATABASE, lexer.SERVICE, lexer.PROCESS, lexer.CRON, lexer.IDENT:
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected resource type or identifier, got %s", p.peekToken.Type))
	return false
}

func (p *Parser) isResourceType(t lexer.TokenType) bool {
	switch t {
	case lexer.FILE, lexer.DIRECTORY, lexer.HTTP, lexer.DATABASE, lexer.SERVICE, lexer.PROCESS, lexer.CRON:
		return true
	}
	return false
}

func (p *Parser) parseResourceRef() *ast.ResourceRef {
	ref := &ast.ResourceRef{Position: p.curToken.Pos}

	switch p.curToken.Type {
	case lexer.FILE, lexer.DIRECTORY, lexer.HTTP, lexer.DATABASE, lexer.SERVICE, lexer.PROCESS, lexer.CRON:
		ref.ResourceType = p.curToken.Literal
		if !p.expectPeek(lexer.STRING) {
			return nil
		}
		ref.Path = p.curToken.Literal
	case lexer.IDENT:
		// Could be an alias or a resource type
		if p.peekTokenIs(lexer.STRING) {
			ref.ResourceType = p.curToken.Literal
			p.nextToken()
			ref.Path = p.curToken.Literal
		} else {
			ref.Alias = p.curToken.Literal
		}
	default:
		p.addError(fmt.Sprintf("expected resource reference, got %s", p.curToken.Type))
		return nil
	}

	return ref
}

func (p *Parser) parseEnsureStmt() *ast.EnsureStmt {
	stmt := &ast.EnsureStmt{Position: p.curToken.Pos}

	// ensure <condition>
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Condition = p.curToken.Literal

	// Parse optional clauses
	for {
		switch {
		case p.peekTokenIs(lexer.ON):
			// Check if this is "on violation" or "on <resource>"
			if stmt.Subject != nil {
				// Already have a subject, check if this is "on violation"
				p.nextToken() // consume 'on'
				if p.peekTokenIs(lexer.ON_VIOLATION) || (p.peekTokenIs(lexer.IDENT) && p.peekToken.Literal == "violation") {
					p.nextToken() // consume 'violation'
					stmt.ViolationHandler = p.parseViolationHandlerBlock()
					return stmt
				}
				// Not a violation handler, backtrack by returning
				return stmt
			}
			p.nextToken()
			p.nextToken()
			stmt.Subject = p.parseResourceRef()
		case p.peekTokenIs(lexer.WITH):
			p.nextToken()
			stmt.Handler = p.parseHandlerSpec()
		case p.peekTokenIs(lexer.WHEN):
			p.nextToken()
			stmt.Guard = p.parseGuardExpr()
		case p.peekTokenIs(lexer.REQUIRES):
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(lexer.IDENT) {
				// Could be a condition or a resource reference
				if p.peekTokenIs(lexer.STRING) {
					// It's a resource reference with condition
					ref := p.parseResourceRef()
					stmt.RequiresResource = append(stmt.RequiresResource, ref)
				} else {
					stmt.Requires = append(stmt.Requires, p.curToken.Literal)
				}
			}
		case p.peekTokenIs(lexer.AFTER):
			p.nextToken()
			p.nextToken()
			ref := p.parseResourceRef()
			if ref != nil {
				stmt.After = append(stmt.After, ref)
			}
		case p.peekTokenIs(lexer.BEFORE):
			p.nextToken()
			p.nextToken()
			ref := p.parseResourceRef()
			if ref != nil {
				stmt.Before = append(stmt.Before, ref)
			}
		default:
			return stmt
		}
	}
}

func (p *Parser) parseHandlerSpec() *ast.HandlerSpec {
	spec := &ast.HandlerSpec{
		Position: p.curToken.Pos,
		Args:     make(map[string]string),
	}

	// with <handler_name> [key value ...]
	// Handler names can be identifiers OR resource type keywords (http, cron, etc.)
	p.nextToken()
	if p.curTokenIs(lexer.IDENT) || p.curTokenIs(lexer.HTTP) || p.curTokenIs(lexer.CRON) {
		spec.Name = p.curToken.Literal
	} else {
		p.addError(fmt.Sprintf("expected handler name, got %s", p.curToken.Type))
		return nil
	}

	// Check for colon-separated handler name (e.g., AES:256)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		if p.expectPeek(lexer.NUMBER) || p.expectPeek(lexer.IDENT) {
			spec.Name = spec.Name + ":" + p.curToken.Literal
		}
	}

	// Parse key-value arguments
	for p.peekTokenIs(lexer.IDENT) || p.peekTokenIs(lexer.KEY) || p.peekTokenIs(lexer.MODE) {
		p.nextToken()
		key := p.curToken.Literal
		// Value can be a string or an identifier (for policy parameter substitution)
		if p.peekTokenIs(lexer.STRING) || p.peekTokenIs(lexer.IDENT) {
			p.nextToken()
			spec.Args[key] = p.curToken.Literal
		}
	}

	return spec
}

func (p *Parser) parseGuardExpr() *ast.GuardExpr {
	guard := &ast.GuardExpr{Position: p.curToken.Pos}

	// when <left> <op> <right>
	if p.peekTokenIs(lexer.IDENT) || p.peekTokenIs(lexer.ENVIRONMENT) {
		p.nextToken()
	} else {
		p.addError(fmt.Sprintf("expected identifier, got %s", p.peekToken.Type))
		return nil
	}
	guard.Left = p.curToken.Literal

	if p.peekTokenIs(lexer.EQUALS) {
		p.nextToken()
		guard.Operator = "=="
	} else if p.peekTokenIs(lexer.NOTEQUALS) {
		p.nextToken()
		guard.Operator = "!="
	} else {
		p.addError("expected == or !=")
		return nil
	}

	if !p.expectPeek(lexer.STRING) {
		return nil
	}
	guard.Right = p.curToken.Literal

	return guard
}

func (p *Parser) parseOnBlock() ast.Statement {
	pos := p.curToken.Pos

	p.nextToken()

	// Check if it's "on violation"
	if p.curTokenIs(lexer.ON_VIOLATION) || (p.curTokenIs(lexer.IDENT) && p.curToken.Literal == "violation") {
		return p.parseOnViolationBlock(pos)
	}

	block := &ast.OnBlock{Position: pos}
	block.Subject = p.parseResourceRef()

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	block.Statements = p.parseBlockStatements()

	return block
}

func (p *Parser) parseOnViolationBlock(pos lexer.Position) *ast.OnViolationBlock {
	block := &ast.OnViolationBlock{
		Position: pos,
		Handler:  &ast.ViolationHandler{Position: pos},
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		switch p.curToken.Type {
		case lexer.RETRY:
			if p.expectPeek(lexer.NUMBER) {
				n, _ := strconv.Atoi(p.curToken.Literal)
				block.Handler.Retry = n
			}
		case lexer.NOTIFY:
			if p.expectPeek(lexer.STRING) {
				block.Handler.Notify = append(block.Handler.Notify, p.curToken.Literal)
			}
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseViolationHandlerBlock() *ast.ViolationHandler {
	handler := &ast.ViolationHandler{Position: p.curToken.Pos}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		switch p.curToken.Type {
		case lexer.RETRY:
			if p.expectPeek(lexer.NUMBER) {
				n, _ := strconv.Atoi(p.curToken.Literal)
				handler.Retry = n
			}
		case lexer.NOTIFY:
			if p.expectPeek(lexer.STRING) {
				handler.Notify = append(handler.Notify, p.curToken.Literal)
			}
		}
		p.nextToken()
	}

	return handler
}

func (p *Parser) parseBlockStatements() []ast.Statement {
	var statements []ast.Statement

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}

	return statements
}

func (p *Parser) parsePolicyDecl() *ast.PolicyDecl {
	decl := &ast.PolicyDecl{Position: p.curToken.Pos}

	// policy <name> [(<params>)] { ... }
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal

	// Optional parameters
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken()
		decl.Params = p.parsePolicyParams()
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	decl.Statements = p.parseBlockStatements()

	return decl
}

func (p *Parser) parsePolicyParams() []ast.PolicyParam {
	var params []ast.PolicyParam

	p.nextToken()

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.IDENT) {
			params = append(params, ast.PolicyParam{Name: p.curToken.Literal})
		}
		p.nextToken()
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	return params
}

func (p *Parser) parseApplyStmt() *ast.ApplyStmt {
	stmt := &ast.ApplyStmt{Position: p.curToken.Pos}

	// apply <policy_name> [(<args>)]
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.PolicyName = p.curToken.Literal

	// Optional arguments
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken()
		stmt.Args = p.parseApplyArgs()
	}

	return stmt
}

func (p *Parser) parseApplyArgs() []string {
	var args []string

	p.nextToken()

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.STRING) {
			args = append(args, p.curToken.Literal)
		}
		p.nextToken()
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	return args
}

func (p *Parser) parseForEachStmt() *ast.ForEachStmt {
	stmt := &ast.ForEachStmt{Position: p.curToken.Pos}

	// for each <type> in <container> { ... }
	if !p.expectPeek(lexer.EACH) {
		return nil
	}

	if !p.expectResourceTypeOrIdent() {
		return nil
	}
	stmt.ItemType = p.curToken.Literal

	if !p.expectPeek(lexer.IN) {
		return nil
	}

	p.nextToken()
	stmt.Container = p.parseResourceRef()

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Statements = p.parseBlockStatements()

	return stmt
}

func (p *Parser) parseInvariantBlock() *ast.InvariantBlock {
	block := &ast.InvariantBlock{Position: p.curToken.Pos}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	block.Statements = p.parseBlockStatements()

	return block
}

func (p *Parser) parseAssumeStmt() *ast.AssumeStmt {
	stmt := &ast.AssumeStmt{Position: p.curToken.Pos}

	p.nextToken()

	// Check if it's a guard expression or simple assumption
	if p.curTokenIs(lexer.IDENT) || p.curTokenIs(lexer.ENVIRONMENT) {
		if p.peekTokenIs(lexer.EQUALS) || p.peekTokenIs(lexer.NOTEQUALS) {
			// It's a guard expression
			stmt.Guard = &ast.GuardExpr{Position: p.curToken.Pos}
			stmt.Guard.Left = p.curToken.Literal
			p.nextToken()
			if p.curTokenIs(lexer.EQUALS) {
				stmt.Guard.Operator = "=="
			} else {
				stmt.Guard.Operator = "!="
			}
			if p.expectPeek(lexer.STRING) {
				stmt.Guard.Right = p.curToken.Literal
			}
		} else {
			// Simple assumption - read until end of statement
			stmt.Simple = p.curToken.Literal
			for p.peekTokenIs(lexer.IDENT) {
				p.nextToken()
				stmt.Simple += " " + p.curToken.Literal
			}
		}
	}

	return stmt
}

func (p *Parser) parseParallelBlock() *ast.ParallelBlock {
	block := &ast.ParallelBlock{Position: p.curToken.Pos}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	block.Statements = p.parseBlockStatements()

	return block
}
