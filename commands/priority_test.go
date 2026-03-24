package commands

import "testing"

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
