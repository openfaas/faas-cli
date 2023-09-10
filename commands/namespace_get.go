// Copyright (c) OpenFaaS Author(s) 2023. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/openfaas/faas-provider/types"
	"github.com/spf13/cobra"
)

var namespaceGetCmd = &cobra.Command{
	Use:     `get NAMESPACE_NAME`,
	Short:   "Get existing namespace",
	Long:    "Get existing namespace",
	Example: `faas-cli namespace get NAME`,
	RunE:    get_namespace,
	PreRunE: preGetNamespace,
}

func init() {
	namespaceCmd.AddCommand(namespaceGetCmd)
}

func preGetNamespace(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("namespace name required")
	}

	if len(args) > 1 {
		return fmt.Errorf("too many values for namespace name")
	}

	return nil
}

func get_namespace(cmd *cobra.Command, args []string) error {
	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	ns := args[0]

	res, err := client.GetNamespace(context.Background(), ns)
	if err != nil {
		return err
	}

	printNamespaceDetail(cmd.OutOrStdout(), res)

	return nil
}

func printNamespaceDetail(dst io.Writer, nsDetail types.FunctionNamespace) {
	w := tabwriter.NewWriter(dst, 0, 0, 1, ' ', tabwriter.TabIndent)
	defer w.Flush()

	out := printer{
		w:       w,
		verbose: verbose,
	}
	out.Printf("Name:\t%s\n", nsDetail.Name)
	if len(nsDetail.Labels) > 1 {
		out.Printf("Labels", nsDetail.Labels)
	} else {
		out.Printf("Labels", map[string]string{})
	}
	if len(nsDetail.Annotations) > 1 {
		out.Printf("Annotations", nsDetail.Annotations)
	} else {
		out.Printf("Annotations", map[string]string{})
	}
}
