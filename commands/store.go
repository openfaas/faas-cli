// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

var (
	storeAddress     string
	verbose          bool
	storeDeployFlags DeployFlags
)

const (
	defaultStore      = "https://cdn.rawgit.com/openfaas/store/master/store.json"
	maxDescriptionLen = 40
)

func init() {
	storeCmd.PersistentFlags().StringVarP(&storeAddress, "url", "u", defaultStore, "Alternative Store URL starting with http(s)://")

	faasCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   `store`,
	Short: "OpenFaaS store commands",
	Long:  "Allows browsing and deploying OpenFaaS functions from a store",
}

func storeFindFunction(functionName string, storeItems []schema.StoreItem) *schema.StoreItem {
	var item schema.StoreItem

	for _, item = range storeItems {
		if item.Name == functionName || item.Title == functionName {
			return &item
		}
	}

	return nil
}
