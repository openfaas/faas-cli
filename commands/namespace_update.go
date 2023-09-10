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

// NamespaceUpdateFlags holds flags that are to be added to commands.
type NamespaceUpdateFlags struct {
	labelOpts      []string
	annotationOpts []string
}

var namespaceUpdateFlags NamespaceCreateFlags

var namespaceUpdateCmd = &cobra.Command{
	Use: `update NAME
			[--label LABEL=VALUE ...]
			[--annotation ANNOTATION=VALUE ...]`,
	Short: "Update a namespace",
	Long:  "Update a namespace",
	Example: `faas-cli namespace update NAME
	faas-cli namespace update NAME --label demo=true
	faas-cli namespace update NAME --annotation demo=true
	faas-cli namespace update NAME --label demo=true \
	  --annotation demo=true`,
	RunE:    updateNamespace,
	PreRunE: preUpdateNamespace,
}

func init() {
	namespaceUpdateCmd.Flags().StringArrayVarP(&namespaceUpdateFlags.labelOpts, "label", "l", []string{}, "Set one or more label (LABEL=VALUE)")
	namespaceUpdateCmd.Flags().StringArrayVarP(&namespaceUpdateFlags.annotationOpts, "annotation", "", []string{}, "Set one or more annotation (ANNOTATION=VALUE)")

	namespaceCmd.AddCommand(namespaceUpdateCmd)
}

func preUpdateNamespace(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("namespace name required")
	}

	if len(args) > 1 {
		return fmt.Errorf("too many values for namespace name")
	}

	return nil
}

func updateNamespace(cmd *cobra.Command, args []string) error {
	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	labels, err := util.ParseMap(namespaceUpdateFlags.labelOpts, "labels")
	if err != nil {
		return err
	}

	annotations, err := util.ParseMap(namespaceUpdateFlags.annotationOpts, "annotations")
	if err != nil {
		return err
	}

	req := types.FunctionNamespace{
		Name:        args[0],
		Labels:      labels,
		Annotations: annotations,
	}

	fmt.Printf("Updating Namespace: %s\n", req.Name)
	if _, err = client.UpdateNamespace(context.Background(), req); err != nil {
		return err
	}

	fmt.Printf("Namespace Updated: %s\n", req.Name)

	return nil
}
