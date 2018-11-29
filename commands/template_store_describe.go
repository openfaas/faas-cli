// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	templateStoreDescribeCmd.PersistentFlags().StringVarP(&templateStoreURL, "url", "u", DefaultTemplatesStore, "Use as alternative store for templates")

	templateStoreCmd.AddCommand(templateStoreDescribeCmd)
}

var templateStoreDescribeCmd = &cobra.Command{
	Use:   `describe`,
	Short: `Describe the template`,
	Long:  `Describe the template by outputting all the fields that the template struct has`,
	Example: `  faas-cli template store describe golang-http
  faas-cli template store describe haskell --url https://raw.githubusercontent.com/custom/store/master/templates.json`,
	RunE: runTemplateStoreDescribe,
}

func runTemplateStoreDescribe(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("\nNeed to specify one of the store templates, check available ones by running the command:\n\nfaas-cli template store list\n")
	}
	if len(args) > 1 {
		return fmt.Errorf("\nNeed to specify single template from the store, check available ones by running the command:\n\nfaas-cli template store list\n")
	}
	envTemplateRepoStore := os.Getenv(templateStoreURLEnvironment)
	storeURL := getTemplateStoreURL(templateStoreURL, envTemplateRepoStore, DefaultTemplatesStore)

	templatesInfo, templatesErr := getTemplateInfo(storeURL)
	if templatesErr != nil {
		return fmt.Errorf("error while getting templates info: %s", templatesErr)
	}
	template := args[0]
	storeTemplate, templateErr := checkExistingTemplate(templatesInfo, template)
	if templateErr != nil {
		return fmt.Errorf("error while searching for template in store: %s", templateErr.Error())
	}

	templateInfo := formatTemplateOutput(storeTemplate)
	fmt.Fprintf(cmd.OutOrStdout(), "%s", templateInfo)

	return nil
}

func checkExistingTemplate(storeTemplates []TemplateInfo, template string) (TemplateInfo, error) {
	var existingTemplate TemplateInfo
	for _, storeTemplate := range storeTemplates {
		sourceName := fmt.Sprintf("%s/%s", storeTemplate.Source, storeTemplate.TemplateName)
		if template == storeTemplate.TemplateName || template == sourceName {
			existingTemplate = storeTemplate
			return existingTemplate, nil
		}
	}
	return existingTemplate, fmt.Errorf("template with name: `%s` does not exist in the store", template)
}

func formatTemplateOutput(storeTemplate TemplateInfo) string {
	var buff bytes.Buffer
	lineWriter := tabwriter.NewWriter(&buff, 0, 0, 1, ' ', 0)
	fmt.Fprintln(lineWriter)
	fmt.Fprintf(lineWriter, "Name:\t%s\n", storeTemplate.TemplateName)
	fmt.Fprintf(lineWriter, "Platform:\t%s\n", storeTemplate.Platform)
	fmt.Fprintf(lineWriter, "Language:\t%s\n", storeTemplate.Language)
	fmt.Fprintf(lineWriter, "Source:\t%s\n", storeTemplate.Source)
	fmt.Fprintf(lineWriter, "Description:\t%s\n", storeTemplate.Description)
	fmt.Fprintf(lineWriter, "Repository:\t%s\n", storeTemplate.Repository)
	fmt.Fprintf(lineWriter, "Official Template:\t%s\n", storeTemplate.Official)
	fmt.Fprintln(lineWriter)

	lineWriter.Flush()

	return buff.String()
}
