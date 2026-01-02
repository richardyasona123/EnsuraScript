package cron

import (
	"context"
	"testing"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/lexer"
)

func TestHandler_Name(t *testing.T) {
	h := New()
	if got := h.Name(); got != "cron.native" {
		t.Errorf("Name() = %v, want cron.native", got)
	}
}

func TestHandler_Check_NoSubject(t *testing.T) {
	h := New()
	ctx := context.Background()

	result := h.Check(ctx, nil, "scheduled", map[string]string{})

	if result.Success {
		t.Error("Check() should fail with no subject")
	}
	if result.Error == nil {
		t.Error("Check() should return an error with no subject")
	}
}

func TestHandler_Check_NoSchedule(t *testing.T) {
	h := New()
	ctx := context.Background()
	subject := &ast.ResourceRef{
		Position:     lexer.Position{},
		ResourceType: "cron",
		Path:         "test_job",
	}

	result := h.Check(ctx, subject, "scheduled", map[string]string{})

	if result.Success {
		t.Error("Check() should fail with no schedule argument")
	}
	if result.Error == nil {
		t.Error("Check() should return an error with no schedule argument")
	}
}

func TestHandler_Check_UnknownCondition(t *testing.T) {
	h := New()
	ctx := context.Background()
	subject := &ast.ResourceRef{
		Position:     lexer.Position{},
		ResourceType: "cron",
		Path:         "test_job",
	}

	result := h.Check(ctx, subject, "unknown", map[string]string{"schedule": "0 2 * * *"})

	if result.Success {
		t.Error("Check() should fail with unknown condition")
	}
	if result.Error == nil {
		t.Error("Check() should return an error with unknown condition")
	}
}

func TestHandler_Enforce_NoSubject(t *testing.T) {
	h := New()
	ctx := context.Background()

	result := h.Enforce(ctx, nil, "scheduled", map[string]string{})

	if result.Success {
		t.Error("Enforce() should fail with no subject")
	}
	if result.Error == nil {
		t.Error("Enforce() should return an error with no subject")
	}
}

func TestHandler_Enforce_NoSchedule(t *testing.T) {
	h := New()
	ctx := context.Background()
	subject := &ast.ResourceRef{
		Position:     lexer.Position{},
		ResourceType: "cron",
		Path:         "test_job",
	}

	result := h.Enforce(ctx, subject, "scheduled", map[string]string{})

	if result.Success {
		t.Error("Enforce() should fail with no schedule argument")
	}
	if result.Error == nil {
		t.Error("Enforce() should return an error with no schedule argument")
	}
}

func TestHandler_Enforce_NoCommand(t *testing.T) {
	h := New()
	ctx := context.Background()
	subject := &ast.ResourceRef{
		Position:     lexer.Position{},
		ResourceType: "cron",
		Path:         "test_job",
	}

	result := h.Enforce(ctx, subject, "scheduled", map[string]string{
		"schedule": "0 2 * * *",
	})

	if result.Success {
		t.Error("Enforce() should fail with no command argument")
	}
	if result.Error == nil {
		t.Error("Enforce() should return an error with no command argument")
	}
}

func TestHandler_Enforce_UnknownCondition(t *testing.T) {
	h := New()
	ctx := context.Background()
	subject := &ast.ResourceRef{
		Position:     lexer.Position{},
		ResourceType: "cron",
		Path:         "test_job",
	}

	result := h.Enforce(ctx, subject, "unknown", map[string]string{
		"schedule": "0 2 * * *",
		"command":  "echo test",
	})

	if result.Success {
		t.Error("Enforce() should fail with unknown condition")
	}
	if result.Error == nil {
		t.Error("Enforce() should return an error with unknown condition")
	}
}

// Note: Testing actual cron job creation/checking is platform-specific
// and would require mocking or integration tests. These tests verify
// the basic validation logic and error handling.
