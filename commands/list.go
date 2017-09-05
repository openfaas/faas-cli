// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"

	"github.com/alexellis/faas-cli/proxy"
	"github.com/alexellis/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	verboseList bool
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	listCmd.Flags().StringVar(&fprocess, "fprocess", "", "Fprocess to be run by the watchdog")
	listCmd.Flags().StringVar(&gateway, "gateway", "http://localhost:8080", "Gateway URI")
	listCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	listCmd.Flags().StringVar(&image, "image", "", "Docker image name to build")
	listCmd.Flags().StringVar(&language, "lang", "node", "Programming language template")
	listCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")

	listCmd.Flags().BoolVar(&verboseList, "verbose", false, "Verbose output for the function list")

	faasCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use: `list [--gateway GATEWAY_URL] [--verbose]`,

	Short: "List OpenFaaS functions",
	Long:  `Lists OpenFaaS functions either on a local or remote gateway`,
	Example: `  faas-cli list
  faas-cli list --gateway https://localhost:8080 --verbose`,
	Run: runList,
}

func runList(cmd *cobra.Command, args []string) {
	var services stack.Services
	var gatewayAddress string

	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAML(yamlFile)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		if parsedServices != nil {
			services = *parsedServices
			gatewayAddress = services.Provider.GatewayURL
		}
	}
	if len(gatewayAddress) == 0 {
		gatewayAddress = gateway
	}

	// fmt.Println(gatewayAddress)
	functions, err := proxy.ListFunctions(gatewayAddress)
	if err != nil {
		log.Println(err)
		return
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

}
