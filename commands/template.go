// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(templateCmd)
}

// templateCmd allows access to store and pull commands
var templateCmd = &cobra.Command{
	Use:   `template [COMMAND]`,
	Short: "OpenFaaS template store and pull commands",
	Long:  "Allows browsing templates from store or pulling custom templates",
	Example: `  faas-cli template pull https://github.com/custom/template
  faas-cli template store list
  faas-cli template store ls
  faas-cli template store pull ruby-http
  faas-cli template store pull openfaas-incubator/ruby-http`,
}
