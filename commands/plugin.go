package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {

	faasCmd.AddCommand(pluginCmd)
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long:  `Manage plugins`,
	RunE:  runPlugin,
}

// preRunPublish validates args & flags
func runPlugin(cmd *cobra.Command, args []string) error {

	fmt.Println("Run plugin get --help")

	return nil
}
