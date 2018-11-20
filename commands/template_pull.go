// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package commands

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

const (
	gitRemoteRepoRegex = `(?:git|ssh|https?|git@[-\w.]+):(\/\/)?(.*?)(\.git)?(\/?|\#[-\d\w._]+?)$`
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
		var validURL = regexp.MustCompile(gitRemoteRepoRegex)
		if !validURL.MatchString(repository) {
			return fmt.Errorf("The repository URL must be a valid git repo uri")
		}
	}

	fmt.Println("Fetch templates from repository: " + repository)
	if err := fetchTemplates(repository, overwrite); err != nil {
		return fmt.Errorf("error while fetching templates: %s", err)
	}
	return nil
}

func pullDebugPrint(message string) {
	if pullDebug {
		fmt.Println(message)
	}
}
