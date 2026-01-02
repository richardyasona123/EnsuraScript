// Package binder resolves implicit subjects and binds resources in the AST.
package binder

import (
	"fmt"

	"github.com/ensurascript/ensura/pkg/ast"
)

// ResourceTable holds declared resources and their aliases.
type ResourceTable struct {
	byPath  map[string]*ast.ResourceDecl
	byAlias map[string]*ast.ResourceDecl
}

// NewResourceTable creates a new resource table.
func NewResourceTable() *ResourceTable {
	return &ResourceTable{
		byPath:  make(map[string]*ast.ResourceDecl),
		byAlias: make(map[string]*ast.ResourceDecl),
	}
}

// Add adds a resource declaration to the table.
func (rt *ResourceTable) Add(decl *ast.ResourceDecl) error {
	key := fmt.Sprintf("%s:%s", decl.ResourceType, decl.Path)
	if existing, ok := rt.byPath[key]; ok {
		return fmt.Errorf("duplicate resource declaration: %s (first declared at %s)", key, existing.Position)
	}
	rt.byPath[key] = decl

	if decl.Alias != "" {
		if existing, ok := rt.byAlias[decl.Alias]; ok {
			return fmt.Errorf("duplicate alias: %s (first declared at %s)", decl.Alias, existing.Position)
		}
		rt.byAlias[decl.Alias] = decl
	}

	return nil
}

// Lookup looks up a resource by reference.
func (rt *ResourceTable) Lookup(ref *ast.ResourceRef) (*ast.ResourceDecl, bool) {
	if ref.Alias != "" {
		decl, ok := rt.byAlias[ref.Alias]
		return decl, ok
	}
	key := fmt.Sprintf("%s:%s", ref.ResourceType, ref.Path)
	decl, ok := rt.byPath[key]
	return decl, ok
}

// PolicyTable holds declared policies.
type PolicyTable struct {
	policies map[string]*ast.PolicyDecl
}

// NewPolicyTable creates a new policy table.
func NewPolicyTable() *PolicyTable {
	return &PolicyTable{
		policies: make(map[string]*ast.PolicyDecl),
	}
}

// Add adds a policy to the table.
func (pt *PolicyTable) Add(decl *ast.PolicyDecl) error {
	if existing, ok := pt.policies[decl.Name]; ok {
		return fmt.Errorf("duplicate policy: %s (first declared at %s)", decl.Name, existing.Position)
	}
	pt.policies[decl.Name] = decl
	return nil
}

// Lookup looks up a policy by name.
func (pt *PolicyTable) Lookup(name string) (*ast.PolicyDecl, bool) {
	decl, ok := pt.policies[name]
	return decl, ok
}

// Binder resolves implicit subjects and validates references.
type Binder struct {
	resources *ResourceTable
	policies  *PolicyTable
	errors    []string
}

// New creates a new Binder.
func New() *Binder {
	return &Binder{
		resources: NewResourceTable(),
		policies:  NewPolicyTable(),
	}
}

// Errors returns all binding errors.
func (b *Binder) Errors() []string {
	return b.errors
}

func (b *Binder) addError(pos interface{}, msg string) {
	b.errors = append(b.errors, fmt.Sprintf("%v: %s", pos, msg))
}

// Bind processes the AST and resolves implicit subjects.
func (b *Binder) Bind(program *ast.Program) *ast.Program {
	// First pass: collect all resource and policy declarations
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.ResourceDecl:
			if err := b.resources.Add(s); err != nil {
				b.addError(s.Position, err.Error())
			}
		case *ast.PolicyDecl:
			if err := b.policies.Add(s); err != nil {
				b.addError(s.Position, err.Error())
			}
		}
	}

	// Second pass: resolve implicit subjects and validate references
	var lastSubject *ast.ResourceRef
	var boundStatements []ast.Statement

	for _, stmt := range program.Statements {
		boundStmt := b.bindStatement(stmt, &lastSubject)
		if boundStmt != nil {
			boundStatements = append(boundStatements, boundStmt)
		}
	}

	program.Statements = boundStatements
	return program
}

func (b *Binder) bindStatement(stmt ast.Statement, lastSubject **ast.ResourceRef) ast.Statement {
	switch s := stmt.(type) {
	case *ast.ResourceDecl:
		return s

	case *ast.EnsureStmt:
		return b.bindEnsureStmt(s, lastSubject)

	case *ast.OnBlock:
		return b.bindOnBlock(s, lastSubject)

	case *ast.PolicyDecl:
		return b.bindPolicyDecl(s)

	case *ast.ApplyStmt:
		return b.bindApplyStmt(s, *lastSubject)

	case *ast.ForEachStmt:
		return b.bindForEachStmt(s)

	case *ast.InvariantBlock:
		return b.bindInvariantBlock(s)

	case *ast.OnViolationBlock:
		return s

	case *ast.AssumeStmt:
		return s

	case *ast.ParallelBlock:
		return b.bindParallelBlock(s)

	default:
		return stmt
	}
}

func (b *Binder) bindEnsureStmt(stmt *ast.EnsureStmt, lastSubject **ast.ResourceRef) *ast.EnsureStmt {
	// If no subject specified, inherit from last subject
	if stmt.Subject == nil {
		if *lastSubject == nil {
			b.addError(stmt.Position, "ensure statement has no subject and no implicit subject available")
			return nil
		}
		stmt.Subject = *lastSubject
	} else {
		// Validate and update last subject
		b.validateResourceRef(stmt.Subject)
		*lastSubject = stmt.Subject
	}

	// Validate handler if specified
	if stmt.Handler != nil {
		b.validateHandler(stmt.Handler, stmt.Condition)
	}

	return stmt
}

func (b *Binder) bindOnBlock(block *ast.OnBlock, lastSubject **ast.ResourceRef) *ast.OnBlock {
	// Validate the block's subject
	b.validateResourceRef(block.Subject)
	*lastSubject = block.Subject

	// Bind statements within the block with the block's subject as context
	var boundStatements []ast.Statement
	blockSubject := block.Subject

	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.EnsureStmt:
			if s.Subject == nil {
				s.Subject = blockSubject
			}
			boundStmt := b.bindEnsureStmt(s, &blockSubject)
			if boundStmt != nil {
				boundStatements = append(boundStatements, boundStmt)
			}
		case *ast.ApplyStmt:
			boundStmt := b.bindApplyStmt(s, blockSubject)
			if boundStmt != nil {
				boundStatements = append(boundStatements, boundStmt)
			}
		default:
			boundStatements = append(boundStatements, stmt)
		}
	}

	block.Statements = boundStatements
	return block
}

func (b *Binder) bindPolicyDecl(decl *ast.PolicyDecl) *ast.PolicyDecl {
	// Policies are validated but their statements are bound when applied
	return decl
}

func (b *Binder) bindApplyStmt(stmt *ast.ApplyStmt, currentSubject *ast.ResourceRef) *ast.ApplyStmt {
	// Validate policy exists
	policy, ok := b.policies.Lookup(stmt.PolicyName)
	if !ok {
		b.addError(stmt.Position, fmt.Sprintf("undefined policy: %s", stmt.PolicyName))
		return nil
	}

	// Validate argument count
	if len(stmt.Args) != len(policy.Params) {
		b.addError(stmt.Position, fmt.Sprintf("policy %s expects %d arguments, got %d",
			stmt.PolicyName, len(policy.Params), len(stmt.Args)))
		return nil
	}

	return stmt
}

func (b *Binder) bindForEachStmt(stmt *ast.ForEachStmt) *ast.ForEachStmt {
	// Validate container reference
	b.validateResourceRef(stmt.Container)

	// Bind statements - they will reference an implicit item variable
	var boundStatements []ast.Statement
	for _, s := range stmt.Statements {
		var dummy *ast.ResourceRef
		boundStmt := b.bindStatement(s, &dummy)
		if boundStmt != nil {
			boundStatements = append(boundStatements, boundStmt)
		}
	}
	stmt.Statements = boundStatements

	return stmt
}

func (b *Binder) bindInvariantBlock(block *ast.InvariantBlock) *ast.InvariantBlock {
	var boundStatements []ast.Statement
	for _, stmt := range block.Statements {
		var lastSubject *ast.ResourceRef
		boundStmt := b.bindStatement(stmt, &lastSubject)
		if boundStmt != nil {
			boundStatements = append(boundStatements, boundStmt)
		}
	}
	block.Statements = boundStatements
	return block
}

func (b *Binder) bindParallelBlock(block *ast.ParallelBlock) *ast.ParallelBlock {
	var boundStatements []ast.Statement
	for _, stmt := range block.Statements {
		var lastSubject *ast.ResourceRef
		boundStmt := b.bindStatement(stmt, &lastSubject)
		if boundStmt != nil {
			boundStatements = append(boundStatements, boundStmt)
		}
	}
	block.Statements = boundStatements
	return block
}

func (b *Binder) validateResourceRef(ref *ast.ResourceRef) {
	if ref == nil {
		return
	}

	// If it's an alias, look it up
	if ref.Alias != "" {
		if _, ok := b.resources.Lookup(ref); !ok {
			b.addError(ref.Position, fmt.Sprintf("undefined resource alias: %s", ref.Alias))
		}
	}
	// Inline references don't need to be declared (they're implicit declarations)
}

func (b *Binder) validateHandler(handler *ast.HandlerSpec, condition string) {
	// Handler validation is done by the runtime/adapter system
	// Here we just ensure basic syntax is correct
}

// ExpandPolicies expands all apply statements into their constituent ensure statements.
func (b *Binder) ExpandPolicies(program *ast.Program) *ast.Program {
	var expandedStatements []ast.Statement

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.OnBlock:
			expanded := b.expandOnBlock(s)
			expandedStatements = append(expandedStatements, expanded)
		default:
			expandedStatements = append(expandedStatements, stmt)
		}
	}

	program.Statements = expandedStatements
	return program
}

func (b *Binder) expandOnBlock(block *ast.OnBlock) *ast.OnBlock {
	var expandedStatements []ast.Statement

	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.ApplyStmt:
			expanded := b.expandApply(s, block.Subject)
			expandedStatements = append(expandedStatements, expanded...)
		default:
			expandedStatements = append(expandedStatements, stmt)
		}
	}

	block.Statements = expandedStatements
	return block
}

func (b *Binder) expandApply(apply *ast.ApplyStmt, subject *ast.ResourceRef) []ast.Statement {
	policy, ok := b.policies.Lookup(apply.PolicyName)
	if !ok {
		return nil
	}

	// Build parameter substitution map
	params := make(map[string]string)
	for i, param := range policy.Params {
		if i < len(apply.Args) {
			params[param.Name] = apply.Args[i]
		}
	}

	// Expand policy statements
	var expanded []ast.Statement
	for _, stmt := range policy.Statements {
		if ensure, ok := stmt.(*ast.EnsureStmt); ok {
			// Clone the ensure statement and substitute parameters
			newEnsure := &ast.EnsureStmt{
				Position:  apply.Position,
				Condition: ensure.Condition,
				Subject:   subject,
				Guard:     ensure.Guard,
				Requires:  ensure.Requires,
			}

			// Substitute handler parameters
			if ensure.Handler != nil {
				newHandler := &ast.HandlerSpec{
					Position: ensure.Handler.Position,
					Name:     ensure.Handler.Name,
					Args:     make(map[string]string),
				}
				for k, v := range ensure.Handler.Args {
					if subst, ok := params[v]; ok {
						newHandler.Args[k] = subst
					} else {
						newHandler.Args[k] = v
					}
				}
				newEnsure.Handler = newHandler
			}

			expanded = append(expanded, newEnsure)
		}
	}

	return expanded
}
