// Copyright (c) OpenFaaS Author(s) 2023. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var namespaceDeleteCmd = &cobra.Command{
	Use:     `delete NAME`,
	Short:   "Delete existing namespace",
	Long:    "Delete existing namespace",
	Example: `  faas-cli namespace delete NAME`,
	RunE:    deleteNamespace,
	PreRunE: preDeleteNamespace,
}

func init() {
	namespaceCmd.AddCommand(namespaceDeleteCmd)
}

func preDeleteNamespace(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("namespace name required")
	}

	if len(args) > 1 {
		return fmt.Errorf("too many values for namespace name")
	}

	return nil
}

func deleteNamespace(cmd *cobra.Command, args []string) error {
	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	ns := args[0]

	fmt.Printf("Deleting Namespace: %s\n", ns)
	if err = client.DeleteNamespace(context.Background(), ns); err != nil {
		return err
	}

	fmt.Printf("Namespace Deleted: %s\n", ns)

	return nil
}
