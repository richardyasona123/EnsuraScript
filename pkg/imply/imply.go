// Package imply handles implication expansion for EnsuraScript conditions.
package imply

import (
	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/lexer"
)

// ConditionMeta defines metadata for a condition.
type ConditionMeta struct {
	Name            string
	ApplicableTypes []string          // resource types this condition applies to
	Implies         []string          // conditions this implies
	Conflicts       []string          // conditions this conflicts with
	DefaultHandler  string            // default handler if none specified
}

// Registry holds all known conditions and their metadata.
type Registry struct {
	conditions map[string]*ConditionMeta
}

// NewRegistry creates a new condition registry with built-in conditions.
func NewRegistry() *Registry {
	r := &Registry{
		conditions: make(map[string]*ConditionMeta),
	}
	r.registerBuiltins()
	return r
}

func (r *Registry) registerBuiltins() {
	// Filesystem conditions
	r.Register(&ConditionMeta{
		Name:            "exists",
		ApplicableTypes: []string{"file", "directory"},
		Implies:         nil,
		Conflicts:       nil,
		DefaultHandler:  "fs.native",
	})

	r.Register(&ConditionMeta{
		Name:            "readable",
		ApplicableTypes: []string{"file"},
		Implies:         []string{"exists"},
		Conflicts:       nil,
		DefaultHandler:  "fs.native",
	})

	r.Register(&ConditionMeta{
		Name:            "writable",
		ApplicableTypes: []string{"file"},
		Implies:         []string{"exists"},
		Conflicts:       nil,
		DefaultHandler:  "fs.native",
	})

	r.Register(&ConditionMeta{
		Name:            "encrypted",
		ApplicableTypes: []string{"file"},
		Implies:         []string{"exists", "readable", "writable"},
		Conflicts:       []string{"unencrypted"},
		DefaultHandler:  "AES:256",
	})

	r.Register(&ConditionMeta{
		Name:            "unencrypted",
		ApplicableTypes: []string{"file"},
		Implies:         []string{"exists"},
		Conflicts:       []string{"encrypted"},
		DefaultHandler:  "",
	})

	r.Register(&ConditionMeta{
		Name:            "permissions",
		ApplicableTypes: []string{"file", "directory"},
		Implies:         []string{"exists"},
		Conflicts:       nil,
		DefaultHandler:  "posix",
	})

	r.Register(&ConditionMeta{
		Name:            "checksum",
		ApplicableTypes: []string{"file"},
		Implies:         []string{"exists", "readable"},
		Conflicts:       nil,
		DefaultHandler:  "fs.native",
	})

	r.Register(&ConditionMeta{
		Name:            "content",
		ApplicableTypes: []string{"file"},
		Implies:         []string{"exists"},
		Conflicts:       nil,
		DefaultHandler:  "fs.native",
	})

	// Process/Service conditions
	r.Register(&ConditionMeta{
		Name:            "running",
		ApplicableTypes: []string{"process", "service"},
		Implies:         nil,
		Conflicts:       []string{"stopped"},
		DefaultHandler:  "process.native",
	})

	r.Register(&ConditionMeta{
		Name:            "stopped",
		ApplicableTypes: []string{"process", "service"},
		Implies:         nil,
		Conflicts:       []string{"running"},
		DefaultHandler:  "process.native",
	})

	r.Register(&ConditionMeta{
		Name:            "listening",
		ApplicableTypes: []string{"service"},
		Implies:         []string{"running"},
		Conflicts:       nil,
		DefaultHandler:  "service.native",
	})

	r.Register(&ConditionMeta{
		Name:            "healthy",
		ApplicableTypes: []string{"service"},
		Implies:         []string{"running"},
		Conflicts:       nil,
		DefaultHandler:  "service.native",
	})

	// HTTP conditions
	r.Register(&ConditionMeta{
		Name:            "reachable",
		ApplicableTypes: []string{"http"},
		Implies:         nil,
		Conflicts:       nil,
		DefaultHandler:  "http.get",
	})

	r.Register(&ConditionMeta{
		Name:            "status_code",
		ApplicableTypes: []string{"http"},
		Implies:         []string{"reachable"},
		Conflicts:       nil,
		DefaultHandler:  "http.get",
	})

	r.Register(&ConditionMeta{
		Name:            "tls",
		ApplicableTypes: []string{"http"},
		Implies:         []string{"reachable"},
		Conflicts:       nil,
		DefaultHandler:  "http.get",
	})

	// Scheduling conditions
	r.Register(&ConditionMeta{
		Name:            "scheduled",
		ApplicableTypes: []string{"cron"},
		Implies:         nil,
		Conflicts:       nil,
		DefaultHandler:  "cron.native",
	})

	// Backup conditions
	r.Register(&ConditionMeta{
		Name:            "backed_up",
		ApplicableTypes: []string{"file", "database"},
		Implies:         []string{"exists"},
		Conflicts:       nil,
		DefaultHandler:  "backup.native",
	})

	// Database conditions
	r.Register(&ConditionMeta{
		Name:            "stable",
		ApplicableTypes: []string{"database"},
		Implies:         nil,
		Conflicts:       nil,
		DefaultHandler:  "db.native",
	})
}

// Register adds a condition to the registry.
func (r *Registry) Register(meta *ConditionMeta) {
	r.conditions[meta.Name] = meta
}

// Get retrieves condition metadata.
func (r *Registry) Get(name string) (*ConditionMeta, bool) {
	meta, ok := r.conditions[name]
	return meta, ok
}

// Expander handles implication expansion.
type Expander struct {
	registry *Registry
	errors   []string
}

// NewExpander creates a new implication expander.
func NewExpander() *Expander {
	return &Expander{
		registry: NewRegistry(),
	}
}

// Errors returns all expansion errors.
func (e *Expander) Errors() []string {
	return e.errors
}

// Expand expands all implied conditions in the program.
func (e *Expander) Expand(program *ast.Program) *ast.Program {
	var expandedStatements []ast.Statement

	for _, stmt := range program.Statements {
		expanded := e.expandStatement(stmt)
		expandedStatements = append(expandedStatements, expanded...)
	}

	// Deduplicate guarantees
	program.Statements = e.deduplicate(expandedStatements)
	return program
}

func (e *Expander) expandStatement(stmt ast.Statement) []ast.Statement {
	switch s := stmt.(type) {
	case *ast.EnsureStmt:
		return e.expandEnsure(s)
	case *ast.OnBlock:
		return []ast.Statement{e.expandOnBlock(s)}
	case *ast.InvariantBlock:
		return []ast.Statement{e.expandInvariantBlock(s)}
	case *ast.ForEachStmt:
		return []ast.Statement{e.expandForEachStmt(s)}
	case *ast.ParallelBlock:
		return []ast.Statement{e.expandParallelBlock(s)}
	default:
		return []ast.Statement{stmt}
	}
}

func (e *Expander) expandEnsure(stmt *ast.EnsureStmt) []ast.Statement {
	var result []ast.Statement

	// Get the condition metadata
	meta, ok := e.registry.Get(stmt.Condition)
	if !ok {
		// Unknown condition - just return as-is
		return []ast.Statement{stmt}
	}

	// Validate resource type
	if stmt.Subject != nil && stmt.Subject.ResourceType != "" {
		valid := false
		for _, t := range meta.ApplicableTypes {
			if t == stmt.Subject.ResourceType {
				valid = true
				break
			}
		}
		if !valid {
			e.errors = append(e.errors,
				stmt.Position.String()+": condition '"+stmt.Condition+
				"' is not applicable to resource type '"+stmt.Subject.ResourceType+"'")
		}
	}

	// Expand implied conditions first (they must be satisfied before this one)
	for _, implied := range meta.Implies {
		impliedStmt := &ast.EnsureStmt{
			Position:  stmt.Position,
			Condition: implied,
			Subject:   stmt.Subject,
			Guard:     stmt.Guard,
		}
		// Recursively expand implied conditions
		result = append(result, e.expandEnsure(impliedStmt)...)
	}

	// Add the original statement
	result = append(result, stmt)

	return result
}

func (e *Expander) expandOnBlock(block *ast.OnBlock) *ast.OnBlock {
	var expandedStatements []ast.Statement

	for _, stmt := range block.Statements {
		expanded := e.expandStatement(stmt)
		expandedStatements = append(expandedStatements, expanded...)
	}

	block.Statements = expandedStatements
	return block
}

func (e *Expander) expandInvariantBlock(block *ast.InvariantBlock) *ast.InvariantBlock {
	var expandedStatements []ast.Statement

	for _, stmt := range block.Statements {
		expanded := e.expandStatement(stmt)
		expandedStatements = append(expandedStatements, expanded...)
	}

	block.Statements = expandedStatements
	return block
}

func (e *Expander) expandForEachStmt(stmt *ast.ForEachStmt) *ast.ForEachStmt {
	var expandedStatements []ast.Statement

	for _, s := range stmt.Statements {
		expanded := e.expandStatement(s)
		expandedStatements = append(expandedStatements, expanded...)
	}

	stmt.Statements = expandedStatements
	return stmt
}

func (e *Expander) expandParallelBlock(block *ast.ParallelBlock) *ast.ParallelBlock {
	var expandedStatements []ast.Statement

	for _, stmt := range block.Statements {
		expanded := e.expandStatement(stmt)
		expandedStatements = append(expandedStatements, expanded...)
	}

	block.Statements = expandedStatements
	return block
}

// deduplicate removes duplicate guarantees.
func (e *Expander) deduplicate(statements []ast.Statement) []ast.Statement {
	seen := make(map[string]bool)
	var result []ast.Statement

	for _, stmt := range statements {
		key := e.statementKey(stmt)
		if key == "" || !seen[key] {
			result = append(result, stmt)
			if key != "" {
				seen[key] = true
			}
		}
	}

	return result
}

func (e *Expander) statementKey(stmt ast.Statement) string {
	if ensure, ok := stmt.(*ast.EnsureStmt); ok {
		if ensure.Subject != nil {
			return ensure.Condition + ":" + ensure.Subject.String()
		}
		return ensure.Condition
	}
	return ""
}

// CheckConflicts checks for conflicting conditions.
func (e *Expander) CheckConflicts(program *ast.Program) []string {
	var conflicts []string

	// Collect all ensure statements by subject
	bySubject := make(map[string][]*ast.EnsureStmt)
	e.collectEnsures(program.Statements, bySubject)

	// Check for conflicts within each subject
	for subject, ensures := range bySubject {
		conditions := make(map[string]lexer.Position)
		for _, ensure := range ensures {
			conditions[ensure.Condition] = ensure.Position

			// Check if this condition conflicts with any previously seen
			meta, ok := e.registry.Get(ensure.Condition)
			if !ok {
				continue
			}

			for _, conflict := range meta.Conflicts {
				if pos, exists := conditions[conflict]; exists {
					conflicts = append(conflicts,
						ensure.Position.String()+": '"+ensure.Condition+
						"' conflicts with '"+conflict+"' on "+subject+
						" (declared at "+pos.String()+")")
				}
			}
		}
	}

	return conflicts
}

func (e *Expander) collectEnsures(statements []ast.Statement, bySubject map[string][]*ast.EnsureStmt) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.EnsureStmt:
			key := ""
			if s.Subject != nil {
				key = s.Subject.String()
			}
			bySubject[key] = append(bySubject[key], s)
		case *ast.OnBlock:
			e.collectEnsures(s.Statements, bySubject)
		case *ast.InvariantBlock:
			e.collectEnsures(s.Statements, bySubject)
		case *ast.ForEachStmt:
			e.collectEnsures(s.Statements, bySubject)
		case *ast.ParallelBlock:
			e.collectEnsures(s.Statements, bySubject)
		}
	}
}
