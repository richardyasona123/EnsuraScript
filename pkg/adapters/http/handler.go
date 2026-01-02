// Package http provides HTTP endpoint handling for EnsuraScript.
package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/runtime"
)

// Handler implements HTTP endpoint operations.
type Handler struct {
	client *http.Client
}

// New creates a new HTTP handler.
func New() *Handler {
	return &Handler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the handler name.
func (h *Handler) Name() string {
	return "http.get"
}

// Check verifies an HTTP endpoint condition.
func (h *Handler) Check(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	url := subject.Path

	switch condition {
	case "reachable":
		return h.checkReachable(ctx, url)
	case "status_code":
		return h.checkStatusCode(ctx, url, args["expected_status"])
	case "tls":
		return h.checkTLS(ctx, url)
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("unknown condition: %s", condition),
		}
	}
}

// Enforce is not applicable for HTTP endpoints (read-only).
func (h *Handler) Enforce(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	return runtime.HandlerResult{
		Success: false,
		Error:   fmt.Errorf("HTTP endpoints cannot be enforced, only checked"),
	}
}

func (h *Handler) checkReachable(ctx context.Context, url string) runtime.HandlerResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s is not reachable", url),
			Error:   err,
		}
	}
	defer resp.Body.Close()

	// Any successful response (2xx, 3xx) is considered reachable
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("%s is reachable (status: %d)", url, resp.StatusCode),
		}
	}

	return runtime.HandlerResult{
		Success: false,
		Message: fmt.Sprintf("%s returned status %d", url, resp.StatusCode),
	}
}

func (h *Handler) checkStatusCode(ctx context.Context, url, expectedStatus string) runtime.HandlerResult {
	if expectedStatus == "" {
		expectedStatus = "200"
	}

	expected, err := strconv.Atoi(expectedStatus)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("invalid expected status: %s", expectedStatus),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s is not reachable", url),
			Error:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == expected {
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("%s returned expected status %d", url, expected),
		}
	}

	return runtime.HandlerResult{
		Success: false,
		Message: fmt.Sprintf("%s returned status %d, expected %d", url, resp.StatusCode, expected),
	}
}

func (h *Handler) checkTLS(ctx context.Context, url string) runtime.HandlerResult {
	// Create a custom client that checks TLS
	tlsClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	resp, err := tlsClient.Do(req)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s TLS check failed", url),
			Error:   err,
		}
	}
	defer resp.Body.Close()

	if resp.TLS == nil {
		return runtime.HandlerResult{
			Success: false,
			Message: fmt.Sprintf("%s is not using TLS", url),
		}
	}

	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("%s is using TLS %s", url, tlsVersionString(resp.TLS.Version)),
	}
}

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}
