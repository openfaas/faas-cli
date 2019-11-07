// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repository string
	overwrite  bool
	pullDebug  bool
)

func init() {
	templatePullCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing templates?")
	templatePullCmd.Flags().BoolVar(&pullDebug, "debug", false, "Enable debug output")

	templateCmd.AddCommand(templatePullCmd)
}

// templatePullCmd allows the user to fetch a template from a repository
var templatePullCmd = &cobra.Command{
	Use:   `pull [REPOSITORY_URL]`,
	Short: `Downloads templates from the specified git repo`,
	Long: `Downloads templates from the specified git repo specified by [REPOSITORY_URL], and copies the 'template'
directory from the root of the repo, if it exists.

[REPOSITORY_URL] may specify a specific branch or tag to copy by adding a URL fragment with the branch or tag name.
	`,
	Example: `
  faas-cli template pull https://github.com/openfaas/templates
  faas-cli template pull https://github.com/openfaas/templates#1.0
`,
	RunE: runTemplatePull,
}

func runTemplatePull(cmd *cobra.Command, args []string) error {
	repository := ""
	if len(args) > 0 {
		repository = args[0]
	}
	repository = getTemplateURL(repository, os.Getenv(templateURLEnvironment), DefaultTemplateRepository)
	return pullTemplate(repository)
}

func pullDebugPrint(message string) {
	if pullDebug {
		fmt.Println(message)
	}
}
