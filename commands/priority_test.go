package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/go-sdk/stack"
)

func TestGetNamespacePrecedence(t *testing.T) {
	tests := []struct {
		name                 string
		flagNamespace        string
		stackNamespace       string
		environmentNamespace string
		want                 string
	}{
		{
			name:                 "environment namespace is the default",
			environmentNamespace: "environment",
			want:                 "environment",
		},
		{
			name:                 "stack namespace overrides environment",
			stackNamespace:       "stack",
			environmentNamespace: "environment",
			want:                 "stack",
		},
		{
			name:                 "flag overrides stack and environment",
			flagNamespace:        "flag",
			stackNamespace:       "stack",
			environmentNamespace: "environment",
			want:                 "flag",
		},
		{
			name: "empty values retain the existing default",
			want: defaultFunctionNamespace,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(openFaaSNamespaceEnvironment, test.environmentNamespace)
			got := getNamespace(test.flagNamespace, test.stackNamespace)
			if got != test.want {
				t.Fatalf("want namespace %q, got %q", test.want, got)
			}
		})
	}
}

func TestGetNamespaceUsesSubstitutedStackNamespaceBeforeEnvironment(t *testing.T) {
	t.Setenv("STACK_NAMESPACE", "substituted-stack")
	t.Setenv(openFaaSNamespaceEnvironment, "environment")
	path := filepath.Join(t.TempDir(), "stack.yaml")
	contents := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  echo:
    image: echo:latest
    namespace: ${STACK_NAMESPACE}
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	services, err := stack.ParseYAMLFile(path, "", "", true)
	if err != nil {
		t.Fatal(err)
	}

	got := getNamespace("", services.Functions["echo"].Namespace)
	if got != "substituted-stack" {
		t.Fatalf("want substituted stack namespace, got %q", got)
	}
}

func TestApplyRemoteBuilderEnvironmentUsesEnvFallbacks(t *testing.T) {
	t.Setenv(remoteBuilderEnvironment, "http://builder.example.com")
	t.Setenv(payloadSecretEnvironment, "/var/run/secrets/payload-secret")
	t.Setenv(builderPublicKeyEnvironment, "/var/run/secrets/builder-public-key")

	remoteBuilder = ""
	payloadSecretPath = ""
	builderPublicKeyPath = ""

	applyRemoteBuilderEnvironment()

	if remoteBuilder != "http://builder.example.com" {
		t.Fatalf("want remoteBuilder from env, got %q", remoteBuilder)
	}
	if payloadSecretPath != "/var/run/secrets/payload-secret" {
		t.Fatalf("want payloadSecretPath from env, got %q", payloadSecretPath)
	}
	if builderPublicKeyPath != "/var/run/secrets/builder-public-key" {
		t.Fatalf("want builderPublicKeyPath from env, got %q", builderPublicKeyPath)
	}
}

func TestApplyRemoteBuilderEnvironmentPreservesFlags(t *testing.T) {
	t.Setenv(remoteBuilderEnvironment, "http://builder.example.com")
	t.Setenv(payloadSecretEnvironment, "/var/run/secrets/payload-secret")
	t.Setenv(builderPublicKeyEnvironment, "/var/run/secrets/builder-public-key")

	remoteBuilder = "http://flag-builder.example.com"
	payloadSecretPath = "/tmp/payload-secret"
	builderPublicKeyPath = "/tmp/public-key"

	applyRemoteBuilderEnvironment()

	if remoteBuilder != "http://flag-builder.example.com" {
		t.Fatalf("want remoteBuilder flag value, got %q", remoteBuilder)
	}
	if payloadSecretPath != "/tmp/payload-secret" {
		t.Fatalf("want payloadSecretPath flag value, got %q", payloadSecretPath)
	}
	if builderPublicKeyPath != "/tmp/public-key" {
		t.Fatalf("want builderPublicKeyPath flag value, got %q", builderPublicKeyPath)
	}
}
