// Package fs provides the filesystem handler for EnsuraScript.
package fs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/runtime"
)

// Handler implements filesystem operations.
type Handler struct{}

// New creates a new filesystem handler.
func New() *Handler {
	return &Handler{}
}

// Name returns the handler name.
func (h *Handler) Name() string {
	return "fs.native"
}

// Check verifies a filesystem condition.
func (h *Handler) Check(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	path := subject.Path

	switch condition {
	case "exists":
		return h.checkExists(path)
	case "readable":
		return h.checkReadable(path)
	case "writable":
		return h.checkWritable(path)
	case "checksum":
		return h.checkChecksum(path, args["expected"])
	case "content":
		return h.checkContent(path, args["expected"])
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("unknown condition: %s", condition),
		}
	}
}

// Enforce ensures a filesystem condition is met.
func (h *Handler) Enforce(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	path := subject.Path

	switch condition {
	case "exists":
		return h.enforceExists(path, subject.ResourceType)
	case "content":
		return h.enforceContent(path, args["content"])
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("cannot enforce condition: %s", condition),
		}
	}
}

func (h *Handler) checkExists(path string) runtime.HandlerResult {
	_, err := os.Stat(path)
	if err == nil {
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("%s exists", path),
		}
	}
	if os.IsNotExist(err) {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s does not exist", path),
		}
	}
	return runtime.HandlerResult{
		Success: false,
		Error:   err,
	}
}

func (h *Handler) checkReadable(path string) runtime.HandlerResult {
	f, err := os.Open(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s is not readable", path),
			Error:   err,
		}
	}
	f.Close()
	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("%s is readable", path),
	}
}

func (h *Handler) checkWritable(path string) runtime.HandlerResult {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s is not writable", path),
			Error:   err,
		}
	}
	f.Close()
	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("%s is writable", path),
	}
}

func (h *Handler) checkChecksum(path, expected string) runtime.HandlerResult {
	if expected == "" {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("expected checksum not specified"),
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual == expected {
		return runtime.HandlerResult{
			Success: true,
			Message: "checksum matches",
		}
	}

	return runtime.HandlerResult{
		Success: false,
		Message: fmt.Sprintf("checksum mismatch: expected %s, got %s", expected, actual),
	}
}

func (h *Handler) checkContent(path, expected string) runtime.HandlerResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	if string(data) == expected {
		return runtime.HandlerResult{
			Success: true,
			Message: "content matches",
		}
	}

	return runtime.HandlerResult{
		Success: false,
		Message: "content does not match expected",
	}
}

func (h *Handler) enforceExists(path, resourceType string) runtime.HandlerResult {
	if resourceType == "directory" {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return runtime.HandlerResult{
				Success: false,
				Error:   err,
			}
		}
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("created directory %s", path),
		}
	}

	// For files, ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	// Create empty file if it doesn't exist
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return runtime.HandlerResult{
				Success: true,
				Message: fmt.Sprintf("%s already exists", path),
			}
		}
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}
	f.Close()

	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("created file %s", path),
	}
}

func (h *Handler) enforceContent(path, content string) runtime.HandlerResult {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}
	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("wrote content to %s", path),
	}
}
