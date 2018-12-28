// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"fmt"
	"os"

	"github.com/openfaas/faas-cli/versioncontrol"
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
	Short: `Downloads templates from the specified github repo`,
	Long: `Downloads the compressed github repo specified by [URL], and extracts the 'template'
	directory from the root of the repo, if it exists.`,
	Example: "faas-cli template pull https://github.com/openfaas/faas-cli",
	RunE:    runTemplatePull,
}

func runTemplatePull(cmd *cobra.Command, args []string) error {
	repository := ""
	if len(args) > 0 {
		repository = args[0]
	}
	repository = getTemplateURL(repository, os.Getenv(templateURLEnvironment), DefaultTemplateRepository)

	if _, err := os.Stat(repository); err != nil {
		if !versioncontrol.IsGitRemote(repository) && !versioncontrol.IsPinnedGitRemote(repository) {
			return fmt.Errorf("The repository URL must be a valid git repo uri")
		}
	}

	repository, refName := versioncontrol.ParsePinnedRemote(repository)

	if err := versioncontrol.GitCheckRefName.Invoke("", map[string]string{"refname": refName}); err != nil {
		fmt.Printf("Invalid tag or branch name `%s`\n", refName)
		fmt.Println("See https://git-scm.com/docs/git-check-ref-format for more details of the rules Git enforces on branch and reference names.")

		return err
	}

	fmt.Printf("Fetch templates from repository: %s at %s\n", repository, refName)
	if err := fetchTemplates(repository, refName, overwrite); err != nil {
		return fmt.Errorf("error while fetching templates: %s", err)
	}
	return nil
}

func pullDebugPrint(message string) {
	if pullDebug {
		fmt.Println(message)
	}
}
