package commands

import "testing"

func TestApplyRemoteBuilderEnvironmentUsesEnvFallbacks(t *testing.T) {
	t.Setenv(remoteBuilderEnvironment, "http://builder.example.com")
	t.Setenv(payloadSecretEnvironment, "/var/run/secrets/payload-secret")
	t.Setenv(builderPublicKeyEnvironment, "/var/run/secrets/builder-public-key")
	t.Setenv(builderKeyIDEnvironment, "builder-key-1")

	remoteBuilder = ""
	payloadSecretPath = ""
	builderPublicKeyPath = ""
	builderKeyID = ""

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
	if builderKeyID != "builder-key-1" {
		t.Fatalf("want builderKeyID from env, got %q", builderKeyID)
	}
}

func TestApplyRemoteBuilderEnvironmentPreservesFlags(t *testing.T) {
	t.Setenv(remoteBuilderEnvironment, "http://builder.example.com")
	t.Setenv(payloadSecretEnvironment, "/var/run/secrets/payload-secret")
	t.Setenv(builderPublicKeyEnvironment, "/var/run/secrets/builder-public-key")
	t.Setenv(builderKeyIDEnvironment, "builder-key-1")

	remoteBuilder = "http://flag-builder.example.com"
	payloadSecretPath = "/tmp/payload-secret"
	builderPublicKeyPath = "/tmp/public-key"
	builderKeyID = "flag-key-id"

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
	if builderKeyID != "flag-key-id" {
		t.Fatalf("want builderKeyID flag value, got %q", builderKeyID)
	}
}
