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
	overwrite bool
	pullDebug bool
)

func init() {
	templateCmd := newTemplateCmd()
	templateCmd.AddCommand(newTemplatePullCmd())

	faasCmd.AddCommand(templateCmd)
}

// newTemplatePullCmd creates a new 'template' command
func newTemplateCmd() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage templates",
	}

	return templateCmd
}

// newTemplatePullCmd creates a new 'template pull' command which allows the user to fetch a template from a repository
func newTemplatePullCmd() *cobra.Command {
	templatePullCmd := &cobra.Command{
		Use: "pull <repository URL>",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				// assume it is a local repo
				if _, err := os.Stat(args[0]); err == nil {
					return nil
				}

				var validURL = regexp.MustCompile(gitRemoteRepoRegex)
				if !validURL.MatchString(args[0]) {
					return fmt.Errorf("the repository URL must be a valid git repo uri")
				}
			}

			return nil
		},
		Short: "Downloads templates from the specified github repo",
		Long: `Downloads the compressed github repo specified by [URL], and extracts the 'template'
	directory from the root of the repo, if it exists.`,
		Example: "faas-cli template pull https://github.com/openfaas/faas-cli",
		Run:     runTemplatePull,
	}

	templatePullCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing templates?")
	templatePullCmd.Flags().BoolVar(&pullDebug, "debug", false, "Enable debug output")

	return templatePullCmd
}

func runTemplatePull(cmd *cobra.Command, args []string) {
	repository := ""
	if len(args) == 1 {
		repository = args[0]
	}

	fmt.Println("Fetch templates from repository: " + repository)
	if err := fetchTemplates(repository, overwrite); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}

func pullDebugPrint(message string) {
	if pullDebug {
		fmt.Println(message)
	}
}
