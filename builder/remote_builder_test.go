package builder

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexellis/hmac/v2"
	"github.com/openfaas/go-sdk/seal"
)

func TestRunRemoteBuildWithSecrets(t *testing.T) {
	pub, priv, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("seal.GenerateKeyPair: %v", err)
	}

	tarPath := filepath.Join(t.TempDir(), "req.tar")
	if err := os.WriteFile(tarPath, createTestTar(t), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	payloadSecretPath := filepath.Join(t.TempDir(), "payload-secret")
	if err := os.WriteFile(payloadSecretPath, []byte("payload-secret"), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/publickey":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(remoteBuilderPublicKeyResponse{
				KeyID:     "builder-key-1",
				Algorithm: naclBoxAlgorithm,
				PublicKey: string(pub),
			})
		case "/build":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("io.ReadAll returned error: %v", err)
			}

			// Verify HMAC over entire tar
			wantDigest := hmac.Sign(body, []byte("payload-secret"), sha256.New)
			gotDigest := r.Header.Get("X-Build-Signature")
			if gotDigest != "sha256="+hex.EncodeToString(wantDigest) {
				t.Fatalf("unexpected signature: %s", gotDigest)
			}

			// Body should be a tar, not multipart
			if ct := r.Header.Get("Content-Type"); ct != "application/octet-stream" {
				t.Fatalf("unexpected content-type: %s", ct)
			}

			// Extract sealed secrets from tar
			tr := tar.NewReader(bytes.NewReader(body))
			var sealedData []byte
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("tar.Next returned error: %v", err)
				}
				if hdr.Name == "com.openfaas.secrets" {
					sealedData, err = io.ReadAll(tr)
					if err != nil {
						t.Fatalf("io.ReadAll sealed secrets: %v", err)
					}
				}
			}

			if sealedData == nil {
				t.Fatal("sealed secrets file not found in tar")
			}

			// Unseal and verify
			secrets, err := seal.Unseal(priv, sealedData)
			if err != nil {
				t.Fatalf("seal.Unseal returned error: %v", err)
			}

			if got := string(secrets["pip_token"]); got != "s3cr3t" {
				t.Fatalf("want pip_token to be %q, got %q", "s3cr3t", got)
			}

			w.Header().Set("Content-Type", "application/x-ndjson")
			io.WriteString(w, `{"status":"in_progress","log":["step 1"]}`+"\n")
			io.WriteString(w, `{"status":"success","image":"ttl.sh/test:latest"}`+"\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	builderURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}

	if err := runRemoteBuild(builderURL, tarPath, payloadSecretPath, "", "", map[string]string{
		"pip_token": "s3cr3t",
	}, true, "fn", "ttl.sh/test:latest"); err != nil {
		t.Fatalf("runRemoteBuild returned error: %v", err)
	}
}

func TestRunRemoteBuildWithPinnedPublicKey(t *testing.T) {
	pub, _, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("seal.GenerateKeyPair: %v", err)
	}

	tarPath := filepath.Join(t.TempDir(), "req.tar")
	if err := os.WriteFile(tarPath, createTestTar(t), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	payloadSecretPath := filepath.Join(t.TempDir(), "payload-secret")
	if err := os.WriteFile(payloadSecretPath, []byte("payload-secret"), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	publicKeyPath := filepath.Join(t.TempDir(), "public-key.json")
	publicKeyJSON, err := json.Marshal(remoteBuilderPublicKeyResponse{
		KeyID:     "builder-key-1",
		Algorithm: naclBoxAlgorithm,
		PublicKey: string(pub),
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, publicKeyJSON, 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/publickey" {
			t.Fatal("did not expect /publickey to be called when a pinned key file is provided")
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		io.WriteString(w, `{"status":"success","image":"ttl.sh/test:latest"}`+"\n")
	}))
	defer server.Close()

	builderURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}

	if err := runRemoteBuild(builderURL, tarPath, payloadSecretPath, publicKeyPath, "", map[string]string{
		"pip_token": "s3cr3t",
	}, true, "fn", "ttl.sh/test:latest"); err != nil {
		t.Fatalf("runRemoteBuild returned error: %v", err)
	}
}

func TestRunRemoteBuildWithLiteralPublicKey(t *testing.T) {
	pub, _, err := seal.GenerateKeyPair()
	if err != nil {
		t.Fatalf("seal.GenerateKeyPair: %v", err)
	}

	tarPath := filepath.Join(t.TempDir(), "req.tar")
	if err := os.WriteFile(tarPath, createTestTar(t), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	payloadSecretPath := filepath.Join(t.TempDir(), "payload-secret")
	if err := os.WriteFile(payloadSecretPath, []byte("payload-secret"), 0o600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/publickey" {
			t.Fatal("did not expect /publickey to be called when a literal key is provided")
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		io.WriteString(w, `{"status":"success","image":"ttl.sh/test:latest"}`+"\n")
	}))
	defer server.Close()

	builderURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}

	if err := runRemoteBuild(builderURL, tarPath, payloadSecretPath, string(pub), "builder-key-1", map[string]string{
		"pip_token": "s3cr3t",
	}, true, "fn", "ttl.sh/test:latest"); err != nil {
		t.Fatalf("runRemoteBuild returned error: %v", err)
	}
}

func createTestTar(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	data := []byte(`{"image":"test:latest"}`)
	if err := tw.WriteHeader(&tar.Header{
		Name: "com.openfaas.docker.config",
		Mode: 0600,
		Size: int64(len(data)),
	}); err != nil {
		t.Fatalf("tar.WriteHeader: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("tar.Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar.Close: %v", err)
	}
	return buf.Bytes()
}
