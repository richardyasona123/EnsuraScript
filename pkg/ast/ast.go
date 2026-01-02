// Package ast defines the Abstract Syntax Tree for EnsuraScript.
package ast

import (
	"fmt"
	"strings"

	"github.com/ensurascript/ensura/pkg/lexer"
)

// Node represents any node in the AST.
type Node interface {
	Pos() lexer.Position
	String() string
}

// Statement represents a statement node.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST.
type Program struct {
	Statements []Statement
	Position   lexer.Position
}

func (p *Program) Pos() lexer.Position { return p.Position }
func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// ResourceDecl represents a resource declaration.
// Example: resource file "secrets.db" as secrets_db
type ResourceDecl struct {
	Position     lexer.Position
	ResourceType string // file, directory, http, database, etc.
	Path         string // the resource path/identifier
	Alias        string // optional alias (from "as")
}

func (r *ResourceDecl) statementNode()        {}
func (r *ResourceDecl) Pos() lexer.Position   { return r.Position }
func (r *ResourceDecl) String() string {
	if r.Alias != "" {
		return fmt.Sprintf("resource %s %q as %s", r.ResourceType, r.Path, r.Alias)
	}
	return fmt.Sprintf("resource %s %q", r.ResourceType, r.Path)
}

// ResourceRef references a resource (inline or by alias).
type ResourceRef struct {
	Position     lexer.Position
	ResourceType string // file, directory, http, etc.
	Path         string // the resource path (if inline)
	Alias        string // the alias (if referencing by alias)
}

func (r *ResourceRef) expressionNode()       {}
func (r *ResourceRef) Pos() lexer.Position   { return r.Position }
func (r *ResourceRef) String() string {
	if r.Alias != "" {
		return r.Alias
	}
	return fmt.Sprintf("%s %q", r.ResourceType, r.Path)
}

// HandlerSpec represents a handler specification with its arguments.
// Example: AES:256 key "env:SECRET_KEY"
type HandlerSpec struct {
	Position lexer.Position
	Name     string            // e.g., "AES:256", "posix", "http.get"
	Args     map[string]string // key-value arguments
}

func (h *HandlerSpec) expressionNode()       {}
func (h *HandlerSpec) Pos() lexer.Position   { return h.Position }
func (h *HandlerSpec) String() string {
	var args []string
	for k, v := range h.Args {
		args = append(args, fmt.Sprintf("%s %q", k, v))
	}
	if len(args) > 0 {
		return fmt.Sprintf("%s %s", h.Name, strings.Join(args, " "))
	}
	return h.Name
}

// ViolationHandler represents violation handling configuration.
type ViolationHandler struct {
	Position lexer.Position
	Retry    int      // number of retries
	Notify   []string // notification targets
}

func (v *ViolationHandler) expressionNode()       {}
func (v *ViolationHandler) Pos() lexer.Position   { return v.Position }
func (v *ViolationHandler) String() string {
	var parts []string
	if v.Retry > 0 {
		parts = append(parts, fmt.Sprintf("retry %d", v.Retry))
	}
	for _, n := range v.Notify {
		parts = append(parts, fmt.Sprintf("notify %q", n))
	}
	return strings.Join(parts, "\n  ")
}

// GuardExpr represents a conditional guard.
// Example: environment == "prod"
type GuardExpr struct {
	Position lexer.Position
	Left     string // e.g., "environment"
	Operator string // "==" or "!="
	Right    string // e.g., "prod"
}

func (g *GuardExpr) expressionNode()       {}
func (g *GuardExpr) Pos() lexer.Position   { return g.Position }
func (g *GuardExpr) String() string {
	return fmt.Sprintf("%s %s %q", g.Left, g.Operator, g.Right)
}

// EnsureStmt represents an ensure statement.
// Example: ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
type EnsureStmt struct {
	Position         lexer.Position
	Condition        string            // exists, encrypted, permissions, etc.
	Subject          *ResourceRef      // the resource (may be nil if inherited)
	Handler          *HandlerSpec      // optional handler specification
	Guard            *GuardExpr        // optional when clause
	Requires         []string          // required conditions
	RequiresResource []*ResourceRef    // required resources with conditions
	After            []*ResourceRef    // ordering: after these
	Before           []*ResourceRef    // ordering: before these
	ViolationHandler *ViolationHandler // per-ensure violation handling
}

func (e *EnsureStmt) statementNode()        {}
func (e *EnsureStmt) Pos() lexer.Position   { return e.Position }
func (e *EnsureStmt) String() string {
	var out strings.Builder
	out.WriteString("ensure ")
	out.WriteString(e.Condition)
	if e.Subject != nil {
		out.WriteString(" on ")
		out.WriteString(e.Subject.String())
	}
	if e.Handler != nil {
		out.WriteString(" with ")
		out.WriteString(e.Handler.String())
	}
	if e.Guard != nil {
		out.WriteString(" when ")
		out.WriteString(e.Guard.String())
	}
	for _, r := range e.Requires {
		out.WriteString(" requires ")
		out.WriteString(r)
	}
	return out.String()
}

// OnBlock represents an "on resource { ... }" block.
type OnBlock struct {
	Position   lexer.Position
	Subject    *ResourceRef
	Statements []Statement
}

func (o *OnBlock) statementNode()        {}
func (o *OnBlock) Pos() lexer.Position   { return o.Position }
func (o *OnBlock) String() string {
	var out strings.Builder
	out.WriteString("on ")
	out.WriteString(o.Subject.String())
	out.WriteString(" {\n")
	for _, s := range o.Statements {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// PolicyParam represents a policy parameter.
type PolicyParam struct {
	Name string
}

// PolicyDecl represents a policy declaration.
type PolicyDecl struct {
	Position   lexer.Position
	Name       string
	Params     []PolicyParam
	Statements []Statement
}

func (p *PolicyDecl) statementNode()        {}
func (p *PolicyDecl) Pos() lexer.Position   { return p.Position }
func (p *PolicyDecl) String() string {
	var out strings.Builder
	out.WriteString("policy ")
	out.WriteString(p.Name)
	if len(p.Params) > 0 {
		out.WriteString("(")
		for i, param := range p.Params {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(param.Name)
		}
		out.WriteString(")")
	}
	out.WriteString(" {\n")
	for _, s := range p.Statements {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// ApplyStmt represents a policy application.
// Example: apply secure_file("env:SECRET_KEY")
type ApplyStmt struct {
	Position   lexer.Position
	PolicyName string
	Args       []string
}

func (a *ApplyStmt) statementNode()        {}
func (a *ApplyStmt) Pos() lexer.Position   { return a.Position }
func (a *ApplyStmt) String() string {
	if len(a.Args) > 0 {
		return fmt.Sprintf("apply %s(%s)", a.PolicyName, strings.Join(a.Args, ", "))
	}
	return fmt.Sprintf("apply %s", a.PolicyName)
}

// ForEachStmt represents a for each loop.
// Example: for each file in directory "/secrets" { ... }
type ForEachStmt struct {
	Position   lexer.Position
	ItemType   string      // file, etc.
	ItemVar    string      // implicit variable name
	Container  *ResourceRef // directory, etc.
	Statements []Statement
}

func (f *ForEachStmt) statementNode()        {}
func (f *ForEachStmt) Pos() lexer.Position   { return f.Position }
func (f *ForEachStmt) String() string {
	var out strings.Builder
	out.WriteString("for each ")
	out.WriteString(f.ItemType)
	out.WriteString(" in ")
	out.WriteString(f.Container.String())
	out.WriteString(" {\n")
	for _, s := range f.Statements {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// InvariantBlock represents an invariant block.
type InvariantBlock struct {
	Position   lexer.Position
	Statements []Statement
}

func (i *InvariantBlock) statementNode()        {}
func (i *InvariantBlock) Pos() lexer.Position   { return i.Position }
func (i *InvariantBlock) String() string {
	var out strings.Builder
	out.WriteString("invariant {\n")
	for _, s := range i.Statements {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// OnViolationBlock represents a global violation handler.
type OnViolationBlock struct {
	Position lexer.Position
	Handler  *ViolationHandler
}

func (o *OnViolationBlock) statementNode()        {}
func (o *OnViolationBlock) Pos() lexer.Position   { return o.Position }
func (o *OnViolationBlock) String() string {
	return fmt.Sprintf("on violation {\n  %s\n}", o.Handler.String())
}

// AssumeStmt represents an assumption.
// Example: assume environment == "dev"
type AssumeStmt struct {
	Position lexer.Position
	Guard    *GuardExpr
	Simple   string // for simple assumptions like "filesystem reliable"
}

func (a *AssumeStmt) statementNode()        {}
func (a *AssumeStmt) Pos() lexer.Position   { return a.Position }
func (a *AssumeStmt) String() string {
	if a.Guard != nil {
		return fmt.Sprintf("assume %s", a.Guard.String())
	}
	return fmt.Sprintf("assume %s", a.Simple)
}

// ParallelBlock represents a parallel execution block (v2 feature).
type ParallelBlock struct {
	Position   lexer.Position
	Statements []Statement
}

func (p *ParallelBlock) statementNode()        {}
func (p *ParallelBlock) Pos() lexer.Position   { return p.Position }
func (p *ParallelBlock) String() string {
	var out strings.Builder
	out.WriteString("parallel {\n")
	for _, s := range p.Statements {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}
