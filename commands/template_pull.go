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
	gitRemoteRepoRegex = `(?:git|ssh|https?|git@[-\w.]+):(\/\/)?(.*?)(\.git)(\/?|\#[-\d\w._]+?)$`
)

var (
	repository string
	overwrite  bool
)

var supportedVerbs = [...]string{"pull"}

func init() {
	templatePullCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing templates?")

	faasCmd.AddCommand(templatePullCmd)
}

// templatePullCmd allows the user to fetch a template from a repository
var templatePullCmd = &cobra.Command{
	Use: "template pull <repository URL>",
	Args: func(cmd *cobra.Command, args []string) error {
		msg := fmt.Sprintf(`Must use a supported verb for 'faas-cli template'
Currently supported verbs: %v`, supportedVerbs)

		if len(args) == 0 {
			return fmt.Errorf(msg)
		}

		if args[0] != "pull" {
			return fmt.Errorf(msg)
		}

		if len(args) > 1 {

			// assume it is a local repo
			if _, err := os.Stat(args[1]); err == nil {
				return nil
			}

			var validURL = regexp.MustCompile(gitRemoteRepoRegex)
			if !validURL.MatchString(args[1]) {
				return fmt.Errorf("The repository URL must be a valid git repo uri")
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

func runTemplatePull(cmd *cobra.Command, args []string) {
	repository := ""
	if len(args) > 1 {
		repository = args[1]
	}

	fmt.Println("Fetch templates from repository: " + repository)
	if err := fetchTemplates(repository, overwrite); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
