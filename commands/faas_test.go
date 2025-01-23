package commands

import (
	"io"
	"os"
	"testing"
)

var mockStatParams string

func setupFaas(statError error) {
	yamlFile = ""
	mockStatParams = ""
	faasCmd.SetOutput(io.Discard)

	stat = func(f string) (os.FileInfo, error) {
		mockStatParams = f
		return nil, statError
	}
}

func TestCallsStatWithDefaultYAMLFileName(t *testing.T) {
	setupFaas(nil)

	Execute([]string{"help"})

	if mockStatParams != defaultYAML {
		t.Fatalf("Expected yamlFile to equal %v got %v\n", defaultYAML, yamlFile)
	}
}

func TestLoadsDefaultYAMLWhenPresent(t *testing.T) {
	setupFaas(nil)

	Execute([]string{"help"})

	if yamlFile != defaultYAML {
		t.Fatalf("Expected yamlFile to equal %v got %v\n", defaultYAML, yamlFile)
	}
}

func TestLoadsFromParmetersYAMLWhenPresentAndDefaultYAMLFileAlsoPresent(t *testing.T) {
	setupFaas(nil)

	Execute([]string{"help", "--yaml=myfile.yml"})

	if yamlFile != "myfile.yml" {
		t.Fatalf("Expected yamlFile to equal %v got %v\n", "myfile.yml", yamlFile)
	}
}

func TestDoesNotLoadDefaultYAMLWhenMissing(t *testing.T) {
	setupFaas(os.ErrNotExist)

	Execute([]string{"help"})

	if yamlFile != "" {
		t.Fatalf("Expected yamlFile to be blank got %v\n", yamlFile)
	}
}
