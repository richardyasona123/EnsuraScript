package parser

import (
	"testing"

	"github.com/ensurascript/ensura/pkg/ast"
)

func TestParseResourceDecl(t *testing.T) {
	input := `resource file "secrets.db"
resource http "https://example.com" as api`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// First resource
	res1, ok := program.Statements[0].(*ast.ResourceDecl)
	if !ok {
		t.Fatalf("Expected ResourceDecl, got %T", program.Statements[0])
	}
	if res1.ResourceType != "file" {
		t.Errorf("Expected type 'file', got %q", res1.ResourceType)
	}
	if res1.Path != "secrets.db" {
		t.Errorf("Expected path 'secrets.db', got %q", res1.Path)
	}

	// Second resource with alias
	res2, ok := program.Statements[1].(*ast.ResourceDecl)
	if !ok {
		t.Fatalf("Expected ResourceDecl, got %T", program.Statements[1])
	}
	if res2.Alias != "api" {
		t.Errorf("Expected alias 'api', got %q", res2.Alias)
	}
}

func TestParseEnsureStmt(t *testing.T) {
	input := `ensure exists on file "secrets.db"
ensure encrypted on file "secrets.db" with AES:256 key "env:KEY"`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// First ensure
	ensure1, ok := program.Statements[0].(*ast.EnsureStmt)
	if !ok {
		t.Fatalf("Expected EnsureStmt, got %T", program.Statements[0])
	}
	if ensure1.Condition != "exists" {
		t.Errorf("Expected condition 'exists', got %q", ensure1.Condition)
	}
	if ensure1.Subject == nil {
		t.Fatal("Expected subject, got nil")
	}
	if ensure1.Subject.Path != "secrets.db" {
		t.Errorf("Expected path 'secrets.db', got %q", ensure1.Subject.Path)
	}

	// Second ensure with handler
	ensure2, ok := program.Statements[1].(*ast.EnsureStmt)
	if !ok {
		t.Fatalf("Expected EnsureStmt, got %T", program.Statements[1])
	}
	if ensure2.Handler == nil {
		t.Fatal("Expected handler, got nil")
	}
	if ensure2.Handler.Name != "AES:256" {
		t.Errorf("Expected handler 'AES:256', got %q", ensure2.Handler.Name)
	}
	if ensure2.Handler.Args["key"] != "env:KEY" {
		t.Errorf("Expected key 'env:KEY', got %q", ensure2.Handler.Args["key"])
	}
}

func TestParseOnBlock(t *testing.T) {
	input := `on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:KEY"
}`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	block, ok := program.Statements[0].(*ast.OnBlock)
	if !ok {
		t.Fatalf("Expected OnBlock, got %T", program.Statements[0])
	}

	if block.Subject.Path != "secrets.db" {
		t.Errorf("Expected path 'secrets.db', got %q", block.Subject.Path)
	}

	if len(block.Statements) != 2 {
		t.Fatalf("Expected 2 inner statements, got %d", len(block.Statements))
	}
}

func TestParsePolicy(t *testing.T) {
	input := `policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  apply secure_file("env:KEY")
}`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// Policy declaration
	policy, ok := program.Statements[0].(*ast.PolicyDecl)
	if !ok {
		t.Fatalf("Expected PolicyDecl, got %T", program.Statements[0])
	}
	if policy.Name != "secure_file" {
		t.Errorf("Expected name 'secure_file', got %q", policy.Name)
	}
	if len(policy.Params) != 1 {
		t.Fatalf("Expected 1 param, got %d", len(policy.Params))
	}
	if policy.Params[0].Name != "key_ref" {
		t.Errorf("Expected param 'key_ref', got %q", policy.Params[0].Name)
	}

	// On block with apply
	block, ok := program.Statements[1].(*ast.OnBlock)
	if !ok {
		t.Fatalf("Expected OnBlock, got %T", program.Statements[1])
	}

	apply, ok := block.Statements[0].(*ast.ApplyStmt)
	if !ok {
		t.Fatalf("Expected ApplyStmt, got %T", block.Statements[0])
	}
	if apply.PolicyName != "secure_file" {
		t.Errorf("Expected policy 'secure_file', got %q", apply.PolicyName)
	}
	if len(apply.Args) != 1 || apply.Args[0] != "env:KEY" {
		t.Errorf("Expected args ['env:KEY'], got %v", apply.Args)
	}
}

func TestParseForEach(t *testing.T) {
	input := `for each file in directory "/secrets" {
  ensure encrypted with AES:256 key "env:KEY"
}`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	forEach, ok := program.Statements[0].(*ast.ForEachStmt)
	if !ok {
		t.Fatalf("Expected ForEachStmt, got %T", program.Statements[0])
	}

	if forEach.ItemType != "file" {
		t.Errorf("Expected item type 'file', got %q", forEach.ItemType)
	}
	if forEach.Container.Path != "/secrets" {
		t.Errorf("Expected container '/secrets', got %q", forEach.Container.Path)
	}
}

func TestParseInvariant(t *testing.T) {
	input := `invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:KEY"
  }
}`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	inv, ok := program.Statements[0].(*ast.InvariantBlock)
	if !ok {
		t.Fatalf("Expected InvariantBlock, got %T", program.Statements[0])
	}

	if len(inv.Statements) != 1 {
		t.Fatalf("Expected 1 inner statement, got %d", len(inv.Statements))
	}
}

func TestParseOnViolation(t *testing.T) {
	input := `on violation {
  retry 3
  notify "ops"
  notify "security"
}`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	violation, ok := program.Statements[0].(*ast.OnViolationBlock)
	if !ok {
		t.Fatalf("Expected OnViolationBlock, got %T", program.Statements[0])
	}

	if violation.Handler.Retry != 3 {
		t.Errorf("Expected retry 3, got %d", violation.Handler.Retry)
	}
	if len(violation.Handler.Notify) != 2 {
		t.Fatalf("Expected 2 notify targets, got %d", len(violation.Handler.Notify))
	}
}

func TestParseGuard(t *testing.T) {
	input := `ensure encrypted on file "secrets.db" when environment == "prod"`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	ensure, ok := program.Statements[0].(*ast.EnsureStmt)
	if !ok {
		t.Fatalf("Expected EnsureStmt, got %T", program.Statements[0])
	}

	if ensure.Guard == nil {
		t.Fatal("Expected guard, got nil")
	}
	if ensure.Guard.Left != "environment" {
		t.Errorf("Expected left 'environment', got %q", ensure.Guard.Left)
	}
	if ensure.Guard.Operator != "==" {
		t.Errorf("Expected operator '==', got %q", ensure.Guard.Operator)
	}
	if ensure.Guard.Right != "prod" {
		t.Errorf("Expected right 'prod', got %q", ensure.Guard.Right)
	}
}

func TestParseAssume(t *testing.T) {
	input := `assume environment == "dev"
assume filesystem reliable`

	program, errors := ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// Guard-style assume
	assume1, ok := program.Statements[0].(*ast.AssumeStmt)
	if !ok {
		t.Fatalf("Expected AssumeStmt, got %T", program.Statements[0])
	}
	if assume1.Guard == nil {
		t.Fatal("Expected guard, got nil")
	}

	// Simple assume
	assume2, ok := program.Statements[1].(*ast.AssumeStmt)
	if !ok {
		t.Fatalf("Expected AssumeStmt, got %T", program.Statements[1])
	}
	if assume2.Simple != "filesystem reliable" {
		t.Errorf("Expected 'filesystem reliable', got %q", assume2.Simple)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`ensure`, "missing condition"},
		{`on {`, "missing subject"},
		{`policy {`, "missing name"},
	}

	for _, tt := range tests {
		_, errors := ParseString(tt.input)
		if len(errors) == 0 {
			t.Errorf("%s: expected errors, got none", tt.desc)
		}
	}
}
