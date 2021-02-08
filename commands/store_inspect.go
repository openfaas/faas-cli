// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	storeV2 "github.com/openfaas/faas-cli/schema/store/v2"
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

	targetPlatform := getTargetPlatform(platformValue)
	storeItems, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	platformFunctions := filterStoreList(storeItems, targetPlatform)

	functionName := args[0]
	item := storeFindFunction(functionName, platformFunctions)
	if item == nil {
		return fmt.Errorf("function '%s' not found for platform '%s'", functionName, targetPlatform)
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	content := storeRenderItem(item, targetPlatform, verbose)
	fmt.Print(content)

	return nil
}

func storeRenderItem(item *storeV2.StoreFunction, platform string, verbose bool) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Info for: %s\n\n", item.Title)

	fmt.Fprintf(w, "%s\t%s\n", "Name", item.Name)

	desc := item.Description
	if !verbose {
		desc = storeRenderDescription(desc)
	}

	fmt.Fprintf(w, "%s\t%s\n", "Description", desc)
	fmt.Fprintf(w, "%s\t%s\n", "Image", item.GetImageName(platform))
	fmt.Fprintf(w, "%s\t%s\n", "Process", item.Fprocess)
	fmt.Fprintf(w, "%s\t%s\n", "Repo URL", item.RepoURL)

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}
