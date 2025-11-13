package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/go-sdk/stack"
)

func Test_pullStackTemplates(t *testing.T) {
	tests := []struct {
		title            string
		templateSources  []stack.TemplateSource
		missingTemplates []string
		expectedError    bool
	}{
		{
			title: "Pull specific template",
			templateSources: []stack.TemplateSource{
				{Name: "my_powershell", Source: "https://github.com/openfaas-incubator/powershell-http-template"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			missingTemplates: []string{"my_powershell"},
			expectedError:    false,
		},
		{
			title: "Pull all templates",
			templateSources: []stack.TemplateSource{
				{Name: "my_powershell", Source: "https://github.com/openfaas-incubator/powershell-http-template"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			missingTemplates: []string{"my_powershell", "my_rust"},
			expectedError:    false,
		},
		{
			title: "Pull custom template and template from store without source",
			templateSources: []stack.TemplateSource{
				{Name: "perl-alpine"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			missingTemplates: []string{"perl-alpine"},
			expectedError:    false,
		},
		{
			title: "Pull template from invalid URL",
			templateSources: []stack.TemplateSource{
				{Name: "my_powershell", Source: "invalidURL"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			missingTemplates: []string{"my_powershell"},
			expectedError:    true,
		},
		{
			title: "Pull template which does not exist in store, which has no URL given",
			templateSources: []stack.TemplateSource{
				{Name: "my_powershell"},
			},
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			actualError := pullStackTemplates(test.missingTemplates, test.templateSources, templatePullStackCmd)
			if actualError != nil && test.expectedError == false {
				t.Errorf("Unexpected error: %s", actualError.Error())
			}
		})
	}
}

func Test_getMissingTemplates_finds_missing_template(t *testing.T) {
	templatesDir := "./template"
	defer os.RemoveAll(templatesDir)

	stackYaml := `version: 1.0
provider:
  name: openfaas
  gateway: http://127.0.0.1:8080
functions:
  docker-fn:
    lang: dockerfile
    handler: ./dockerfile
  ruby-fn:
    lang: ruby
    handler: ./ruby
    image: ttl.sh/alexellis/ruby:latest
  perl-fn:
    lang: perl
    handler: ./perl
    image: ttl.sh/alexellis/perl:latest

configuration:
  templates:
   - name: dockerfile
     source: https://github.com/openfaas/templates
   - name: ruby
     source: https://github.com/openfaas/classic-templates
   - name: perl
     source: https://github.com/openfaas-incubator/perl-template
`

	parsedFns, err := stack.ParseYAMLData([]byte(stackYaml), "", "", false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	functions := parsedFns.Functions

	// Copy the submodule to temp directory to avoid altering it during tests
	testRepoGit := filepath.Join("testdata", "templates", "template")
	builder.CopyFiles(testRepoGit, templatesDir)

	// Assert that dockerfile and ruby exist, and that perl is missing

	if _, err := os.Stat(filepath.Join(templatesDir, "dockerfile")); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if _, err := os.Stat(filepath.Join(templatesDir, "ruby")); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if _, err := os.Stat(filepath.Join(templatesDir, "perl")); err == nil {
		t.Errorf("perl should not exist at this point")
	}

	newTemplateInfos, err := getMissingTemplates(functions, templatesDir)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if len(newTemplateInfos) != 1 {
		t.Errorf("Wanted new templates: `%d` got `%d`", 1, len(newTemplateInfos))
	}

	wantMissingName := "perl"
	gotMissingName := newTemplateInfos[0]
	if gotMissingName != wantMissingName {
		t.Errorf("Wanted template: `%s` got `%s`", wantMissingName, gotMissingName)
	}
}
