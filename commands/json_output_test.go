package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/flags"
	storeV2 "github.com/openfaas/faas-cli/schema/store/v2"
	"github.com/openfaas/faas-provider/logs"
	types "github.com/openfaas/faas-provider/types"
	"github.com/spf13/cobra"
)

func TestSecretListJSONEmptyResultUsesJSONStdout(t *testing.T) {
	resetJSONCommandTestState(t)

	s := newInsecureWarningHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/system/secrets" {
			t.Fatalf("expected /system/secrets, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))

	gateway = s.URL
	jsonOutput = true

	stdout, stderr := captureStdoutStderr(t, func() {
		cmd := &cobra.Command{}
		cmd.SetOut(os.Stdout)

		if err := runSecretList(cmd, nil); err != nil {
			t.Fatalf("runSecretList returned error: %s", err)
		}
	})

	if strings.Contains(stdout, "No secrets found") {
		t.Fatalf("expected JSON stdout, got text output: %q", stdout)
	}
	if strings.Contains(stdout, NoTLSWarn) {
		t.Fatalf("expected TLS warning off stdout, got: %q", stdout)
	}
	if strings.Contains(stderr, NoTLSWarn) {
		t.Fatalf("expected TLS warning omitted for JSON output, got stderr: %q", stderr)
	}

	var secrets []types.Secret
	if err := json.Unmarshal([]byte(stdout), &secrets); err != nil {
		t.Fatalf("expected valid JSON stdout, got %q: %s", stdout, err)
	}
	if len(secrets) != 0 {
		t.Fatalf("expected empty secret list, got %d entries", len(secrets))
	}
}

func TestSecretListTextEmptyResultKeepsWarningVisible(t *testing.T) {
	resetJSONCommandTestState(t)

	s := newInsecureWarningHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))

	gateway = s.URL

	stdout, stderr := captureStdoutStderr(t, func() {
		if err := runSecretList(&cobra.Command{}, nil); err != nil {
			t.Fatalf("runSecretList returned error: %s", err)
		}
	})

	if !strings.Contains(stdout, "No secrets found.") {
		t.Fatalf("expected empty text message on stdout, got: %q", stdout)
	}
	if !strings.Contains(stdout, NoTLSWarn) {
		t.Fatalf("expected TLS warning on stdout, got: %q", stdout)
	}
	if strings.Contains(stderr, NoTLSWarn) {
		t.Fatalf("expected no TLS warning on stderr, got: %q", stderr)
	}
}

func TestStoreListJSONEmptyFilterUsesJSONStdout(t *testing.T) {
	resetJSONCommandTestState(t)

	s := newInsecureWarningHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"version": "1.0",
			"functions": [
				{
					"title": "NodeInfo",
					"name": "nodeinfo",
					"images": {
						"x86_64": "functions/nodeinfo:latest"
					}
				}
			]
		}`))
	}))

	storeAddress = s.URL
	platformValue = "missing"
	jsonOutput = true

	stdout, _ := captureStdoutStderr(t, func() {
		cmd := &cobra.Command{}
		cmd.SetOut(os.Stdout)

		if err := runStoreList(cmd, nil); err != nil {
			t.Fatalf("runStoreList returned error: %s", err)
		}
	})

	if strings.Contains(stdout, "No functions found") {
		t.Fatalf("expected JSON stdout, got text output: %q", stdout)
	}

	var functions []storeV2.StoreFunction
	if err := json.Unmarshal([]byte(stdout), &functions); err != nil {
		t.Fatalf("expected valid JSON stdout, got %q: %s", stdout, err)
	}
	if len(functions) != 0 {
		t.Fatalf("expected empty store list, got %d entries", len(functions))
	}
}

func TestLogsJSONOmitsTLSWarning(t *testing.T) {
	resetJSONCommandTestState(t)

	s := newInsecureWarningHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/system/logs" {
			t.Fatalf("expected /system/logs, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(logs.Message{Name: "fn", Text: "hello"})
	}))

	gateway = s.URL
	jsonOutput = true
	logFlagValues.timeFormat = flags.TimeFormat("")
	logFlagValues.tail = false
	logFlagValues.lines = -1

	stdout, stderr := captureStdoutStderr(t, func() {
		cmd := newLogsTestCommand()

		if err := runLogs(cmd, []string{"fn"}); err != nil {
			t.Fatalf("runLogs returned error: %s", err)
		}
	})

	if strings.Contains(stdout, NoTLSWarn) {
		t.Fatalf("expected TLS warning off stdout, got: %q", stdout)
	}
	if strings.Contains(stderr, NoTLSWarn) {
		t.Fatalf("expected TLS warning omitted for JSON output, got stderr: %q", stderr)
	}

	var msg logs.Message
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &msg); err != nil {
		t.Fatalf("expected valid JSON log stdout, got %q: %s", stdout, err)
	}
	if msg.Text != "hello" {
		t.Fatalf("expected log text %q, got %q", "hello", msg.Text)
	}
}

func TestLogsTextWarningUsesStdout(t *testing.T) {
	resetJSONCommandTestState(t)

	s := newInsecureWarningHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(logs.Message{Name: "fn", Text: "hello"})
	}))

	gateway = s.URL
	logFlagValues.timeFormat = flags.TimeFormat("")
	logFlagValues.tail = false
	logFlagValues.lines = -1

	stdout, stderr := captureStdoutStderr(t, func() {
		cmd := newLogsTestCommand()

		if err := runLogs(cmd, []string{"fn"}); err != nil {
			t.Fatalf("runLogs returned error: %s", err)
		}
	})

	if !strings.Contains(stdout, "hello") {
		t.Fatalf("expected log output on stdout, got: %q", stdout)
	}
	if !strings.Contains(stdout, NoTLSWarn) {
		t.Fatalf("expected TLS warning on stdout, got: %q", stdout)
	}
	if strings.Contains(stderr, NoTLSWarn) {
		t.Fatalf("expected no TLS warning on stderr, got: %q", stderr)
	}
}

func resetJSONCommandTestState(t *testing.T) {
	t.Helper()

	reset := func() {
		resetForTest()
		jsonOutput = false
		gateway = defaultGateway
		tlsInsecure = false
		token = ""
		functionNamespace = ""
		storeAddress = defaultStore
		platformValue = ""
		verbose = true
		logFlagValues = logFlags{}
	}

	reset()

	t.Setenv("NO_PROXY", "*")
	t.Setenv("no_proxy", "*")

	t.Cleanup(reset)
}

func newLogsTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("namespace", "", "")
	return cmd
}

func newInsecureWarningHTTPServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	listener, err := net.Listen("tcp4", "127.0.0.2:0")
	if err != nil {
		t.Fatalf("listen on 127.0.0.2:0: %s", err)
	}

	s := httptest.NewUnstartedServer(handler)
	s.Listener = listener
	s.Start()

	t.Cleanup(s.Close)

	return s
}

func captureStdoutStderr(t *testing.T, f func()) (string, string) {
	t.Helper()

	stdout := os.Stdout
	stderr := os.Stderr

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %s", err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %s", err)
	}

	os.Stdout = outWriter
	os.Stderr = errWriter

	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		_ = outReader.Close()
		_ = errReader.Close()
	}()

	f()

	_ = outWriter.Close()
	_ = errWriter.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	_, _ = io.Copy(&out, outReader)
	_, _ = io.Copy(&errOut, errReader)

	return out.String(), errOut.String()
}
