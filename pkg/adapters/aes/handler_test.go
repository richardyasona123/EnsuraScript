package aes

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ensurascript/ensura/pkg/ast"
)

func TestCheckEncrypted(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create an unencrypted file
	unencryptedFile := filepath.Join(tmpDir, "plain.txt")
	if err := os.WriteFile(unencryptedFile, []byte("plain text"), 0644); err != nil {
		t.Fatal(err)
	}

	subject := &ast.ResourceRef{Path: unencryptedFile, ResourceType: "file"}
	result := h.Check(ctx, subject, "encrypted", nil)
	if result.Success {
		t.Error("Expected unencrypted file to fail check")
	}

	// Create an encrypted file (with magic header)
	encryptedFile := filepath.Join(tmpDir, "encrypted.txt")
	content := append(MagicHeader, []byte("encrypted data")...)
	if err := os.WriteFile(encryptedFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	subject = &ast.ResourceRef{Path: encryptedFile, ResourceType: "file"}
	result = h.Check(ctx, subject, "encrypted", nil)
	if !result.Success {
		t.Errorf("Expected encrypted file to pass check: %s", result.Message)
	}
}

func TestEnforceEncrypted(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "toencrypt.txt")
	plaintext := []byte("secret data")
	if err := os.WriteFile(tmpFile, plaintext, 0644); err != nil {
		t.Fatal(err)
	}

	// Set the encryption key
	os.Setenv("TEST_KEY", "my-secret-key")
	defer os.Unsetenv("TEST_KEY")

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Enforce(ctx, subject, "encrypted", map[string]string{"key": "env:TEST_KEY"})
	if !result.Success {
		t.Errorf("Expected enforce to succeed: %v", result.Error)
	}

	// Verify file has magic header
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) < len(MagicHeader) {
		t.Fatal("File too short after encryption")
	}

	if !bytes.Equal(data[:len(MagicHeader)], MagicHeader) {
		t.Error("Expected magic header after encryption")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("Hello, World!")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ciphertext, err := encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted text doesn't match original")
	}
}

func TestResolveKeyFromEnv(t *testing.T) {
	os.Setenv("TEST_SECRET", "my-secret-key")
	defer os.Unsetenv("TEST_SECRET")

	key, err := resolveKey("env:TEST_SECRET")
	if err != nil {
		t.Fatalf("Failed to resolve key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected 32-byte key, got %d", len(key))
	}
}

func TestResolveKeyFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.txt")
	if err := os.WriteFile(keyFile, []byte("file-key"), 0600); err != nil {
		t.Fatal(err)
	}

	key, err := resolveKey("file:" + keyFile)
	if err != nil {
		t.Fatalf("Failed to resolve key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected 32-byte key, got %d", len(key))
	}
}

func TestResolveKeyDirect(t *testing.T) {
	key, err := resolveKey("direct-key")
	if err != nil {
		t.Fatalf("Failed to resolve key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected 32-byte key, got %d", len(key))
	}
}

func TestMissingEnvKey(t *testing.T) {
	_, err := resolveKey("env:NONEXISTENT_VAR")
	if err == nil {
		t.Error("Expected error for missing env var")
	}
}

func TestAlreadyEncrypted(t *testing.T) {
	h := New()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "already.txt")

	// Write already encrypted content
	content := append(MagicHeader, []byte("already encrypted")...)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("TEST_KEY", "key")
	defer os.Unsetenv("TEST_KEY")

	subject := &ast.ResourceRef{Path: tmpFile, ResourceType: "file"}
	result := h.Enforce(ctx, subject, "encrypted", map[string]string{"key": "env:TEST_KEY"})

	if !result.Success {
		t.Error("Expected success for already encrypted file")
	}
}
