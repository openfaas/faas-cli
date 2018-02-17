// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags used by store command
	storeListCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output for the field values")

	storeCmd.AddCommand(storeListCmd)
}

var storeListCmd = &cobra.Command{
	Use:     `list [--url STORE_URL]`,
	Short:   "List available OpenFaaS functions in a store",
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

func storeRenderItems(items []schema.StoreItem) string {
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
