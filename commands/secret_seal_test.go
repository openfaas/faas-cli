package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/go-sdk/seal"
)

func TestSecretSealFromLiteral(t *testing.T) {
	dir := t.TempDir()

	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	pubPath := filepath.Join(dir, "test.pub")
	if err := os.WriteFile(pubPath, pub, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outPath := filepath.Join(dir, "com.openfaas.secrets")

	sealOutput = outPath
	sealFromLiteral = []string{"pip_token=s3cr3t", "npm_token=tok123"}
	sealFromFile = nil

	if err := runSecretSeal(nil, []string{pubPath}); err != nil {
		t.Fatalf("runSecretSeal: %v", err)
	}

	sealed, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	values, err := seal.Unseal(priv, sealed)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}

	if got := string(values["pip_token"]); got != "s3cr3t" {
		t.Fatalf("want pip_token %q, got %q", "s3cr3t", got)
	}
	if got := string(values["npm_token"]); got != "tok123" {
		t.Fatalf("want npm_token %q, got %q", "tok123", got)
	}
}

func TestSecretSealFromFile(t *testing.T) {
	dir := t.TempDir()

	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	pubPath := filepath.Join(dir, "test.pub")
	if err := os.WriteFile(pubPath, pub, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Write a binary file (fake cert)
	certData := []byte("-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n")
	certPath := filepath.Join(dir, "ca.crt")
	if err := os.WriteFile(certPath, certData, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	outPath := filepath.Join(dir, "com.openfaas.secrets")

	sealOutput = outPath
	sealFromLiteral = []string{"token=abc"}
	sealFromFile = []string{"ca.crt=" + certPath}

	if err := runSecretSeal(nil, []string{pubPath}); err != nil {
		t.Fatalf("runSecretSeal: %v", err)
	}

	sealed, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	values, err := seal.Unseal(priv, sealed)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}

	if got := string(values["token"]); got != "abc" {
		t.Fatalf("want token %q, got %q", "abc", got)
	}
	if got := string(values["ca.crt"]); got != string(certData) {
		t.Fatalf("want ca.crt %q, got %q", certData, got)
	}
}

func TestSecretSealPreRunValidation(t *testing.T) {
	sealFromLiteral = nil
	sealFromFile = nil

	if err := preRunSecretSeal(nil, nil); err == nil {
		t.Fatal("expected error when no secrets provided")
	}
}
