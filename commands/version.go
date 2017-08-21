// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// GitCommit injected at build-time
var GitCommit string

func init() {
	faasCmd.AddCommand(versionCmd)
}

// versionCmd displays version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the clients version information",
	Long: fmt.Sprintf(`The version command returns the current clients version information.

This currently consists of the GitSHA from which the client was built.
- https://github.com/alexellis/faas-cli/tree/%s`, GitCommit),
	Run: runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Git Commit: %s\n", GitCommit)
	return
}
