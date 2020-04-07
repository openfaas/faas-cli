package commands

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Test_TemplateStoreDescribe(t *testing.T) {
	localTemplateRepository := setupLocalTemplateRepo(t, "")
	defer os.RemoveAll(localTemplateRepository)

	const templatePath = "path/to/template"
	nestedTemplateRepository := setupLocalTemplateRepo(t, templatePath)
	defer os.RemoveAll(nestedTemplateRepository)

	defer tearDownFetchTemplates(t)

	templateInfo := TemplateInfo{
		TemplateName: "ruby",
		Platform:     "x86_64",
		Language:     "Ruby",
		Source:       "openfaas",
		Description:  "Classic Ruby 2.5 template",
		Repository:   "https://github.com/openfaas/templates",
		Official:     "true",
	}

	t.Run("simple", func(t *testing.T) {
		defer tearDownFetchTemplates(t)
		out := strings.TrimSpace(formatTemplateOutput(templateInfo))
		expect := strings.TrimSpace(`
Name:              ruby
Platform:          x86_64
Language:          Ruby
Source:            openfaas
Description:       Classic Ruby 2.5 template
Repository:        https://github.com/openfaas/templates
Official Template: true
`)
		if out != expect {
			t.Errorf("expected %q; got %q", expect, out)
		}
	})

	t.Run("nested", func(t *testing.T) {
		defer tearDownFetchTemplates(t)

		templateInfo := templateInfo
		templateInfo.TemplatePath = templatePath
		out := strings.TrimSpace(formatTemplateOutput(templateInfo))
		expect := strings.TrimSpace(fmt.Sprintf(`
Name:              ruby
Platform:          x86_64
Language:          Ruby
Source:            openfaas
Description:       Classic Ruby 2.5 template
Repository:        https://github.com/openfaas/templates
Path:              %s
Official Template: true
`, templatePath))

		if out != expect {
			t.Errorf("expected %q; got %q", expect, out)
		}
	})
}
