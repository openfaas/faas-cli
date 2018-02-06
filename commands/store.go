// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"text/tabwriter"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	"github.com/spf13/cobra"
)

var (
	storeAddress       string
	verboseDescription bool
	storeDeployFlags   DeployFlags
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
	Long:  "Allows browsing and deploying OpenFaaS store functions",
}

func storeRenderItems(items []schema.StoreItem) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION")

	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\n", item.Title, storeRenderDescription(item.Description))
	}

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}

func storeRenderDescription(descr string) string {
	if !verboseDescription && len(descr) > maxDescriptionLen {
		return descr[0:maxDescriptionLen-3] + "..."
	}

	return descr
}

func storeRenderItem(item *schema.StoreItem) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION\tIMAGE\tPROCESS\tREPO")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		item.Title,
		storeRenderDescription(item.Description),
		item.Image,
		item.Fprocess,
		item.RepoURL,
	)

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}

func storeList(store string) ([]schema.StoreItem, error) {
	var results []schema.StoreItem

	store = strings.TrimRight(store, "/")

	timeout := 60 * time.Second
	client := proxy.MakeHTTPClient(&timeout)

	getRequest, err := http.NewRequest(http.MethodGet, store, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS store on URL: %s", store)
	}

	res, err := client.Do(getRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to OpenFaaS store on URL: %s", store)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read result from OpenFaaS store on URL: %s", store)
		}
		jsonErr := json.Unmarshal(bytesOut, &results)
		if jsonErr != nil {
			return nil, fmt.Errorf("cannot parse result from OpenFaaS store on URL: %s\n%s", store, jsonErr.Error())
		}
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return nil, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return results, nil
}

func storeFindFunction(functionName string, storeItems []schema.StoreItem) *schema.StoreItem {
	var item schema.StoreItem

	for _, item = range storeItems {
		if item.Name == functionName || item.Title == functionName {
			return &item
		}
	}

	return &item
}
