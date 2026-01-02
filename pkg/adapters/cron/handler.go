// Package cron provides the cron scheduling handler for EnsuraScript.
package cron

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ensurascript/ensura/pkg/ast"
	pkgruntime "github.com/ensurascript/ensura/pkg/runtime"
)

// Handler implements cron scheduling operations.
type Handler struct{}

// New creates a new cron handler.
func New() *Handler {
	return &Handler{}
}

// Name returns the handler name.
func (h *Handler) Name() string {
	return "cron.native"
}

// Check verifies a cron scheduling condition.
func (h *Handler) Check(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) pkgruntime.HandlerResult {
	if subject == nil {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	if condition != "scheduled" {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("unknown condition: %s", condition),
		}
	}

	schedule := args["schedule"]
	if schedule == "" {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("schedule argument required"),
		}
	}

	jobName := subject.Path

	// Check if cron job exists based on platform
	exists, err := h.cronJobExists(jobName)
	if err != nil {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("failed to check cron job: %w", err),
		}
	}

	if exists {
		return pkgruntime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("cron job %s is scheduled", jobName),
		}
	}

	return pkgruntime.HandlerResult{
		Success: false,
		Message: fmt.Sprintf("cron job %s is not scheduled", jobName),
	}
}

// Enforce ensures a cron scheduling condition is met.
func (h *Handler) Enforce(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) pkgruntime.HandlerResult {
	if subject == nil {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	if condition != "scheduled" {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("cannot enforce condition: %s", condition),
		}
	}

	schedule := args["schedule"]
	if schedule == "" {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("schedule argument required"),
		}
	}

	jobName := subject.Path
	command := args["command"]
	if command == "" {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("command argument required for enforcement"),
		}
	}

	// Add/update cron job based on platform
	if err := h.addCronJob(jobName, schedule, command); err != nil {
		return pkgruntime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("failed to add cron job: %w", err),
		}
	}

	return pkgruntime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("scheduled cron job %s: %s", jobName, schedule),
	}
}

// cronJobExists checks if a cron job with the given identifier exists.
func (h *Handler) cronJobExists(jobName string) (bool, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		// Use crontab -l to list current user's cron jobs
		cmd := exec.Command("crontab", "-l")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Exit status 1 typically means no crontab
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return false, nil
			}
			return false, err
		}

		// Look for a comment marker that identifies this job
		marker := fmt.Sprintf("# EnsuraScript: %s", jobName)
		return strings.Contains(string(output), marker), nil

	default:
		return false, fmt.Errorf("cron scheduling not supported on %s", runtime.GOOS)
	}
}

// addCronJob adds or updates a cron job entry.
func (h *Handler) addCronJob(jobName, schedule, command string) error {
	switch runtime.GOOS {
	case "darwin", "linux":
		// Get existing crontab
		cmd := exec.Command("crontab", "-l")
		output, err := cmd.CombinedOutput()
		var existingCrontab string
		if err != nil {
			// Exit status 1 typically means no crontab exists yet
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				existingCrontab = ""
			} else {
				return err
			}
		} else {
			existingCrontab = string(output)
		}

		// Remove existing entry with this job name
		marker := fmt.Sprintf("# EnsuraScript: %s", jobName)
		lines := strings.Split(existingCrontab, "\n")
		var newLines []string
		skipNext := false
		for _, line := range lines {
			if strings.Contains(line, marker) {
				skipNext = true
				continue
			}
			if skipNext {
				skipNext = false
				continue
			}
			if line != "" {
				newLines = append(newLines, line)
			}
		}

		// Add new entry
		newEntry := fmt.Sprintf("%s\n%s %s", marker, schedule, command)
		newLines = append(newLines, newEntry)

		// Write new crontab
		newCrontab := strings.Join(newLines, "\n") + "\n"
		tmpFile, err := os.CreateTemp("", "ensura-crontab-*")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(newCrontab); err != nil {
			tmpFile.Close()
			return err
		}
		tmpFile.Close()

		// Install new crontab
		installCmd := exec.Command("crontab", tmpFile.Name())
		if output, err := installCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to install crontab: %w, output: %s", err, string(output))
		}

		return nil

	default:
		return fmt.Errorf("cron scheduling not supported on %s", runtime.GOOS)
	}
}
