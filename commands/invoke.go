// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	contentType string
	query       []string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	invokeCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	invokeCmd.Flags().StringVar(&gateway, "gateway", defaultGateway, "Gateway URI")

	invokeCmd.Flags().StringVar(&contentType, "content-type", "text/plain", "The content-type HTTP header such as application/json")
	invokeCmd.Flags().StringArrayVar(&query, "query", []string{}, "pass query-string options")

	faasCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   `invoke FUNCTION_NAME [--gateway GATEWAY_URL] [--content-type CONTENT_TYPE] [--query PARAM=VALUE]`,
	Short: "Invoke an OpenFaaS function",
	Long:  `Invokes an OpenFaaS function and reads from STDIN for the body of the request`,
	Example: `  faas-cli invoke echo --gateway https://domain:port
  faas-cli invoke echo --gateway https://domain:port --content-type application/json
  faas-cli invoke env --query repo=faas-cli --query org=openfaas`,
	Run: runInvoke,
}

func runInvoke(cmd *cobra.Command, args []string) {
	var services stack.Services

	if len(args) < 1 {
		fmt.Println("Please provide a name for the function")
		return
	}
	var yamlGateway string
	functionName = args[0]

	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFileForStack(yamlFile, regex, filter)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		if parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}

	gatewayAddress := getGatewayURL(gateway, defaultGateway, yamlGateway)

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
	}

	functionInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Unable to read standard input: %s\n", err.Error())
		return
	}

	response, err := proxy.InvokeFunction(gatewayAddress, functionName, &functionInput, contentType, query)
	if err != nil {
		fmt.Println(err)
		return
	}

	if response != nil {
		os.Stdout.Write(*response)
	}
}
