// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"

	"github.com/spf13/cobra"
)

func init() {
	describeCmd.Flags().StringVar(&functionName, "name", "", "Name of the function")
	describeCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	describeCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	faasCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe FUNCTION_NAME [--gateway GATEWAY_URL]",
	Short: "Describe an OpenFaaS function",
	Long:  `Display details of an OpenFaaS function`,
	Example: `faas-cli describe figlet 
faas-cli describe env --gateway http://127.0.0.1:8080
faas-cli describe echo -g http://127.0.0.1.8080`,
	PreRunE: preRunDescribe,
	RunE:    runDescribe,
}

func preRunDescribe(cmd *cobra.Command, args []string) error {
	return nil
}

func runDescribe(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}
	var yamlGateway string
	var services stack.Services
	functionName = args[0]

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
	gatewayAddress := getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))

	function, err := proxy.GetFunctionInfo(gatewayAddress, functionName, tlsInsecure)
	if err != nil {
		return err
	}

	var status string
	if function.AvailableReplicas > 0 {
		status = "Ready"
	} else {
		status = "Not Ready"
	}

	fmt.Printf("%s:\t\t\t%s\n", "Name", function.Name)
	fmt.Printf("%s:\t\t\t%s\n", "Status", status)
	fmt.Printf("%s:\t\t%d\n", "Replicas", function.Replicas)
	fmt.Printf("%s:\t%d\n", "Available replicas", function.AvailableReplicas)
	fmt.Printf("%s:\t\t%v\n", "Invocations", function.InvocationCount)
	fmt.Printf("%s:\t\t\t%s\n", "Image", function.Image)
	fmt.Printf("%s:\t%s\n", "Function process", function.EnvProcess)
	fmt.Printf("%s:\t\t\t%s\n", "URL", getFunctionURL(gatewayAddress, functionName))
	fmt.Printf("%s:\t\t%s\n", "Async URL", getFunctionAsyncURL(gatewayAddress, functionName))

	return nil
}

func getFunctionURL(gateway string, functionName string) string {
	gateway = strings.TrimRight(gateway, "/")
	return gateway + "/function/" + functionName
}

func getFunctionAsyncURL(gateway string, functionName string) string {
	gateway = strings.TrimRight(gateway, "/")
	return gateway + "/async-function/" + functionName
}
