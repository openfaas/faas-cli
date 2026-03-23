// Copyright (c) OpenFaaS Author(s) 2024. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(chartCmd)
}

var chartCmd = &cobra.Command{
	Use:   `chart`,
	Short: "Helm chart commands",
	Long:  "Export and manage OpenFaaS Helm charts",
}
