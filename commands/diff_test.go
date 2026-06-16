package commands

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/test"
	types "github.com/openfaas/faas-provider/types"
)

func Test_diff_no_changes(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:latest
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:            "myfunc",
			Image:           "ttl.sh/test/myfunc:latest",
			Replicas:        1,
			InvocationCount: 0,
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	var executeErr error
	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		executeErr = faasCmd.Execute()
	})

	if executeErr != nil {
		t.Fatalf("Expected clean diff to return nil, got: %s", executeErr)
	}

	if !strings.Contains(stdOut, "no differences found") {
		t.Fatalf("Expected no differences, got:\n%s", stdOut)
	}
}

func Test_diff_no_changes_with_deployed_namespace(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:latest
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:      "myfunc",
			Namespace: "openfaas-fn",
			Image:     "ttl.sh/test/myfunc:latest",
			Replicas:  1,
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if strings.Contains(stdOut, "not deployed") {
		t.Fatalf("Expected namespaced deployed function to match unnamespaced stack function, got:\n%s", stdOut)
	}

	if !strings.Contains(stdOut, "no differences found") {
		t.Fatalf("Expected no differences, got:\n%s", stdOut)
	}
}

func Test_diff_image_changed(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:v1.0.0
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:            "myfunc",
			Image:           "ttl.sh/test/myfunc:v2.0.0",
			Replicas:        1,
			InvocationCount: 5,
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	var executeErr error
	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		executeErr = faasCmd.Execute()
	})

	if executeErr == nil || !strings.Contains(executeErr.Error(), "differences found") {
		t.Fatalf("Expected changed diff to return differences found, got: %v", executeErr)
	}

	if !strings.Contains(stdOut, "-") || !strings.Contains(stdOut, "+") {
		t.Fatalf("Expected diff output with changes, got:\n%s", stdOut)
	}

	if !strings.Contains(stdOut, "ttl.sh/test/myfunc:v1.0.0") {
		t.Fatalf("Expected YAML image in diff, got:\n%s", stdOut)
	}

	if !strings.Contains(stdOut, "ttl.sh/test/myfunc:v2.0.0") {
		t.Fatalf("Expected deployed image in diff, got:\n%s", stdOut)
	}
}

func Test_diff_tag_digest_matches_deployed_image(t *testing.T) {
	tmpDir := t.TempDir()
	handlerDir := filepath.Join(tmpDir, "handler")
	if err := os.Mkdir(handlerDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(handlerDir, "handler.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, digest, err := builder.GetImageTagValues(schema.DigestFormat, handlerDir)
	if err != nil {
		t.Fatal(err)
	}
	imageName := schema.BuildImageName(schema.DigestFormat, "ttl.sh/test/myfunc:latest", digest, "")

	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ` + handlerDir + `
    image: ttl.sh/test/myfunc:latest
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:       "myfunc",
			Image:      imageName,
			EnvProcess: "python index.py",
			EnvVars: map[string]string{
				"LANG":           "C.UTF-8",
				"PYTHON_VERSION": "3.12.13",
				"mode":           "http",
				"upstream_url":   "http://127.0.0.1:5000",
			},
			Replicas: 1,
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	var executeErr error
	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
			"--tag=digest",
		})
		executeErr = faasCmd.Execute()
	})

	if executeErr != nil {
		t.Fatalf("Expected digest-tagged image to match deployed image, got: %s\n%s", executeErr, stdOut)
	}

	if !strings.Contains(stdOut, "no differences found") {
		t.Fatalf("Expected no differences, got:\n%s", stdOut)
	}
}

func Test_diff_env_all_reports_deployed_only_env(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:latest
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:  "myfunc",
			Image: "ttl.sh/test/myfunc:latest",
			EnvVars: map[string]string{
				"LANG": "C.UTF-8",
			},
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	var executeErr error
	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
			"--env=all",
		})
		executeErr = faasCmd.Execute()
	})

	if executeErr == nil || !strings.Contains(executeErr.Error(), "differences found") {
		t.Fatalf("Expected env=all to return differences found, got: %v", executeErr)
	}

	if !strings.Contains(stdOut, "env.LANG:") {
		t.Fatalf("Expected env=all to report deployed-only env var, got:\n%s", stdOut)
	}
}

func Test_diff_env_invalid(t *testing.T) {
	resetForTest()

	faasCmd.SetArgs([]string{
		"diff",
		"-f", "stack.yaml",
		"--env=invalid",
	})

	err := faasCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid env mode, got nil")
	}

	if !strings.Contains(err.Error(), "unknown env diff mode") {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}
}

func Test_diff_env_changed(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:latest
    environment:
      LOG_LEVEL: debug
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:            "myfunc",
			Image:           "ttl.sh/test/myfunc:latest",
			Replicas:        1,
			InvocationCount: 0,
			EnvVars: map[string]string{
				"LOG_LEVEL": "info",
			},
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if !strings.Contains(stdOut, "LOG_LEVEL") {
		t.Fatalf("Expected LOG_LEVEL in diff output, got:\n%s", stdOut)
	}

	if !strings.Contains(stdOut, "debug") {
		t.Fatalf("Expected YAML env value 'debug' in diff output, got:\n%s", stdOut)
	}
}

func Test_diff_deploy_spec_changed(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:latest
    fprocess: ./handler
    readonly_root_filesystem: true
    secrets:
      - db-password
    constraints:
      - zone=east
    labels:
      com.openfaas.scale.min: "1"
    annotations:
      topic: payments
    limits:
      cpu: 500m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	labels := map[string]string{
		"com.openfaas.scale.min": "2",
	}
	annotations := map[string]string{
		"topic": "orders",
	}
	expectedFunctions := []types.FunctionStatus{
		{
			Name:                   "myfunc",
			Image:                  "ttl.sh/test/myfunc:latest",
			EnvProcess:             "node index.js",
			ReadOnlyRootFilesystem: false,
			Secrets:                []string{"api-key"},
			Constraints:            []string{"zone=west"},
			Labels:                 &labels,
			Annotations:            &annotations,
			Limits:                 &types.FunctionResources{CPU: "1", Memory: "512Mi"},
			Requests:               &types.FunctionResources{CPU: "250m", Memory: "256Mi"},
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	expectedParts := []string{
		"fprocess:",
		"./handler",
		"node index.js",
		"readonly_root_filesystem:",
		"true",
		"false",
		"secret.db-password:",
		"secret.api-key:",
		"constraint.zone=east:",
		"constraint.zone=west:",
		"label.com.openfaas.scale.min:",
		"annotation.topic:",
		"limits.cpu:",
		"limits.memory:",
		"requests.cpu:",
		"requests.memory:",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(stdOut, expected) {
			t.Fatalf("Expected deploy spec diff output to contain %q, got:\n%s", expected, stdOut)
		}
	}
}

func Test_diff_function_not_deployed(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  missing-func:
    lang: golang-middleware
    handler: ./missing-func
    image: ttl.sh/test/missing-func:latest
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if !strings.Contains(stdOut, "missing-func") {
		t.Fatalf("Expected missing-func in diff output, got:\n%s", stdOut)
	}

	if !strings.Contains(stdOut, "defined in stack.yaml") || !strings.Contains(stdOut, "not deployed") {
		t.Fatalf("Expected missing function status in diff output, got:\n%s", stdOut)
	}
}

func Test_diff_ignores_extra_deployed_function(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "stack.yaml")
	yamlContent := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  myfunc:
    lang: golang-middleware
    handler: ./myfunc
    image: ttl.sh/test/myfunc:latest
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	expectedFunctions := []types.FunctionStatus{
		{
			Name:            "myfunc",
			Image:           "ttl.sh/test/myfunc:latest",
			Replicas:        1,
			InvocationCount: 0,
		},
		{
			Name:            "extra-func",
			Image:           "ttl.sh/test/extra-func:latest",
			Replicas:        1,
			InvocationCount: 10,
		},
	}

	s := test.MockHttpServer(t, []test.Request{
		{
			Method:             http.MethodGet,
			Uri:                "/system/functions",
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       expectedFunctions,
		},
	})
	defer s.Close()

	resetForTest()

	stdOut := test.CaptureStdout(func() {
		faasCmd.SetArgs([]string{
			"diff",
			"-f", yamlPath,
			"--gateway=" + s.URL,
		})
		faasCmd.Execute()
	})

	if strings.Contains(stdOut, "extra-func") {
		t.Fatalf("Expected extra deployed function to be ignored, got:\n%s", stdOut)
	}

	if !strings.Contains(stdOut, "no differences found") {
		t.Fatalf("Expected no differences for matching YAML function, got:\n%s", stdOut)
	}
}

func Test_diff_no_yaml_file(t *testing.T) {
	resetForTest()

	faasCmd.SetArgs([]string{
		"diff",
		"--gateway=http://127.0.0.1:8080",
	})

	err := faasCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing YAML file, got nil")
	}

	if !strings.Contains(err.Error(), "no YAML file specified") {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}
}

func Test_printDifftool_side_by_side(t *testing.T) {
	output := test.CaptureStdout(func() {
		printDifftool("myfunc", []diffRow{
			{
				left:  diffCell{marker: "-", field: "image", value: "ttl.sh/test/myfunc:v1.0.0"},
				right: diffCell{marker: "+", field: "image", value: "ttl.sh/test/myfunc:v2.0.0"},
			},
			{
				left:  diffCell{marker: "-", field: "env.LOG_LEVEL", value: "debug"},
				right: diffCell{marker: "+", field: "env.LOG_LEVEL", value: "info"},
			},
		})
	})

	expectedParts := []string{
		"myfunc\n",
		"stack.yaml",
		"deployed",
		" | ",
		"- image:",
		"+ image:",
		"- env.LOG_LEVEL:",
		"+ env.LOG_LEVEL:",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(output, expected) {
			t.Fatalf("Expected side-by-side diff output to contain %q, got:\n%s", expected, output)
		}
	}
}
