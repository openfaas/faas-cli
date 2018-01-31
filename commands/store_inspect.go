// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// Setup flags used by store command
	storeInspectCmd.Flags().BoolVarP(&verboseDescription, "verbose", "v", false, "Verbose output for the field values")

	storeCmd.AddCommand(storeInspectCmd)
}

var storeInspectCmd = &cobra.Command{
	Use:   `inspect (FUNCTION_NAME|FUNCTION_TITLE) [--url STORE_URL]`,
	Short: "Show OpenFaaS store function details",
	Long:  "Prints the detailed informations of the specified OpenFaaS function",
	Example: `  faas-cli store inspect NodeInfo
  faas-cli store inspect NodeInfo --url https://domain:port/store.json`,
	RunE: runStoreInspect,
}

func runStoreInspect(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	storeItems, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	item := findFunction(args[0], storeItems)
	if item == nil {
		return fmt.Errorf("function '%s' not found", functionName)
	}

	content := renderStoreItem(item)
	fmt.Print(content)

	return nil
}
