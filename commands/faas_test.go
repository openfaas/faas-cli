package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadsDefaultYAMLWhenPresent(t *testing.T) {
	rs := resetState()
	defer rs()
	os.Chdir("./testdata")

	Execute([]string{"help"})

	if yamlFile != "stack.yml" {
		t.Fatalf("Expected yamlFile to equal %v got %v\n", "stack.yml", yamlFile)
	}
}

func TestLoadsFromParmetersYAMLWhenPresentAndDefaultYAMLFileAlsoPresent(t *testing.T) {
	rs := resetState()
	defer rs()
	os.Chdir("./testdata")

	Execute([]string{"help", "--yaml=myfile.yml"})

	if yamlFile != "myfile.yml" {
		t.Fatalf("Expected yamlFile to equal %v got %v\n", "stack.yml", yamlFile)
	}
}

func TestDoesNotLoadDefaultYAMLWhenMissing(t *testing.T) {
	rs := resetState()
	defer rs()

	Execute([]string{"help"})

	if yamlFile != "" {
		t.Fatalf("Expected yamlFile to be blank got %v\n", yamlFile)
	}
}

func resetState() func() {
	faasCmd.SetOutput(ioutil.Discard)
	dir, _ := filepath.Abs("./")
	return func() {
		os.Chdir(dir)
		yamlFile = ""
	}
}
