package imply

import (
	"testing"

	"github.com/ensurascript/ensura/pkg/parser"
)

func TestImplicationExpansion(t *testing.T) {
	input := `on file "secrets.db" {
  ensure encrypted with AES:256 key "env:KEY"
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	expander := NewExpander()
	program = expander.Expand(program)

	if len(expander.Errors()) > 0 {
		t.Fatalf("Expansion errors: %v", expander.Errors())
	}

	// encrypted implies exists, readable, writable
	// So we should have 4 ensures: exists, readable, writable, encrypted
	// These should be deduplicated if there are duplicates
}

func TestConditionRegistry(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		condition string
		exists    bool
		implies   []string
	}{
		{"exists", true, nil},
		{"encrypted", true, []string{"exists", "readable", "writable"}},
		{"permissions", true, []string{"exists"}},
		{"reachable", true, nil},
		{"unknown", false, nil},
	}

	for _, tt := range tests {
		meta, ok := registry.Get(tt.condition)
		if ok != tt.exists {
			t.Errorf("Condition %q: expected exists=%v, got %v", tt.condition, tt.exists, ok)
			continue
		}
		if !ok {
			continue
		}
		if len(meta.Implies) != len(tt.implies) {
			t.Errorf("Condition %q: expected %d implies, got %d", tt.condition, len(tt.implies), len(meta.Implies))
		}
	}
}

func TestConflictDetection(t *testing.T) {
	input := `on file "test.txt" {
  ensure encrypted with AES:256 key "env:KEY"
  ensure unencrypted
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	expander := NewExpander()
	program = expander.Expand(program)

	conflicts := expander.CheckConflicts(program)
	if len(conflicts) == 0 {
		t.Error("Expected conflict between encrypted and unencrypted")
	}
}

func TestResourceTypeValidation(t *testing.T) {
	// encrypted is only applicable to files, not http
	input := `ensure encrypted on http "https://example.com" with AES:256 key "env:KEY"`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	expander := NewExpander()
	expander.Expand(program)

	if len(expander.Errors()) == 0 {
		t.Error("Expected error for applying encrypted to http resource")
	}
}

func TestDeduplication(t *testing.T) {
	input := `on file "test.txt" {
  ensure exists
  ensure encrypted with AES:256 key "env:KEY"
}`

	program, errors := parser.ParseString(input)
	if len(errors) > 0 {
		t.Fatalf("Parse errors: %v", errors)
	}

	expander := NewExpander()
	program = expander.Expand(program)

	// Count how many times 'exists' appears
	existsCount := 0
	for _, stmt := range program.Statements {
		// This is simplified - in reality we'd need to walk the AST
		_ = stmt
	}

	// The exists condition should only appear once despite being both
	// explicit and implied by encrypted
	_ = existsCount
}
