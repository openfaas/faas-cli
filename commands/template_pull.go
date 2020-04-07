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
	Use:   `pull [REPOSITORY_URL [TEMPLATE_PATH]] `,
	Short: `Downloads templates from the specified git repo`,
	Long: `Downloads templates from the specified git repo specified by [REPOSITORY_URL], and copies the 'template'
directory from the root of the repo, if it exists.

[REPOSITORY_URL] may specify a specific branch or tag to copy by adding a URL fragment with the branch or tag name.
[TEMPLATE_PATH] may specify an alternate 'template' directory path, relative to the root of the repo.
	`,
	Example: `
  faas-cli template pull https://github.com/openfaas/templates
  faas-cli template pull https://github.com/openfaas/templates#1.0
  faas-cli template pull https://github.com/openfaas/templates path/to/template
`,
	RunE: runTemplatePull,
}

func runTemplatePull(cmd *cobra.Command, args []string) error {
	repository := ""
	path := ""
	if len(args) > 0 {
		repository = args[0]
		if len(args) > 1 {
			path = args[1]
		}
	}
	repository = getTemplateURL(repository, os.Getenv(templateURLEnvironment), DefaultTemplateRepository)
	return pullTemplatePath(repository, path)
}

func pullDebugPrint(message string) {
	if pullDebug {
		fmt.Println(message)
	}
}
