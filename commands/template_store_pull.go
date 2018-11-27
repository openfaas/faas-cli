// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	templateStorePullCmd.PersistentFlags().StringVarP(&templateStoreURL, "url", "u", DefaultTemplatesStore, "Use as alternative store for templates")
	templatePull, _, _ := faasCmd.Find([]string{"template", "pull"})
	templateStoreCmd.PersistentFlags().AddFlagSet(templatePull.Flags())

	templateStoreCmd.AddCommand(templateStorePullCmd)
}

// templateStorePullCmd pulls templates from default store or custom store if set
var templateStorePullCmd = &cobra.Command{
	Use:   `pull [TEMPLATE_NAME]`,
	Short: `Pull templates from store`,
	Long:  `Pull templates from store supported by openfaas or openfaas-incubator organizations or your custom store`,
	Example: `  faas-cli template store pull ruby-http ...
  faas-cli template store pull openfaas/go openfaas-incubator/node10-express ...
  faas-cli template store pull go ... --debug
  faas-cli template store pull openfaas/go openfaas-incubator/node10-express --overwrite
  faas-cli template store pull golang-middleware --url https://raw.githubusercontent.com/openfaas/store/master/templates.json`,
	RunE: runTemplateStorePull,
}

func runTemplateStorePull(cmd *cobra.Command, args []string) error {
	var nonExistingTemplates []string
	if len(args) == 0 {
		return fmt.Errorf("\nNeed to specify one of the store templates check available ones by running the command:\n\nfaas-cli template store list\n")
	}

	envTemplateRepoStore := os.Getenv(templateStoreURLEnvironment)
	storeURL := getTemplateStoreURL(templateStoreURL, envTemplateRepoStore, DefaultTemplatesStore)

	storeTemplates, templatesErr := getTemplateInfo(storeURL)
	if templatesErr != nil {
		return fmt.Errorf("error while fetching templates from store: %s", templatesErr)
	}

	for _, template := range args {
		found := false
		for _, storeTemplate := range storeTemplates {
			sourceName := fmt.Sprintf("%s/%s", storeTemplate.Source, storeTemplate.TemplateName)
			if template == storeTemplate.TemplateName || template == sourceName {
				err := runTemplatePull(cmd, []string{storeTemplate.Repository})
				if err != nil {
					return fmt.Errorf("error while pulling template: %s : %s", storeTemplate.TemplateName, err.Error())
				}
				found = true
				break
			}
		}
		if !found {
			nonExistingTemplates = append(nonExistingTemplates, fmt.Sprintf("\nThere is no template with name: `%s` in the store.\n", template))
		}
	}
	if len(nonExistingTemplates) > 0 {
		nonExistingTemplates = append(nonExistingTemplates, "\ncheck available options by running:\n\nfaas-cli template store pull <NAME>|<SOURCE/NAME>\n")
		return fmt.Errorf("%s", strings.Join(nonExistingTemplates, ""))
	}

	return nil
}
