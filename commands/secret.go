// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(secretCmd)
}

var secretCmd = &cobra.Command{
	Use:   `secret`,
	Short: "OpenFaaS secret commands",
	Long:  "Manage function secrets",
}
