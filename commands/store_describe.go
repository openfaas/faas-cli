// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/mitchellh/go-wordwrap"
	"github.com/openfaas/faas-cli/proxy"
	storeV2 "github.com/openfaas/faas-cli/schema/store/v2"
	"github.com/spf13/cobra"
)

func init() {
	storeCmd.AddCommand(storeDescribeCmd)
}

var storeDescribeCmd = &cobra.Command{
	Use:   `describe (FUNCTION_NAME|FUNCTION_TITLE) [--url STORE_URL]`,
	Short: "Show details of OpenFaaS function from a store",
	Example: `  faas-cli store describe nodeinfo
  faas-cli store describe nodeinfo --url https://host:port/store.json
`,
	Aliases: []string{"inspect"},
	RunE:    runStoreDescribe,
}

func runStoreDescribe(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	targetPlatform := getTargetPlatform(platformValue)
	storeItems, err := proxy.FunctionStoreList(storeAddress)
	if err != nil {
		return err
	}

	platformFunctions := filterStoreList(storeItems, targetPlatform)

	functionName := args[0]
	item := storeFindFunction(functionName, platformFunctions)
	if item == nil {
		return fmt.Errorf("function '%s' not found for platform '%s'", functionName, targetPlatform)
	}

	content := storeRenderItem(item, targetPlatform)
	fmt.Print(content)

	return nil
}

func storeRenderItem(item *storeV2.StoreFunction, platform string) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)

	author := item.Author
	if author == "" {
		author = "unknown"
	}

	fmt.Fprintf(w, "%s\t%s\n", "Title:", item.Title)
	fmt.Fprintf(w, "%s\t%s\n", "Author:", item.Author)
	fmt.Fprintf(w, "%s\t\n%s\n\n", "Description:", wordwrap.WrapString(item.Description, 80))

	fmt.Fprintf(w, "%s\t%s\n", "Image:", item.GetImageName(platform))
	fmt.Fprintf(w, "%s\t%s\n", "Process:", item.Fprocess)
	fmt.Fprintf(w, "%s\t%s\n", "Repo URL:", item.RepoURL)
	if len(item.Environment) > 0 {
		fmt.Fprintf(w, "Environment:\n")
		for k, v := range item.Environment {
			fmt.Fprintf(w, "- \t%s:\t%s\n", k, v)
		}
		fmt.Fprintln(w)
	}

	if item.Labels != nil {
		fmt.Fprintf(w, "Labels:\n")
		for k, v := range item.Labels {
			fmt.Fprintf(w, "- \t%s:\t%s\n", k, v)
		}
		fmt.Fprintln(w)
	}

	if len(item.Annotations) > 0 {
		fmt.Fprintf(w, "Annotations:\n")
		for k, v := range item.Annotations {
			fmt.Fprintf(w, "- \t%s:\t%s\n", k, v)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
	return b.String()
}
