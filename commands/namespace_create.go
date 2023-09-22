// Copyright (c) OpenFaaS Author(s) 2023. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"

	"github.com/openfaas/faas-cli/util"
	"github.com/openfaas/faas-provider/types"
	"github.com/spf13/cobra"
)

// NamespaceCreateFlags holds flags that are to be added to commands.
type NamespaceCreateFlags struct {
	labelOpts      []string
	annotationOpts []string
}

var namespaceCreateFlags NamespaceCreateFlags

var namespaceCreateCmd = &cobra.Command{
	Use: `create NAME
			[--label LABEL=VALUE ...]
			[--annotation ANNOTATION=VALUE ...]`,
	Short: "Create a new namespace",
	Long:  "Create command creates a new namespace",
	Example: `  faas-cli namespace create NAME
  faas-cli namespace create NAME --label demo=true
  faas-cli namespace create NAME --annotation demo=true
  faas-cli namespace create NAME --label demo=true \
    --annotation demo=true`,
	RunE:    createNamespace,
	PreRunE: preCreateNamespace,
}

func init() {
	namespaceCreateCmd.Flags().StringArrayVarP(&namespaceCreateFlags.labelOpts, "label", "l", []string{}, "Set one or more label (LABEL=VALUE)")
	namespaceCreateCmd.Flags().StringArrayVarP(&namespaceCreateFlags.annotationOpts, "annotation", "", []string{}, "Set one or more annotation (ANNOTATION=VALUE)")

	namespaceCmd.AddCommand(namespaceCreateCmd)
}

func preCreateNamespace(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("namespace name required")
	}

	if len(args) > 1 {
		return fmt.Errorf("too many values for namespace name")
	}

	return nil
}

func createNamespace(cmd *cobra.Command, args []string) error {
	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	labels, err := util.ParseMap(namespaceCreateFlags.labelOpts, "labels")
	if err != nil {
		return err
	}

	annotations, err := util.ParseMap(namespaceCreateFlags.annotationOpts, "annotations")
	if err != nil {
		return err
	}

	req := types.FunctionNamespace{
		Name:        args[0],
		Labels:      labels,
		Annotations: annotations,
	}

	fmt.Printf("Creating Namespace: %s\n", req.Name)
	if _, err = client.CreateNamespace(context.Background(), req); err != nil {
		return err
	}

	fmt.Printf("Namespace Created: %s\n", req.Name)

	return nil
}
