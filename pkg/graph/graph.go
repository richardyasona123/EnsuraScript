// Package graph builds and sorts the dependency graph for guarantees.
package graph

import (
	"fmt"
	"sort"

	"github.com/ensurascript/ensura/pkg/ast"
)

// Guarantee represents a single guarantee node in the graph.
type Guarantee struct {
	ID        string
	Statement *ast.EnsureStmt
	Priority  int // higher priority = more important (invariants get higher priority)
	IsImplied bool
}

// Edge represents a dependency edge in the graph.
type Edge struct {
	From string // guarantee ID
	To   string // guarantee ID
	Type string // "requires", "after", "implies"
}

// Graph represents the dependency graph of guarantees.
type Graph struct {
	Nodes      map[string]*Guarantee
	Edges      []Edge
	Invariants map[string]bool // set of guarantee IDs from invariant blocks
	errors     []string
}

// NewGraph creates a new empty graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes:      make(map[string]*Guarantee),
		Invariants: make(map[string]bool),
	}
}

// Errors returns all graph building errors.
func (g *Graph) Errors() []string {
	return g.errors
}

// Build constructs the dependency graph from the AST.
func Build(program *ast.Program) *Graph {
	g := NewGraph()
	g.buildFromStatements(program.Statements, false, 0)
	g.buildImplicitEdges()
	return g
}

func (g *Graph) buildFromStatements(statements []ast.Statement, isInvariant bool, basePriority int) {
	for _, stmt := range statements {
		g.processStatement(stmt, isInvariant, basePriority)
	}
}

func (g *Graph) processStatement(stmt ast.Statement, isInvariant bool, basePriority int) {
	switch s := stmt.(type) {
	case *ast.EnsureStmt:
		g.addGuarantee(s, isInvariant, basePriority)
	case *ast.OnBlock:
		g.buildFromStatements(s.Statements, isInvariant, basePriority)
	case *ast.InvariantBlock:
		// Invariants have higher priority
		g.buildFromStatements(s.Statements, true, basePriority+1000)
	case *ast.ForEachStmt:
		// For-each statements are handled at runtime
		// but we still need to process their templates
		g.buildFromStatements(s.Statements, isInvariant, basePriority)
	case *ast.ParallelBlock:
		g.buildFromStatements(s.Statements, isInvariant, basePriority)
	}
}

func (g *Graph) addGuarantee(stmt *ast.EnsureStmt, isInvariant bool, priority int) {
	id := g.generateID(stmt)

	guarantee := &Guarantee{
		ID:        id,
		Statement: stmt,
		Priority:  priority,
	}

	g.Nodes[id] = guarantee

	if isInvariant {
		g.Invariants[id] = true
	}

	// Add explicit dependency edges
	for _, req := range stmt.Requires {
		// Find the guarantee for this required condition on the same subject
		reqID := g.findGuaranteeByCondition(req, stmt.Subject)
		if reqID != "" {
			g.Edges = append(g.Edges, Edge{From: reqID, To: id, Type: "requires"})
		}
	}

	// Add after/before edges
	for _, after := range stmt.After {
		// Find guarantees on the referenced resource
		afterIDs := g.findGuaranteesByResource(after)
		for _, afterID := range afterIDs {
			g.Edges = append(g.Edges, Edge{From: afterID, To: id, Type: "after"})
		}
	}

	for _, before := range stmt.Before {
		beforeIDs := g.findGuaranteesByResource(before)
		for _, beforeID := range beforeIDs {
			g.Edges = append(g.Edges, Edge{From: id, To: beforeID, Type: "before"})
		}
	}
}

func (g *Graph) generateID(stmt *ast.EnsureStmt) string {
	subject := ""
	if stmt.Subject != nil {
		subject = stmt.Subject.String()
	}
	return fmt.Sprintf("%s:%s@%s", stmt.Condition, subject, stmt.Position)
}

func (g *Graph) findGuaranteeByCondition(condition string, subject *ast.ResourceRef) string {
	subjectStr := ""
	if subject != nil {
		subjectStr = subject.String()
	}

	for id, guarantee := range g.Nodes {
		if guarantee.Statement.Condition == condition {
			guardSubject := ""
			if guarantee.Statement.Subject != nil {
				guardSubject = guarantee.Statement.Subject.String()
			}
			if guardSubject == subjectStr {
				return id
			}
		}
	}
	return ""
}

func (g *Graph) findGuaranteesByResource(ref *ast.ResourceRef) []string {
	var ids []string
	refStr := ref.String()

	for id, guarantee := range g.Nodes {
		if guarantee.Statement.Subject != nil && guarantee.Statement.Subject.String() == refStr {
			ids = append(ids, id)
		}
	}

	return ids
}

// buildImplicitEdges adds edges for implied conditions.
func (g *Graph) buildImplicitEdges() {
	// Group guarantees by subject
	bySubject := make(map[string][]*Guarantee)
	for _, guarantee := range g.Nodes {
		subject := ""
		if guarantee.Statement.Subject != nil {
			subject = guarantee.Statement.Subject.String()
		}
		bySubject[subject] = append(bySubject[subject], guarantee)
	}

	// For each subject, create edges based on condition implications
	impliedBy := map[string][]string{
		"encrypted":   {"exists", "readable", "writable"},
		"permissions": {"exists"},
		"readable":    {"exists"},
		"writable":    {"exists"},
		"checksum":    {"exists", "readable"},
		"content":     {"exists"},
		"listening":   {"running"},
		"healthy":     {"running"},
		"status_code": {"reachable"},
		"tls":         {"reachable"},
		"backed_up":   {"exists"},
	}

	for _, guarantees := range bySubject {
		conditionToID := make(map[string]string)
		for _, g := range guarantees {
			conditionToID[g.Statement.Condition] = g.ID
		}

		for _, guarantee := range guarantees {
			implies, ok := impliedBy[guarantee.Statement.Condition]
			if !ok {
				continue
			}

			for _, implied := range implies {
				if impliedID, exists := conditionToID[implied]; exists {
					// The implied condition must be satisfied before this one
					g.Edges = append(g.Edges, Edge{
						From: impliedID,
						To:   guarantee.ID,
						Type: "implies",
					})
				}
			}
		}
	}
}

// TopoSort returns guarantees in topologically sorted order.
func (g *Graph) TopoSort() ([]*Guarantee, error) {
	// Build adjacency list and in-degree map
	adj := make(map[string][]string)
	inDegree := make(map[string]int)

	for id := range g.Nodes {
		adj[id] = nil
		inDegree[id] = 0
	}

	for _, edge := range g.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		inDegree[edge.To]++
	}

	// Kahn's algorithm with priority-based tie-breaking
	var queue []*Guarantee
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, g.Nodes[id])
		}
	}

	// Sort initial queue by priority (descending) then ID (for stability)
	sortQueue := func(q []*Guarantee) {
		sort.Slice(q, func(i, j int) bool {
			if q[i].Priority != q[j].Priority {
				return q[i].Priority > q[j].Priority
			}
			return q[i].ID < q[j].ID
		})
	}
	sortQueue(queue)

	var result []*Guarantee
	for len(queue) > 0 {
		// Take highest priority node
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Decrease in-degree of neighbors
		for _, neighbor := range adj[node.ID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, g.Nodes[neighbor])
				sortQueue(queue)
			}
		}
	}

	// Check for cycles
	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return result, nil
}

// FindCycle finds a cycle in the graph if one exists.
func (g *Graph) FindCycle() []string {
	// Build adjacency list
	adj := make(map[string][]string)
	for _, edge := range g.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
	}

	// DFS to find cycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	parent := make(map[string]string)

	var cycle []string
	var findCycle func(node string) bool
	findCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range adj[node] {
			if !visited[neighbor] {
				parent[neighbor] = node
				if findCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Found cycle, reconstruct it
				cycle = []string{neighbor}
				for cur := node; cur != neighbor; cur = parent[cur] {
					cycle = append([]string{cur}, cycle...)
				}
				cycle = append([]string{neighbor}, cycle...)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for id := range g.Nodes {
		if !visited[id] {
			if findCycle(id) {
				return cycle
			}
		}
	}

	return nil
}

// Visualize returns a DOT graph representation for debugging.
func (g *Graph) Visualize() string {
	var out string
	out += "digraph G {\n"
	out += "  rankdir=TB;\n"

	for id, node := range g.Nodes {
		label := node.Statement.Condition
		if node.Statement.Subject != nil {
			label += "\\n" + node.Statement.Subject.String()
		}
		shape := "box"
		if g.Invariants[id] {
			shape = "doublebox"
		}
		out += fmt.Sprintf("  %q [label=%q, shape=%s];\n", id, label, shape)
	}

	for _, edge := range g.Edges {
		style := "solid"
		if edge.Type == "implies" {
			style = "dashed"
		}
		out += fmt.Sprintf("  %q -> %q [style=%s, label=%q];\n", edge.From, edge.To, style, edge.Type)
	}

	out += "}\n"
	return out
}
