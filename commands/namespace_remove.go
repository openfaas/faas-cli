// Copyright (c) OpenFaaS Author(s) 2023. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var namespaceRemoveCmd = &cobra.Command{
	Use:     `remove NAME`,
	Short:   "Remove existing namespace",
	Long:    "Remove existing namespace",
	Example: `  faas-cli namespace remove NAME`,
	Aliases: []string{"rm", "delete"},
	RunE:    removeNamespace,
	PreRunE: preRemoveNamespace,
}

func init() {
	namespaceCmd.AddCommand(namespaceRemoveCmd)
}

func preRemoveNamespace(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("namespace name required")
	}

	if len(args) > 1 {
		return fmt.Errorf("too many values for namespace name")
	}

	return nil
}

func removeNamespace(cmd *cobra.Command, args []string) error {
	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	ns := args[0]

	fmt.Printf("Deleting Namespace: %s\n", ns)
	if err = client.DeleteNamespace(context.Background(), ns); err != nil {
		return err
	}

	fmt.Printf("Namespace Removed: %s\n", ns)

	return nil
}
