// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	verboseList bool
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	listCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")

	listCmd.Flags().BoolVarP(&verboseList, "verbose", "v", false, "Verbose output for the function list")

	faasCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     `list [--gateway GATEWAY_URL] [--verbose]`,
	Aliases: []string{"ls"},
	Short:   "List OpenFaaS functions",
	Long:    `Lists OpenFaaS functions either on a local or remote gateway`,
	Example: `  faas-cli list
  faas-cli list --gateway https://localhost:8080 --verbose`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	var services stack.Services
	var gatewayAddress string
	var yamlGateway string
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress = getGatewayURL(gateway, defaultGateway, yamlGateway)

	functions, err := proxy.ListFunctions(gatewayAddress)
	if err != nil {
		return err
	}

	if verboseList {
		fmt.Printf("%-30s\t%-40s\t%-15s\t%-5s\n", "Function", "Image", "Invocations", "Replicas")
		for _, function := range functions {
			functionImage := function.Image
			if len(function.Image) > 40 {
				functionImage = functionImage[0:38] + ".."
			}
			fmt.Printf("%-30s\t%-40s\t%-15d\t%-5d\n", function.Name, functionImage, int64(function.InvocationCount), function.Replicas)
		}
	} else {
		fmt.Printf("%-30s\t%-15s\t%-5s\n", "Function", "Invocations", "Replicas")
		for _, function := range functions {
			fmt.Printf("%-30s\t%-15d\t%-5d\n", function.Name, int64(function.InvocationCount), function.Replicas)

		}
	}
	return nil
}
