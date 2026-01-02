// Package main provides the ensura CLI tool.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ensurascript/ensura/pkg/adapters"
	"github.com/ensurascript/ensura/pkg/binder"
	"github.com/ensurascript/ensura/pkg/graph"
	"github.com/ensurascript/ensura/pkg/imply"
	"github.com/ensurascript/ensura/pkg/parser"
	"github.com/ensurascript/ensura/pkg/planner"
	"github.com/ensurascript/ensura/pkg/runtime"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "compile":
		cmdCompile(os.Args[2:])
	case "explain":
		cmdExplain(os.Args[2:])
	case "plan":
		cmdPlan(os.Args[2:])
	case "run":
		cmdRun(os.Args[2:])
	case "check":
		cmdCheck(os.Args[2:])
	case "version":
		fmt.Printf("ensura version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`ensura - Programming by guarantees, not instructions.

Usage:
  ensura <command> [options] <file.ens>

Commands:
  compile   Validate and print the resolved guarantee graph
  explain   Show implied guarantees and chosen handlers
  plan      Print the deterministic sequential execution plan
  run       Run the continuous enforcement loop
  check     Check guarantees without enforcing (dry run)
  version   Print version information
  help      Show this help message

Options:
  -interval duration   Interval between enforcement loops (default 30s)
  -retries int         Maximum retries per step (default 3)
  -json                Output in JSON format
  -graph               Output dependency graph in DOT format

Examples:
  ensura compile config.ens
  ensura run config.ens -interval 60s
  ensura check config.ens`)
}

func loadAndCompile(filename string) (*compileResult, error) {
	// Read source file
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse
	program, parseErrors := parser.ParseFile(string(source), filename)
	if len(parseErrors) > 0 {
		for _, e := range parseErrors {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", e)
		}
		return nil, fmt.Errorf("parsing failed with %d errors", len(parseErrors))
	}

	// Bind
	b := binder.New()
	program = b.Bind(program)
	if len(b.Errors()) > 0 {
		for _, e := range b.Errors() {
			fmt.Fprintf(os.Stderr, "Binding error: %s\n", e)
		}
		return nil, fmt.Errorf("binding failed with %d errors", len(b.Errors()))
	}

	// Expand policies
	program = b.ExpandPolicies(program)

	// Expand implications
	expander := imply.NewExpander()
	program = expander.Expand(program)
	if len(expander.Errors()) > 0 {
		for _, e := range expander.Errors() {
			fmt.Fprintf(os.Stderr, "Expansion error: %s\n", e)
		}
		return nil, fmt.Errorf("expansion failed with %d errors", len(expander.Errors()))
	}

	// Check conflicts
	conflicts := expander.CheckConflicts(program)
	if len(conflicts) > 0 {
		for _, c := range conflicts {
			fmt.Fprintf(os.Stderr, "Conflict: %s\n", c)
		}
		return nil, fmt.Errorf("found %d conflicting conditions", len(conflicts))
	}

	// Build graph
	g := graph.Build(program)
	if len(g.Errors()) > 0 {
		for _, e := range g.Errors() {
			fmt.Fprintf(os.Stderr, "Graph error: %s\n", e)
		}
		return nil, fmt.Errorf("graph building failed with %d errors", len(g.Errors()))
	}

	// Create plan
	p := planner.New()
	plan, err := p.CreatePlan(g, program)
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}

	return &compileResult{
		graph: g,
		plan:  plan,
	}, nil
}

type compileResult struct {
	graph *graph.Graph
	plan  *planner.Plan
}

func cmdCompile(args []string) {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	graphOutput := fs.Bool("graph", false, "Output dependency graph in DOT format")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ensura compile [options] <file.ens>")
		os.Exit(1)
	}

	result, err := loadAndCompile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *graphOutput {
		fmt.Println(result.graph.Visualize())
		return
	}

	if *jsonOutput {
		output := result.plan.ToJSON()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
		return
	}

	fmt.Println("Compilation successful!")
	fmt.Printf("  Guarantees: %d\n", len(result.graph.Nodes))
	fmt.Printf("  Dependencies: %d\n", len(result.graph.Edges))
	fmt.Printf("  Plan steps: %d\n", len(result.plan.Steps))
}

func cmdExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ensura explain [options] <file.ens>")
		os.Exit(1)
	}

	result, err := loadAndCompile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		explanations := make([]map[string]interface{}, 0)
		for _, step := range result.plan.Steps {
			exp := map[string]interface{}{
				"condition": step.Guarantee.Statement.Condition,
				"handler":   step.Handler,
				"args":      step.HandlerArgs,
			}
			if step.Guarantee.Statement.Subject != nil {
				exp["subject"] = step.Guarantee.Statement.Subject.String()
			}
			if step.Guarantee.IsImplied {
				exp["implied"] = true
			}
			if step.IsInvariant {
				exp["invariant"] = true
			}
			explanations = append(explanations, exp)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(explanations)
		return
	}

	fmt.Println("Guarantee Explanations")
	fmt.Println("======================")
	fmt.Println()

	for i, step := range result.plan.Steps {
		marker := ""
		if step.IsInvariant {
			marker = " [INVARIANT]"
		}
		if step.Guarantee.IsImplied {
			marker += " [IMPLIED]"
		}

		fmt.Printf("%d. %s%s\n", i+1, step.Description, marker)
		fmt.Printf("   Handler: %s\n", step.Handler)
		if len(step.HandlerArgs) > 0 {
			fmt.Printf("   Arguments:\n")
			for k, v := range step.HandlerArgs {
				fmt.Printf("     %s: %s\n", k, v)
			}
		}
		fmt.Println()
	}
}

func cmdPlan(args []string) {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ensura plan [options] <file.ens>")
		os.Exit(1)
	}

	result, err := loadAndCompile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result.plan.ToJSON())
		return
	}

	fmt.Print(result.plan.String())
}

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	interval := fs.Duration("interval", 30*time.Second, "Interval between enforcement loops")
	retries := fs.Int("retries", 3, "Maximum retries per step")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ensura run [options] <file.ens>")
		os.Exit(1)
	}

	result, err := loadAndCompile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create runtime configuration
	config := &runtime.Config{
		Interval:   *interval,
		MaxRetries: *retries,
		DryRun:     false,
		CheckOnly:  false,
		Redact:     true,
		Logger:     os.Stdout,
	}

	// Create runtime with default handlers
	registry := adapters.NewDefaultRegistry()
	rt := runtime.New(result.plan, registry, config)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping...")
		cancel()
	}()

	fmt.Printf("Starting enforcement loop (interval: %s, retries: %d)\n", *interval, *retries)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	if err := rt.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cmdCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ensura check [options] <file.ens>")
		os.Exit(1)
	}

	result, err := loadAndCompile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create runtime configuration for check-only
	config := &runtime.Config{
		DryRun:    true,
		CheckOnly: true,
		Redact:    true,
		Logger:    os.Stdout,
	}

	// Create runtime with default handlers
	registry := adapters.NewDefaultRegistry()
	rt := runtime.New(result.plan, registry, config)

	ctx := context.Background()
	runResult := rt.Check(ctx)

	if *jsonOutput {
		output := map[string]interface{}{
			"allSatisfied":  runResult.AllSatisfied,
			"totalChecks":   runResult.TotalChecks,
			"totalFailures": runResult.TotalFailures,
			"duration":      runResult.EndTime.Sub(runResult.StartTime).String(),
			"steps":         make([]map[string]interface{}, len(runResult.Steps)),
		}
		for i, step := range runResult.Steps {
			stepOutput := map[string]interface{}{
				"description": step.Step.Description,
				"status":      step.Status.String(),
			}
			if step.Message != "" {
				stepOutput["message"] = step.Message
			}
			if step.Error != nil {
				stepOutput["error"] = step.Error.Error()
			}
			output["steps"].([]map[string]interface{})[i] = stepOutput
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
	}

	if !runResult.AllSatisfied {
		os.Exit(1)
	}
}
