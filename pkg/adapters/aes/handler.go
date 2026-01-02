// Package aes provides AES encryption handling for EnsuraScript.
package aes

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/runtime"
)

// MagicHeader identifies encrypted files.
var MagicHeader = []byte("ENSURA_AES256_V1")

// Handler implements AES-256 encryption operations.
type Handler struct{}

// New creates a new AES handler.
func New() *Handler {
	return &Handler{}
}

// Name returns the handler name.
func (h *Handler) Name() string {
	return "AES:256"
}

// Check verifies encryption status.
func (h *Handler) Check(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	path := subject.Path

	switch condition {
	case "encrypted":
		return h.checkEncrypted(path)
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("unknown condition: %s", condition),
		}
	}
}

// Enforce ensures encryption is applied.
func (h *Handler) Enforce(ctx context.Context, subject *ast.ResourceRef, condition string, args map[string]string) runtime.HandlerResult {
	if subject == nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("no subject specified"),
		}
	}

	path := subject.Path

	switch condition {
	case "encrypted":
		return h.enforceEncrypted(path, args["key"])
	default:
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("cannot enforce condition: %s", condition),
		}
	}
}

func (h *Handler) checkEncrypted(path string) runtime.HandlerResult {
	f, err := os.Open(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}
	defer f.Close()

	// Check for magic header
	header := make([]byte, len(MagicHeader))
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	if n == len(MagicHeader) && bytes.Equal(header, MagicHeader) {
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("%s is encrypted", path),
		}
	}

	return runtime.HandlerResult{
		Success: false,
		Message: fmt.Sprintf("%s is not encrypted", path),
	}
}

func (h *Handler) enforceEncrypted(path, keyRef string) runtime.HandlerResult {
	// Get the encryption key
	key, err := resolveKey(keyRef)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("failed to resolve key: %w", err),
		}
	}

	// Read the current file content
	data, err := os.ReadFile(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	// Check if already encrypted
	if len(data) >= len(MagicHeader) && bytes.Equal(data[:len(MagicHeader)], MagicHeader) {
		return runtime.HandlerResult{
			Success: true,
			Message: fmt.Sprintf("%s is already encrypted", path),
		}
	}

	// Encrypt the data
	encrypted, err := encrypt(data, key)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   fmt.Errorf("encryption failed: %w", err),
		}
	}

	// Write back with magic header
	output := append(MagicHeader, encrypted...)

	// Get original file permissions
	info, err := os.Stat(path)
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	err = os.WriteFile(path, output, info.Mode())
	if err != nil {
		return runtime.HandlerResult{
			Success: false,
			Error:   err,
		}
	}

	return runtime.HandlerResult{
		Success: true,
		Message: fmt.Sprintf("encrypted %s", path),
	}
}

func resolveKey(keyRef string) ([]byte, error) {
	if keyRef == "" {
		return nil, fmt.Errorf("key reference is empty")
	}

	// Handle env: prefix
	if strings.HasPrefix(keyRef, "env:") {
		envVar := strings.TrimPrefix(keyRef, "env:")
		value := os.Getenv(envVar)
		if value == "" {
			return nil, fmt.Errorf("environment variable %s is not set", envVar)
		}
		// Hash the key to ensure it's 32 bytes for AES-256
		hash := sha256.Sum256([]byte(value))
		return hash[:], nil
	}

	// Handle file: prefix
	if strings.HasPrefix(keyRef, "file:") {
		filePath := strings.TrimPrefix(keyRef, "file:")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}
		// Hash the key to ensure it's 32 bytes for AES-256
		hash := sha256.Sum256(data)
		return hash[:], nil
	}

	// Use the key directly (hash it to ensure correct length)
	hash := sha256.Sum256([]byte(keyRef))
	return hash[:], nil
}

func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data (for use by other tools).
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
