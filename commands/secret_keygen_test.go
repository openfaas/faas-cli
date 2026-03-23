package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretKeygen(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "testkey")

	keygenOutput = keyPath
	if err := runSecretKeygen(nil, nil); err != nil {
		t.Fatalf("runSecretKeygen: %v", err)
	}

	priv, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("reading private key: %v", err)
	}
	if len(priv) == 0 {
		t.Fatal("private key is empty")
	}

	pub, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		t.Fatalf("reading public key: %v", err)
	}
	if len(pub) == 0 {
		t.Fatal("public key is empty")
	}

	// Keys should be different
	if string(priv) == string(pub) {
		t.Fatal("private and public keys are the same")
	}

	// Check permissions on private key
	info, _ := os.Stat(keyPath)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("want private key perms 0600, got %o", info.Mode().Perm())
	}
}

func TestSecretKeygenSubdirectory(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "sub", "dir", "mykey")

	keygenOutput = keyPath
	if err := runSecretKeygen(nil, nil); err != nil {
		t.Fatalf("runSecretKeygen: %v", err)
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("private key not created: %v", err)
	}
	if _, err := os.Stat(keyPath + ".pub"); err != nil {
		t.Fatalf("public key not created: %v", err)
	}
}
