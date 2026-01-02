// Package posix provides POSIX permissions handling for EnsuraScript.
package posix

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/runtime"
)

// Handler implements POSIX permission operations.
type Handler struct{}

// New creates a new POSIX handler.
func New() *Handler {
	return &Handler{}
}

// Name returns the handler name.
func (h *Handler) Name() string {
	return "posix"
}

// Check verifies POSIX permissions.
func (h *Handler) Check(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	path := subject.Path

	switch condition {
	case "permissions":
		return h.checkPermissions(path, args["mode"])
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("unknown condition: %s", condition),
		}
	}
}

// Enforce ensures POSIX permissions are set.
func (h *Handler) Enforce(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	path := subject.Path

	switch condition {
	case "permissions":
		return h.enforcePermissions(path, args["mode"])
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("cannot enforce condition: %s", condition),
		}
	}
}

func (h *Handler) checkPermissions(path, mode string) runtime.HandlerResult {
	if mode == "" {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("mode not specified"),
		}
	}

	expectedMode, err := parseMode(mode)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	actualMode := info.Mode().Perm()
	if actualMode == expectedMode {
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("%s has permissions %04o", path, actualMode),
		}
	}

	return runtime.HandlerResult{
		Success: false,
		Message: fmt.Sprintf("%s has permissions %04o, expected %04o", path, actualMode, expectedMode),
	}
}

func (h *Handler) enforcePermissions(path, mode string) runtime.HandlerResult {
	if mode == "" {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("mode not specified"),
		}
	}

	expectedMode, err := parseMode(mode)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	err = os.Chmod(path, expectedMode)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("set permissions on %s to %04o", path, expectedMode),
	}
}

func parseMode(mode string) (os.FileMode, error) {
	// Parse octal mode like "0600" or "600"
	modeInt, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid mode %q: %w", mode, err)
	}
	return os.FileMode(modeInt), nil
}
