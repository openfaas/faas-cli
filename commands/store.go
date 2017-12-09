// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"text/tabwriter"

	"github.com/spf13/cobra"
)

var storeAddress string

const defaultStore = "https://cdn.rawgit.com/openfaas/store/master/store.json"

type storeItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Name        string `json:"name"`
	Fprocess    string `json:"fprocess"`
	Network     string `json:"network"`
	RepoURL     string `json:"repo_url"`
}

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	storeListCmd.Flags().StringVarP(&storeAddress, "store", "g", defaultStore, "Store URL starting with http(s)://")
	storeInspectCmd.Flags().StringVarP(&storeAddress, "store", "g", defaultStore, "Store URL starting with http(s)://")

	storeCmd.AddCommand(storeListCmd)
	storeCmd.AddCommand(storeInspectCmd)
	faasCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   `store [--store STORE_URL]`,
	Short: "OpenFaaS store commands",
}

var storeListCmd = &cobra.Command{
	Use:     `list`,
	Short:   "List OpenFaaS store items",
	Long:    `Lists the available items in OpenFaas store`,
	Example: `  faas-cli store list --store https://domain:port`,
	RunE:    runStoreList,
}

var storeInspectCmd = &cobra.Command{
	Use:     `inspect FUNCTION_NAME`,
	Short:   "Show OpenFaaS store function details",
	Long:    `Prints the detailed informations of the specified OpenFaaS function`,
	Example: `  faas-cli store inspect NodeInfo --store https://domain:port`,
	RunE:    runStoreInspect,
}

func runStoreList(cmd *cobra.Command, args []string) error {
	items, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("The store is empty.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION")
	fmt.Fprintln(w, "--------\t-----------")

	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\n", item.Title, item.Description)
	}

	fmt.Fprintln(w)
	w.Flush()

	return nil
}

func runStoreInspect(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	functionName := args[0]

	items, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Title == functionName {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w)
			fmt.Fprintln(w, "FUNCTION\tDESCRIPTION\tIMAGE\tFUNCTION PROCESS\tREPO")
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				item.Title,
				item.Description,
				item.Image,
				item.Fprocess,
				item.RepoURL,
			)

			fmt.Fprintln(w)
			w.Flush()
		}
	}

	return fmt.Errorf("function not found")
}

func storeList(store string) ([]storeItem, error) {
	var results []storeItem

	store = strings.TrimRight(store, "/")

	timeout := 60 * time.Second
	client := makeHTTPClient(&timeout)

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

func makeHTTPClient(timeout *time.Duration) http.Client {
	if timeout != nil {
		return http.Client{
			Timeout: *timeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout: *timeout,
					// KeepAlive: 0,
				}).DialContext,
				// MaxIdleConns:          1,
				// DisableKeepAlives:     true,
				IdleConnTimeout:       120 * time.Millisecond,
				ExpectContinueTimeout: 1500 * time.Millisecond,
			},
		}
	}

	// This should be used for faas-cli invoke etc.
	return http.Client{}
}
