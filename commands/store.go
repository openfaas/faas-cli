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
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	storeCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	storeCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")

	// Setup flags used by store command
	storeListCmd.Flags().StringVarP(&storeAddress, "store", "g", defaultStore, "Store URL starting with http(s)://")
	storeInspectCmd.Flags().StringVarP(&storeAddress, "store", "g", defaultStore, "Store URL starting with http(s)://")
	storeListCmd.Flags().BoolVarP(&verboseDescription, "verbose", "v", false, "Verbose output for the field values")
	storeInspectCmd.Flags().BoolVarP(&verboseDescription, "verbose", "v", false, "Verbose output for the field values")

	// Setup flags that are used only by deploy command (variables defined above)
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.envvarOpts, "env", "e", []string{}, "Adds one or more environment variables to the defined ones by store (ENVVAR=VALUE)")
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.labelOpts, "label", "l", []string{}, "Set one or more label (LABEL=VALUE)")
	storeDeployCmd.Flags().BoolVar(&storeDeployFlags.replace, "replace", false, "Replace any existing function")
	storeDeployCmd.Flags().BoolVar(&storeDeployFlags.update, "update", true, "Update existing functions")
	storeDeployCmd.Flags().StringArrayVar(&storeDeployFlags.constraints, "constraint", []string{}, "Apply a constraint to the function")
	storeDeployCmd.Flags().StringArrayVar(&storeDeployFlags.secrets, "secret", []string{}, "Give the function access to a secure secret")

	// Set bash-completion.
	_ = storeDeployCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	storeCmd.AddCommand(storeListCmd)
	storeCmd.AddCommand(storeInspectCmd)
	storeCmd.AddCommand(storeDeployCmd)
	faasCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   `store`,
	Short: "OpenFaaS store commands",
	Long:  "Allows browsing and deploying OpenFaaS store functions",
}

var storeListCmd = &cobra.Command{
	Use:     `list [--store STORE_URL]`,
	Short:   "List OpenFaaS store items",
	Long:    "Lists the available items in OpenFaas store",
	Example: `  faas-cli store list --store https://domain:port`,
	RunE:    runStoreList,
}

var storeInspectCmd = &cobra.Command{
	Use:     `inspect (FUNCTION_NAME|FUNCTION_TITLE) [--store STORE_URL]`,
	Short:   "Show OpenFaaS store function details",
	Long:    "Prints the detailed informations of the specified OpenFaaS function",
	Example: `  faas-cli store inspect NodeInfo --store https://domain:port`,
	RunE:    runStoreInspect,
}

var storeDeployCmd = &cobra.Command{
	Use: `deploy (FUNCTION_NAME|FUNCTION_TITLE)
							[--gateway GATEWAY_URL]
							[--handler HANDLER_DIR]
							[--env ENVVAR=VALUE ...]
							[--label LABEL=VALUE ...]
							[--replace=false]
							[--update=true]
							[--constraint PLACEMENT_CONSTRAINT ...]
							[--regex "REGEX"]
							[--filter "WILDCARD"]
							[--secret "SECRET_NAME"]`,

	Short: "Deploy OpenFaaS functions from the store",
	Long:  `Same as faas-cli deploy except pre-loaded with arguments from the store`,
	Example: `  faas-cli store deploy figlet
							faas-cli store deploy figlet
									--gateway=http://remote-site.com:8080 --lang=python
									--env=MYVAR=myval`,
	RunE: runStoreDeploy,
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

	content := renderStoreItems(items)
	fmt.Print(content)

	return nil
}

func renderStoreItems(items []schema.StoreItem) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION")

	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\n", item.Title, renderDescription(item.Description))
	}

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}

func renderDescription(descr string) string {
	if !verboseDescription && len(descr) > maxDescriptionLen {
		return descr[0:maxDescriptionLen-3] + "..."
	}

	return descr
}

func runStoreInspect(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	storeItems, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	item := findFunction(args[0], storeItems)
	if item == nil {
		return fmt.Errorf("function '%s' not found", functionName)
	}

	content := renderStoreItem(item)
	fmt.Print(content)

	return nil
}

func renderStoreItem(item *schema.StoreItem) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FUNCTION\tDESCRIPTION\tIMAGE\tPROCESS\tREPO")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		item.Title,
		renderDescription(item.Description),
		item.Image,
		item.Fprocess,
		item.RepoURL,
	)

	fmt.Fprintln(w)
	w.Flush()
	return b.String()
}

func runStoreDeploy(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	storeItems, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	item := findFunction(args[0], storeItems)
	if item == nil {
		return fmt.Errorf("function '%s' not found", functionName)
	}

	// Add the store environement variables to the provided ones from cmd
	if item.Environment != nil {
		for _, env := range item.Environment {
			storeDeployFlags.envvarOpts = append(storeDeployFlags.envvarOpts, env)
		}
	}

	return RunDeploy(
		args,
		item.Image,
		item.Fprocess,
		item.Name,
		storeDeployFlags,
	)
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

func findFunction(functionName string, storeItems []schema.StoreItem) *schema.StoreItem {
	var item schema.StoreItem

	for _, item = range storeItems {
		if item.Name == functionName || item.Title == functionName {
			return &item
		}
	}

	return &item
}
