// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	storeV2 "github.com/openfaas/faas-cli/schema/store/v2"
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags used by store command
	storeListCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output to ssee the full description of each function in the store")

	storeCmd.AddCommand(storeListCmd)
}

var storeListCmd = &cobra.Command{
	Use:     `list [--url STORE_URL]`,
	Aliases: []string{"ls"},
	Short:   "List available OpenFaaS functions in a store",
	Example: `  faas-cli store list
  faas-cli store list --verbose
  faas-cli store list --url https://domain:port/store.json`,
	RunE: runStoreList,
}

func runStoreList(cmd *cobra.Command, args []string) error {
	targetPlatform := getTargetPlatform(platformValue)

	storeList, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	filteredFunctions := filterStoreList(storeList, targetPlatform)

	if len(filteredFunctions) == 0 {
		availablePlatforms := getStorePlatforms(storeList)
		fmt.Printf("No functions found in the store for platform '%s', try one of the following: %s\n", targetPlatform, strings.Join(availablePlatforms, ", "))
		return nil
	}

	fmt.Print(storeRenderItems(filteredFunctions))

	return nil
}

func storeRenderItems(items []storeV2.StoreFunction) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION")

	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\n", item.Title, storeRenderDescription(item.Description))
	}

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}

func storeRenderDescription(descr string) string {
	if !verbose && len(descr) > maxDescriptionLen {
		return descr[0:maxDescriptionLen-3] + "..."
	}

	return descr
}
