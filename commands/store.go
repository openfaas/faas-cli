// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openfaas/faas-cli/platform"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

var (
	storeAddress     string
	verbose          bool
	storeDeployFlags DeployFlags
)

const (
	defaultStore      = "https://cdn.rawgit.com/openfaas/store/master/functions.json"
	maxDescriptionLen = 40
)

var platformValue string

func init() {
	storeCmd.PersistentFlags().StringVarP(&storeAddress, "url", "u", defaultStore, "Alternative Store URL starting with http(s)://")
	storeCmd.PersistentFlags().StringVarP(&platformValue, "platform", "p", "", "Target platform for store")

	faasCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   `store`,
	Short: "OpenFaaS store commands",
	Long:  "Allows browsing and deploying OpenFaaS functions from a store",
}

func storeList(store, platform string) ([]schema.StoreFunction, error) {

	var storeData schema.StoreV2

	store = strings.TrimRight(store, "/")

	timeout := 60 * time.Second
	tlsInsecure := false

	client := proxy.MakeHTTPClient(&timeout, tlsInsecure)

	res, err := client.Get(store)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS store at URL: %s", store)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS store at URL: %s", store)
		}

		jsonErr := json.Unmarshal(bytesOut, &storeData)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS store at URL: %s\n%s", store, jsonErr.Error())
		}
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}

	return storeData.Functions, nil
}

func filterStoreList(functions []schema.StoreFunction, platform string) []schema.StoreFunction {
	var filteredList []schema.StoreFunction

	for _, function := range functions {
		_, ok := function.Images[platform]

		if ok {
			filteredList = append(filteredList, function)
		}
	}

	return filteredList
}

func storeFindFunction(functionName string, storeItems []schema.StoreFunction) *schema.StoreFunction {
	var item schema.StoreFunction

	for _, item = range storeItems {
		if item.Name == functionName || item.Title == functionName {
			return &item
		}
	}

	return nil
}

func getTargetPlatform(inputPlatform string) string {
	if len(inputPlatform) == 0 {
		return platform.GetPlatform()
	}

	return inputPlatform
}

func getStorePlatforms(functions []schema.StoreFunction) []string {
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
