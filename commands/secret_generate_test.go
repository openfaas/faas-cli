package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretGenerate(t *testing.T) {
	generateLength = 32
	generateOutput = ""

	// Just verify it doesn't error — stdout output
	if err := runSecretGenerate(nil, nil); err != nil {
		t.Fatalf("runSecretGenerate: %v", err)
	}
}

func TestSecretGenerateToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "payload.txt")

	generateLength = 32
	generateOutput = outPath

	if err := runSecretGenerate(nil, nil); err != nil {
		t.Fatalf("runSecretGenerate: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// 32 bytes hex-encoded = 64 chars
	if len(data) != 64 {
		t.Fatalf("want 64 hex chars, got %d", len(data))
	}

	info, _ := os.Stat(outPath)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("want perms 0600, got %o", info.Mode().Perm())
	}
}
