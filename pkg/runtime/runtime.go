// Package runtime implements the EnsuraScript enforcement loop.
package runtime

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/planner"
)

// HandlerResult represents the result of a handler check or enforce operation.
type HandlerResult struct {
	Success bool
	Message string
	Error   error
}

// Handler is the interface that all handlers must implement.
type Handler interface {
	Name() string
	Check(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) HandlerResult
	Enforce(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) HandlerResult
}

// HandlerRegistry holds all registered handlers.
type HandlerRegistry struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

// NewHandlerRegistry creates a new handler registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]Handler),
	}
}

// Register adds a handler to the registry.
func (r *HandlerRegistry) Register(h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[h.Name()] = h
}

// Get retrieves a handler by name.
func (r *HandlerRegistry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[name]
	return h, ok
}

// StepStatus represents the status of a step execution.
type StepStatus int

const (
	StepPending StepStatus = iota
	StepSatisfied
	StepViolated
	StepRepaired
	StepFailed
)

func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepSatisfied:
		return "satisfied"
	case StepViolated:
		return "violated"
	case StepRepaired:
		return "repaired"
	case StepFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// StepResult contains the result of executing a step.
type StepResult struct {
	Step     *planner.Step
	Status   StepStatus
	Attempts int
	Message  string
	Error    error
}

// RunResult contains the result of a complete run.
type RunResult struct {
	StartTime     time.Time
	EndTime       time.Time
	Steps         []*StepResult
	AllSatisfied  bool
	TotalChecks   int
	TotalRepairs  int
	TotalFailures int
}

// Config holds runtime configuration.
type Config struct {
	Interval   time.Duration // time between enforcement loops
	MaxRetries int           // default max retries per step
	DryRun     bool          // if true, only check without enforcing
	CheckOnly  bool          // if true, run once and exit
	Redact     bool          // if true, redact secrets in logs
	Logger     io.Writer     // log output
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Interval:   30 * time.Second,
		MaxRetries: 3,
		DryRun:     false,
		CheckOnly:  false,
		Redact:     true,
		Logger:     os.Stdout,
	}
}

// Runtime executes the enforcement loop.
type Runtime struct {
	config   *Config
	registry *HandlerRegistry
	plan     *planner.Plan
	mu       sync.Mutex
}

// New creates a new Runtime.
func New(plan *planner.Plan, registry *HandlerRegistry, config *Config) *Runtime {
	if config == nil {
		config = DefaultConfig()
	}
	return &Runtime{
		config:   config,
		registry: registry,
		plan:     plan,
	}
}

// Run executes the enforcement loop.
func (r *Runtime) Run(ctx context.Context) error {
	if r.config.CheckOnly {
		result := r.runOnce(ctx)
		r.printResult(result)
		if !result.AllSatisfied {
			return fmt.Errorf("one or more guarantees not satisfied")
		}
		return nil
	}

	// Continuous loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			result := r.runOnce(ctx)
			r.printResult(result)

			// Wait for next interval
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.config.Interval):
			}
		}
	}
}

// RunOnce executes a single enforcement pass.
func (r *Runtime) RunOnce(ctx context.Context) *RunResult {
	return r.runOnce(ctx)
}

func (r *Runtime) runOnce(ctx context.Context) *RunResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := &RunResult{
		StartTime: time.Now(),
		Steps:     make([]*StepResult, 0, len(r.plan.Steps)),
	}

	allSatisfied := true

	for _, step := range r.plan.Steps {
		stepResult := r.executeStep(ctx, step)
		result.Steps = append(result.Steps, stepResult)
		result.TotalChecks++

		switch stepResult.Status {
		case StepSatisfied:
			// Continue to next step
		case StepRepaired:
			result.TotalRepairs++
		case StepViolated, StepFailed:
			allSatisfied = false
			result.TotalFailures++
			// For sequential execution, we continue but track failures
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.EndTime = time.Now()
			result.AllSatisfied = false
			return result
		default:
		}
	}

	result.EndTime = time.Now()
	result.AllSatisfied = allSatisfied
	return result
}

func (r *Runtime) executeStep(ctx context.Context, step *planner.Step) *StepResult {
	result := &StepResult{
		Step: step,
	}

	// Get handler
	handler, ok := r.registry.Get(step.Handler)
	if !ok {
		result.Status = StepFailed
		result.Error = fmt.Errorf("handler not found: %s", step.Handler)
		return result
	}

	// Get subject
	subject := step.Guarantee.Statement.Subject

	// Check
	checkResult := handler.Check(ctx, subject, step.Guarantee.Statement.Condition, step.HandlerArgs)
	result.Attempts++

	if checkResult.Success {
		result.Status = StepSatisfied
		result.Message = checkResult.Message
		return result
	}

	// Violated - attempt repair if not dry run
	result.Status = StepViolated
	result.Message = checkResult.Message

	if r.config.DryRun {
		return result
	}

	// Get retry count
	maxRetries := r.config.MaxRetries
	if step.Guarantee.Statement.ViolationHandler != nil && step.Guarantee.Statement.ViolationHandler.Retry > 0 {
		maxRetries = step.Guarantee.Statement.ViolationHandler.Retry
	} else if r.plan.GlobalViolation != nil && r.plan.GlobalViolation.Retry > 0 {
		maxRetries = r.plan.GlobalViolation.Retry
	}

	// Attempt repair with retries
	for attempt := 0; attempt < maxRetries; attempt++ {
		result.Attempts++

		enforceResult := handler.Enforce(ctx, subject, step.Guarantee.Statement.Condition, step.HandlerArgs)
		if enforceResult.Error != nil {
			result.Error = enforceResult.Error
			continue
		}

		// Re-check
		checkResult = handler.Check(ctx, subject, step.Guarantee.Statement.Condition, step.HandlerArgs)
		if checkResult.Success {
			result.Status = StepRepaired
			result.Message = "repaired after " + fmt.Sprintf("%d", attempt+1) + " attempts"
			return result
		}
	}

	result.Status = StepFailed
	result.Message = fmt.Sprintf("failed after %d repair attempts", maxRetries)
	return result
}

func (r *Runtime) printResult(result *RunResult) {
	w := r.config.Logger
	if w == nil {
		return
	}

	duration := result.EndTime.Sub(result.StartTime)

	fmt.Fprintf(w, "\n[%s] Enforcement run completed in %v\n",
		result.EndTime.Format(time.RFC3339), duration)
	fmt.Fprintf(w, "  Checks: %d, Repairs: %d, Failures: %d\n",
		result.TotalChecks, result.TotalRepairs, result.TotalFailures)

	if result.AllSatisfied {
		fmt.Fprintf(w, "  Status: ALL SATISFIED\n")
	} else {
		fmt.Fprintf(w, "  Status: VIOLATIONS DETECTED\n")
		for _, step := range result.Steps {
			if step.Status == StepViolated || step.Status == StepFailed {
				fmt.Fprintf(w, "    - %s: %s\n", step.Step.Description, step.Status)
				if step.Message != "" {
					fmt.Fprintf(w, "      Message: %s\n", step.Message)
				}
				if step.Error != nil {
					fmt.Fprintf(w, "      Error: %v\n", step.Error)
				}
			}
		}
	}
}

// Check runs a check-only pass without enforcement.
func (r *Runtime) Check(ctx context.Context) *RunResult {
	r.config.DryRun = true
	defer func() { r.config.DryRun = false }()
	return r.runOnce(ctx)
}
