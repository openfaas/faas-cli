// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"strings"

	storeV2 "github.com/openfaas/faas-cli/schema/store/v2"
	"github.com/spf13/cobra"
)

var (
	storeAddress     string
	verbose          bool
	storeDeployFlags DeployFlags
	//Platform platform variable set at build time
	Platform string
	// if the CLI is built using buildx, then the Platform value needs to be mapped to
	// one of the supported values used in the store.
	shortPlatform = map[string]string{
		"linux/arm/v6": "armhf",
		"linux/amd64":  "x86_64",
		"linux/arm64":  "arm64",
	}
)

const (
	defaultStore      = "https://raw.githubusercontent.com/openfaas/store/master/functions.json"
	maxDescriptionLen = 40
	storeAddressDoc   = `Alternative path to Function Store metadata. It may be an http(s) URL or a local path to a JSON file.`
)

var platformValue string

func init() {
	storeCmd.PersistentFlags().StringVarP(&storeAddress, "url", "u", defaultStore, storeAddressDoc)
	storeCmd.PersistentFlags().StringVarP(&platformValue, "platform", "p", Platform, "Target platform for store")

	faasCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   `store`,
	Short: "OpenFaaS store commands",
	Long:  "Allows browsing and deploying OpenFaaS functions from a store",
}

func filterStoreList(functions []storeV2.StoreFunction, platform string) []storeV2.StoreFunction {
	var filteredList []storeV2.StoreFunction

	for _, function := range functions {

		_, ok := getValueIgnoreCase(function.Images, platform)

		if ok {
			filteredList = append(filteredList, function)
		}
	}

	return filteredList
}

// getValueIgnoreCase get a key value from map by ignoring case for key
func getValueIgnoreCase(kv map[string]string, key string) (string, bool) {
	for k, v := range kv {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return "", false
}

func storeFindFunction(functionName string, storeItems []storeV2.StoreFunction) *storeV2.StoreFunction {
	var item storeV2.StoreFunction

	for _, item = range storeItems {
		if item.Name == functionName || item.Title == functionName {
			return &item
		}
	}

	return nil
}

func getPlatform() string {
	if len(Platform) == 0 {
		return mainPlatform
	}
	return Platform
}

func getTargetPlatform(inputPlatform string) string {
	if len(inputPlatform) == 0 {
		currentPlatform := getPlatform()
		target, ok := shortPlatform[currentPlatform]
		if ok {
			return target
		}
		return currentPlatform
	}
	return inputPlatform
}

func getStorePlatforms(functions []storeV2.StoreFunction) []string {
	var distinctPlatformMap = make(map[string]bool)
	var result []string

	for _, function := range functions {
		for key := range function.Images {
			_, exists := distinctPlatformMap[key]

			if !exists {
				distinctPlatformMap[key] = true
				result = append(result, key)
			}
		}
	}

	return result
}
