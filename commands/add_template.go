package commands

import (
	"errors"

	"fmt"

	"github.com/spf13/cobra"
)

var (
	repository string
	overwrite  bool
)

func init() {
	addTemplateCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing templates?")

	faasCmd.AddCommand(addTemplateCmd)
}

// addTemplateCmd allows the user to fetch a template from a repository
var addTemplateCmd = &cobra.Command{
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("A repository URL must be sepecified")
		}
		return nil
	},
	Use:     `add-template <repository_url>`,
	Short:   "Add template OpenFaaS",
	Example: `  faas-cli add-template https://domain/path/myfunctions.yml`,
	Run:     runAddTemplate,
}

func runAddTemplate(cmd *cobra.Command, args []string) {
	repository = args[0]

	if err := fetchTemplates(repository, overwrite); err != nil {
		fmt.Println(err)
	}
}
