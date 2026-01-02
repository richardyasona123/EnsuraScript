// Package planner creates an executable plan from the guarantee graph.
package planner

import (
	"fmt"
	"strings"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/graph"
)

// Step represents a single step in the execution plan.
type Step struct {
	ID          string
	Guarantee   *graph.Guarantee
	Description string
	Handler     string
	HandlerArgs map[string]string
	IsInvariant bool
}

// Plan represents the complete execution plan.
type Plan struct {
	Steps           []*Step
	GlobalViolation *ast.ViolationHandler
}

// NewPlan creates a new empty plan.
func NewPlan() *Plan {
	return &Plan{}
}

// Planner creates execution plans from guarantee graphs.
type Planner struct {
	errors []string
}

// New creates a new Planner.
func New() *Planner {
	return &Planner{}
}

// Errors returns all planning errors.
func (p *Planner) Errors() []string {
	return p.errors
}

// CreatePlan creates an execution plan from a graph.
func (p *Planner) CreatePlan(g *graph.Graph, program *ast.Program) (*Plan, error) {
	plan := NewPlan()

	// Get topologically sorted guarantees
	sorted, err := g.TopoSort()
	if err != nil {
		// Report cycle if found
		cycle := g.FindCycle()
		if cycle != nil {
			return nil, fmt.Errorf("cyclic dependency detected: %s", strings.Join(cycle, " -> "))
		}
		return nil, err
	}

	// Convert guarantees to steps
	for _, guarantee := range sorted {
		step := p.createStep(guarantee, g.Invariants[guarantee.ID])
		plan.Steps = append(plan.Steps, step)
	}

	// Extract global violation handler
	plan.GlobalViolation = p.extractGlobalViolationHandler(program)

	return plan, nil
}

func (p *Planner) createStep(guarantee *graph.Guarantee, isInvariant bool) *Step {
	stmt := guarantee.Statement

	step := &Step{
		ID:          guarantee.ID,
		Guarantee:   guarantee,
		Description: p.generateDescription(stmt),
		IsInvariant: isInvariant,
	}

	// Extract handler information
	if stmt.Handler != nil {
		step.Handler = stmt.Handler.Name
		step.HandlerArgs = stmt.Handler.Args
	} else {
		// Use default handler based on condition
		step.Handler = p.getDefaultHandler(stmt.Condition)
		step.HandlerArgs = make(map[string]string)
	}

	return step
}

func (p *Planner) generateDescription(stmt *ast.EnsureStmt) string {
	var parts []string
	parts = append(parts, "Ensure", stmt.Condition)

	if stmt.Subject != nil {
		parts = append(parts, "on", stmt.Subject.String())
	}

	if stmt.Handler != nil {
		parts = append(parts, "with", stmt.Handler.Name)
	}

	return strings.Join(parts, " ")
}

func (p *Planner) getDefaultHandler(condition string) string {
	defaults := map[string]string{
		"exists":      "fs.native",
		"readable":    "fs.native",
		"writable":    "fs.native",
		"encrypted":   "AES:256",
		"permissions": "posix",
		"checksum":    "fs.native",
		"content":     "fs.native",
		"running":     "process.native",
		"stopped":     "process.native",
		"listening":   "service.native",
		"healthy":     "service.native",
		"reachable":   "http.get",
		"status_code": "http.get",
		"tls":         "http.get",
		"scheduled":   "cron.native",
		"backed_up":   "backup.native",
		"stable":      "db.native",
	}

	if handler, ok := defaults[condition]; ok {
		return handler
	}
	return ""
}

func (p *Planner) extractGlobalViolationHandler(program *ast.Program) *ast.ViolationHandler {
	for _, stmt := range program.Statements {
		if v, ok := stmt.(*ast.OnViolationBlock); ok {
			return v.Handler
		}
	}
	return nil
}

// String returns a human-readable representation of the plan.
func (p *Plan) String() string {
	var out strings.Builder

	out.WriteString("Execution Plan\n")
	out.WriteString("==============\n\n")

	for i, step := range p.Steps {
		marker := "  "
		if step.IsInvariant {
			marker = "! "
		}
		out.WriteString(fmt.Sprintf("%s%d. %s\n", marker, i+1, step.Description))
		out.WriteString(fmt.Sprintf("      Handler: %s\n", step.Handler))
		if len(step.HandlerArgs) > 0 {
			out.WriteString("      Args:\n")
			for k, v := range step.HandlerArgs {
				out.WriteString(fmt.Sprintf("        %s: %s\n", k, v))
			}
		}
	}

	if p.GlobalViolation != nil {
		out.WriteString("\nGlobal Violation Handler\n")
		out.WriteString("------------------------\n")
		if p.GlobalViolation.Retry > 0 {
			out.WriteString(fmt.Sprintf("  Retry: %d times\n", p.GlobalViolation.Retry))
		}
		for _, n := range p.GlobalViolation.Notify {
			out.WriteString(fmt.Sprintf("  Notify: %s\n", n))
		}
	}

	return out.String()
}

// ToJSON returns a JSON-compatible structure for the plan.
func (p *Plan) ToJSON() map[string]interface{} {
	steps := make([]map[string]interface{}, len(p.Steps))
	for i, step := range p.Steps {
		steps[i] = map[string]interface{}{
			"id":          step.ID,
			"description": step.Description,
			"handler":     step.Handler,
			"args":        step.HandlerArgs,
			"isInvariant": step.IsInvariant,
		}
	}

	result := map[string]interface{}{
		"steps": steps,
	}

	if p.GlobalViolation != nil {
		result["globalViolation"] = map[string]interface{}{
			"retry":  p.GlobalViolation.Retry,
			"notify": p.GlobalViolation.Notify,
		}
	}

	return result
}
