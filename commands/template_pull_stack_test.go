package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
)

func Test_pullAllTemplates(t *testing.T) {
	tests := []struct {
		title             string
		existingTemplates []stack.TemplateSource
		expectedError     bool
	}{
		{
			title: "Pull specific Template",
			existingTemplates: []stack.TemplateSource{
				{Name: "my_powershell", Source: "https://github.com/openfaas-incubator/powershell-http-template"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			expectedError: false,
		},
		{
			title: "Pull all templates",
			existingTemplates: []stack.TemplateSource{
				{Name: "my_powershell", Source: "https://github.com/openfaas-incubator/powershell-http-template"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			expectedError: false,
		},
		{
			title: "Pull custom template and template from store without source",
			existingTemplates: []stack.TemplateSource{
				{Name: "perl-alpine"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			expectedError: false,
		},
		{
			title: "Pull non-existant template",
			existingTemplates: []stack.TemplateSource{
				{Name: "my_powershell", Source: "invalidURL"},
				{Name: "my_rust", Source: "https://github.com/openfaas-incubator/openfaas-rust-template"},
			},
			expectedError: true,
		},
		{
			title: "Pull template with invalid URL",
			existingTemplates: []stack.TemplateSource{
				{Name: "my_powershell", Source: "invalidURL"},
			},
			expectedError: true,
		},
		{
			title: "Pull template which does not exist in store",
			existingTemplates: []stack.TemplateSource{
				{Name: "my_powershell"},
			},
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			actualError := pullStackTemplates(test.existingTemplates, templatePullStackCmd)
			if actualError != nil && test.expectedError == false {
				t.Errorf("Unexpected error: %s", actualError.Error())
			}
		})
	}
}

func Test_filterExistingTemplates(t *testing.T) {
	templatesDir := "./template"
	defer os.RemoveAll(templatesDir)

	templates := []stack.TemplateSource{
		{Name: "dockerfile", Source: "https://github.com/openfaas/templates"},
		{Name: "ruby", Source: "https://github.com/openfaas/classic-templates"},
		{Name: "perl", Source: "https://github.com/openfaas-incubator/perl-template"},
	}

	// Copy the submodule to temp directory to avoid altering it during tests
	testRepoGit := filepath.Join("testdata", "templates", "template")
	builder.CopyFiles(testRepoGit, templatesDir)

	newTemplateInfos, err := filterExistingTemplates(templates, templatesDir)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if len(newTemplateInfos) != 1 {
		t.Errorf("Wanted new templates: `%d` got `%d`", 1, len(newTemplateInfos))
	}

	if newTemplateInfos[0].Name != "perl" {
		t.Errorf("Wanted template: `%s` got `%s`", "perl", newTemplateInfos[0].Name)
	}
}
