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
	storeInspectCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output for the field values")

	storeCmd.AddCommand(storeInspectCmd)
}

var storeInspectCmd = &cobra.Command{
	Use:   `inspect (FUNCTION_NAME|FUNCTION_TITLE) [--url STORE_URL]`,
	Short: "Show details of OpenFaaS function from a store",
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

	item := storeFindFunction(args[0], storeItems)
	if item == nil {
		return fmt.Errorf("function '%s' not found", functionName)
	}

	content := storeRenderItem(item)
	fmt.Print(content)

	return nil
}

func storeRenderItem(item *schema.StoreItem) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION\tIMAGE\tPROCESS\tREPO")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		item.Title,
		storeRenderDescription(item.Description),
		item.Image,
		item.Fprocess,
		item.RepoURL,
	)

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}
