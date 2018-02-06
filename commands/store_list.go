// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// Setup flags used by store command
	storeListCmd.Flags().BoolVarP(&verboseDescription, "verbose", "v", false, "Verbose output for the field values")

	storeCmd.AddCommand(storeListCmd)
}

var storeListCmd = &cobra.Command{
	Use:     `list [--url STORE_URL]`,
	Short:   "List OpenFaaS store items",
	Long:    "Lists the available items in OpenFaas store",
	Example: `  faas-cli store list --url https://domain:port/store.json`,
	RunE:    runStoreList,
}

func runStoreList(cmd *cobra.Command, args []string) error {
	items, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("The store is empty.")
		return nil
	}

	fmt.Print(storeRenderItems(items))

	return nil
}
