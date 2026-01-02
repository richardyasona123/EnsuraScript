package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ensurascript/ensura/pkg/ast"
)

func TestCheckExists(t *testing.T) {
	h := New()
	ctx := context.Background()

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test existing file
	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Check(ctx, subject, "exists", nil)
	if !result.Success {
		t.Error("Expected exists check to succeed for existing file")
	}

	// Test non-existing file
	subject = &ast.ResourceRef{Path: filepath.Join(tmpDir, "nonexistent.txt"), ResourceType: "file"}
	result = h.Check(ctx, subject, "exists", nil)
	if result.Success {
		t.Error("Expected exists check to fail for non-existing file")
	}
}

func TestEnforceExists(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	newFile := filepath.Join(tmpDir, "new.txt")

	// Enforce creation
	subject := &ast.ResourceRef{Path: newFile, ResourceType: "file"}
	result := h.Enforce(ctx, subject, "exists", nil)
	if !result.Success {
		t.Errorf("Expected enforce to succeed: %v", result.Error)
	}

	// Verify file exists
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("File should exist after enforce")
	}
}

func TestCheckReadable(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "readable.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Check(ctx, subject, "readable", nil)
	if !result.Success {
		t.Error("Expected readable check to succeed")
	}
}

func TestCheckWritable(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "writable.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Check(ctx, subject, "writable", nil)
	if !result.Success {
		t.Error("Expected writable check to succeed")
	}
}

func TestEnforceExistsDirectory(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "newdir")

	subject := &ast.ResourceRef{Path: newDir, ResourceType: "directory"}
	result := h.Enforce(ctx, subject, "exists", nil)
	if !result.Success {
		t.Errorf("Expected enforce to succeed: %v", result.Error)
	}

	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("Expected directory to be created")
	}
}

func TestCheckChecksum(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "checksum.txt")
	content := []byte("test content")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// SHA256 of "test content"
	expectedHash := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Check(ctx, subject, "checksum", map[string]string{"expected": expectedHash})
	if !result.Success {
		t.Errorf("Expected checksum check to succeed: %s", result.Message)
	}

	// Wrong checksum
	result = h.Check(ctx, subject, "checksum", map[string]string{"expected": "wrong"})
	if result.Success {
		t.Error("Expected checksum check to fail for wrong hash")
	}
}

func TestEnforceContent(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "content.txt")
	if err := os.WriteFile(tmpFile, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Enforce(ctx, subject, "content", map[string]string{"content": "new content"})
	if !result.Success {
		t.Errorf("Expected enforce to succeed: %v", result.Error)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new content" {
		t.Errorf("Expected 'new content', got %q", string(data))
	}
}

func TestNilSubject(t *testing.T) {
	h := New()
	ctx := context.Background()

	result := h.Check(ctx, nil, "exists", nil)
	if result.Success {
		t.Error("Expected failure for nil subject")
	}
	if result.Error == nil {
		t.Error("Expected error for nil subject")
	}
}
