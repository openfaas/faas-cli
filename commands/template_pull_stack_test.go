package commands

import (
	"reflect"
	"testing"

	"github.com/openfaas/faas-cli/stack"
)

func Test_findTemplate(t *testing.T) {
	tests := []struct {
		title             string
		desiredTemplate   string
		existingTemplates []stack.TemplateSource
		expectedTemplate  *stack.TemplateSource
	}{
		{
			title:           "Desired template is found",
			desiredTemplate: "powershell",
			existingTemplates: []stack.TemplateSource{
				{Name: "powershell", Source: "exampleURL"},
				{Name: "rust", Source: "exampleURL"},
			},
			expectedTemplate: &stack.TemplateSource{Name: "powershell", Source: "exampleURL"},
		},
		{
			title:           "Desired template is not found",
			desiredTemplate: "golang",
			existingTemplates: []stack.TemplateSource{
				{Name: "powershell", Source: "exampleURL"},
				{Name: "rust", Source: "exampleURL"},
			},
			expectedTemplate: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := findTemplate(test.existingTemplates, test.desiredTemplate)
			if !reflect.DeepEqual(result, test.expectedTemplate) {
				t.Errorf("Wanted template: `%s` got `%s`", test.expectedTemplate, result)
			}
		})
	}
}

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
