package binder

import (
	"testing"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/parser"
)

func TestBindImplicitSubject(t *testing.T) {
	input := `ensure exists on file "secrets.db"
ensure encrypted with AES:256 key "env:KEY"`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	b := New()
	program = b.Bind(program)

	if len(b.Errors()) > 0 {
		t.Fatalf("Binding errors: %v", b.Errors())
	}

	// Second ensure should inherit subject from first
	ensure2, ok := program.Statements[1].(*ast.EnsureStmt)
	if !ok {
		t.Fatal("Expected EnsureStmt")
	}

	if ensure2.Subject == nil {
		t.Error("Expected subject to be bound")
	}
	if ensure2.Subject.Path != "secrets.db" {
		t.Errorf("Expected path 'secrets.db', got %q", ensure2.Subject.Path)
	}
}

func TestBindOnBlock(t *testing.T) {
	input := `on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:KEY"
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	b := New()
	program = b.Bind(program)

	if len(b.Errors()) > 0 {
		t.Fatalf("Binding errors: %v", b.Errors())
	}

	block, ok := program.Statements[0].(*ast.OnBlock)
	if !ok {
		t.Fatal("Expected OnBlock")
	}

	// Both ensures should have the block's subject
	for i, stmt := range block.Statements {
		ensure, ok := stmt.(*ast.EnsureStmt)
		if !ok {
			continue
		}
		if ensure.Subject == nil {
			t.Errorf("Statement %d: expected subject", i)
		}
		if ensure.Subject.Path != "secrets.db" {
			t.Errorf("Statement %d: expected path 'secrets.db', got %q", i, ensure.Subject.Path)
		}
	}
}

func TestResourceTable(t *testing.T) {
	rt := NewResourceTable()

	decl := &ast.ResourceDecl{
		ResourceType: "file",
		Path:         "secrets.db",
		Alias:        "secrets",
	}

	if err := rt.Add(decl); err != nil {
		t.Fatalf("Failed to add resource: %v", err)
	}

	// Lookup by path
	ref := &ast.ResourceRef{
		ResourceType: "file",
		Path:         "secrets.db",
	}
	found, ok := rt.Lookup(ref)
	if !ok {
		t.Error("Failed to find resource by path")
	}
	if found.Path != "secrets.db" {
		t.Error("Wrong resource returned")
	}

	// Lookup by alias
	ref = &ast.ResourceRef{Alias: "secrets"}
	found, ok = rt.Lookup(ref)
	if !ok {
		t.Error("Failed to find resource by alias")
	}
	if found.Path != "secrets.db" {
		t.Error("Wrong resource returned")
	}
}

func TestPolicyTable(t *testing.T) {
	pt := NewPolicyTable()

	decl := &ast.PolicyDecl{
		Name:   "secure_file",
		Params: []ast.PolicyParam{{Name: "key"}},
	}

	if err := pt.Add(decl); err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}

	found, ok := pt.Lookup("secure_file")
	if !ok {
		t.Error("Failed to find policy")
	}
	if found.Name != "secure_file" {
		t.Error("Wrong policy returned")
	}

	// Duplicate should fail
	if err := pt.Add(decl); err == nil {
		t.Error("Expected error for duplicate policy")
	}
}

func TestExpandPolicies(t *testing.T) {
	input := `policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  ensure exists
  apply secure_file("env:KEY")
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	b := New()
	program = b.Bind(program)
	program = b.ExpandPolicies(program)

	if len(b.Errors()) > 0 {
		t.Fatalf("Binding errors: %v", b.Errors())
	}

	// The on block should now have 3 ensures (exists + 2 from policy)
	block, ok := program.Statements[1].(*ast.OnBlock)
	if !ok {
		t.Fatal("Expected OnBlock")
	}

	if len(block.Statements) != 3 {
		t.Errorf("Expected 3 statements after expansion, got %d", len(block.Statements))
	}
}

func TestMissingImplicitSubject(t *testing.T) {
	input := `ensure encrypted with AES:256 key "env:KEY"`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	b := New()
	b.Bind(program)

	if len(b.Errors()) == 0 {
		t.Error("Expected error for missing implicit subject")
	}
}

func TestUndefinedPolicy(t *testing.T) {
	input := `on file "secrets.db" {
  apply nonexistent_policy("arg")
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	b := New()
	b.Bind(program)

	if len(b.Errors()) == 0 {
		t.Error("Expected error for undefined policy")
	}
}

func TestWrongPolicyArgCount(t *testing.T) {
	input := `policy secure_file(key_ref, extra) {
  ensure encrypted with AES:256 key key_ref
}

on file "secrets.db" {
  apply secure_file("env:KEY")
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	b := New()
	b.Bind(program)

	if len(b.Errors()) == 0 {
		t.Error("Expected error for wrong argument count")
	}
}
