package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/go-sdk/seal"
)

func TestSecretUnsealAll(t *testing.T) {
	dir := t.TempDir()

	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	privPath := filepath.Join(dir, "key")
	if err := os.WriteFile(privPath, priv, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	sealed, err := seal.Seal(pub, map[string][]byte{
		"token": []byte("s3cr3t"),
		"url":   []byte("https://example.com"),
	}, "test")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealedPath := filepath.Join(dir, "com.openfaas.secrets")
	if err := os.WriteFile(sealedPath, sealed, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	unsealInput = sealedPath
	unsealKey = ""

	if err := runSecretUnseal(nil, []string{privPath}); err != nil {
		t.Fatalf("runSecretUnseal: %v", err)
	}
}

func TestSecretUnsealSingleKey(t *testing.T) {
	dir := t.TempDir()

	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	privPath := filepath.Join(dir, "key")
	if err := os.WriteFile(privPath, priv, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	sealed, err := seal.Seal(pub, map[string][]byte{
		"token": []byte("s3cr3t"),
	}, "")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealedPath := filepath.Join(dir, "com.openfaas.secrets")
	if err := os.WriteFile(sealedPath, sealed, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	unsealInput = sealedPath
	unsealKey = "token"

	if err := runSecretUnseal(nil, []string{privPath}); err != nil {
		t.Fatalf("runSecretUnseal: %v", err)
	}
}

func TestSecretUnsealMissingKey(t *testing.T) {
	dir := t.TempDir()

	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	privPath := filepath.Join(dir, "key")
	if err := os.WriteFile(privPath, priv, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	sealed, err := seal.Seal(pub, map[string][]byte{"a": []byte("b")}, "")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealedPath := filepath.Join(dir, "com.openfaas.secrets")
	if err := os.WriteFile(sealedPath, sealed, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	unsealInput = sealedPath
	unsealKey = "nonexistent"

	if err := runSecretUnseal(nil, []string{privPath}); err == nil {
		t.Fatal("expected error for missing key")
	}
}
