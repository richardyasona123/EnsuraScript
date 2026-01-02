package posix

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ensurascript/ensura/pkg/ast"
)

func TestCheckPermissions(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "perms.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}

	// Check current permissions
	result := h.Check(ctx, subject, "permissions", map[string]string{"mode": "0644"})
	if !result.Success {
		t.Errorf("Expected permissions check to succeed: %s", result.Message)
	}

	// Check wrong permissions
	result = h.Check(ctx, subject, "permissions", map[string]string{"mode": "0600"})
	if result.Success {
		t.Error("Expected permissions check to fail for wrong mode")
	}
}

func TestEnforcePermissions(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "perms.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}

	// Enforce new permissions
	result := h.Enforce(ctx, subject, "permissions", map[string]string{"mode": "0600"})
	if !result.Success {
		t.Errorf("Expected enforce to succeed: %v", result.Error)
	}

	// Verify permissions
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %04o", info.Mode().Perm())
	}
}

func TestInvalidMode(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "perms.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}

	// Invalid mode
	result := h.Check(ctx, subject, "permissions", map[string]string{"mode": "invalid"})
	if result.Success {
		t.Error("Expected failure for invalid mode")
	}
}

func TestMissingMode(t *testing.T) {
	h := New()
	ctx := context.Background()

	subject := &ast.ResourceRef{Path: "/tmp/test.txt", ResourceType: "file"}

	result := h.Check(ctx, subject, "permissions", nil)
	if result.Success {
		t.Error("Expected failure for missing mode")
	}
}
