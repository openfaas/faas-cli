package commands

import (
	"testing"
)

func TestValidateLanguageFlag_dockerfile(t *testing.T) {
	_, err := validateLanguageFlag("dockerfile")
	if err != nil {
		t.Logf("Received unexpected error for lang: %s, %s", "dockerfile", err)
		t.Fail()
	}
}

func TestValidateLanguageFlag_Dockerfile(t *testing.T) {
	_, err := validateLanguageFlag("Dockerfile")
	if err == nil {
		t.Logf("Should have received error for lang: %s, %s", "Dockerfile", err)
		t.Fail()
	}
}
