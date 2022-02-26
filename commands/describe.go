// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/schema"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-provider/types"

	"github.com/spf13/cobra"
)

func init() {
	describeCmd.Flags().StringVar(&functionName, "name", "", "Name of the function")
	describeCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	describeCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	describeCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")
	describeCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
	describeCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", "", "Namespace of the function")

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
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
			yamlGateway = services.Provider.GatewayURL
		}
	}
	gatewayAddress := getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))
	cliAuth, err := proxy.NewCLIAuth(token, gatewayAddress)
	if err != nil {
		return err
	}
	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)
	cliClient, err := proxy.NewClient(cliAuth, gatewayAddress, transport, &commandTimeout)
	if err != nil {
		return err
	}

	ctx := context.Background()

	function, err := cliClient.GetFunctionInfo(ctx, functionName, functionNamespace)
	if err != nil {
		return err
	}

	//To get correct value for invocation count from /system/functions endpoint
	functionList, err := cliClient.ListFunctions(ctx, functionNamespace)
	if err != nil {
		return err
	}

	var invocationCount int
	for _, fn := range functionList {
		if fn.Name == function.Name {
			invocationCount = int(fn.InvocationCount)
			break
		}
	}

	var status = "Not Ready"
	if function.AvailableReplicas > 0 {
		status = "Ready"
	}

	url, asyncURL := getFunctionURLs(gatewayAddress, functionName, functionNamespace)

	funcDesc := schema.FunctionDescription{
		FunctionStatus:  function,
		Status:          status,
		InvocationCount: int(invocationCount),
		URL:             url,
		AsyncURL:        asyncURL,
	}

	printFunctionDescription(funcDesc)

	return nil
}

func getFunctionURLs(gateway string, functionName string, functionNamespace string) (string, string) {
	gateway = strings.TrimRight(gateway, "/")

	url := gateway + "/function/" + functionName
	asyncURL := gateway + "/async-function/" + functionName

	if functionNamespace != "" {
		url += "." + functionNamespace
		asyncURL += "." + functionNamespace
	}

	return url, asyncURL
}

func printFunctionDescription(funcDesc schema.FunctionDescription) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	process := "<default>"
	if funcDesc.EnvProcess != "" {
		process = funcDesc.EnvProcess
	}
	fmt.Fprintln(w, "Name:\t "+funcDesc.Name)
	fmt.Fprintln(w, "Status:\t "+funcDesc.Status)
	fmt.Fprintln(w, "Replicas:\t "+strconv.Itoa(int(funcDesc.Replicas)))
	fmt.Fprintln(w, "Available replicas:\t "+strconv.Itoa(int(funcDesc.AvailableReplicas)))
	fmt.Fprintln(w, "Invocations:\t "+strconv.Itoa(funcDesc.InvocationCount))
	fmt.Fprintln(w, "Image:\t "+funcDesc.Image)
	fmt.Fprintln(w, "Function process:\t "+process)
	fmt.Fprintln(w, "URL:\t "+funcDesc.URL)
	fmt.Fprintln(w, "Async URL:\t "+funcDesc.AsyncURL)
	printMap(w, "Labels", *funcDesc.Labels)
	printMap(w, "Annotations", *funcDesc.Annotations)
	printList(w, "Constraints", funcDesc.Constraints)
	printMap(w, "Environment", funcDesc.EnvVars)
	printList(w, "Secrets", funcDesc.Secrets)
	printResources(w, "Requests", funcDesc.Requests)
	printResources(w, "Limits", funcDesc.Limits)
	if funcDesc.Usage != nil {
		fmt.Println()

		fmt.Fprintf(w, "RAM:\t %.2f MB\n", (funcDesc.Usage.TotalMemoryBytes / 1024 / 1024))
		cpu := funcDesc.Usage.CPU
		if cpu < 0 {
			cpu = 1
		}
		fmt.Fprintf(w, "CPU:\t %.0f Mi\n", (cpu))
	}

	w.Flush()
}

func printMap(w *tabwriter.Writer, name string, m map[string]string) {
	fmt.Fprintf(w, name+":")

	if len(m) == 0 {
		fmt.Fprintln(w, " \t <none>")
		return
	}

	for key, value := range m {
		fmt.Fprintln(w, " \t "+key+": "+value)
	}

	return
}

func printList(w *tabwriter.Writer, name string, data []string) {
	fmt.Fprintf(w, name+":")

	if len(data) == 0 {
		fmt.Fprintln(w, " \t <none>")
		return
	}

	for _, value := range data {
		fmt.Fprintln(w, " \t - "+value)
	}

	return
}

func printResources(w *tabwriter.Writer, name string, data *types.FunctionResources) {
	fmt.Fprintf(w, name+":")

	if data == nil {
		fmt.Fprintln(w, " \t <none>")
		return
	}

	fmt.Fprintln(w, " \t CPU: "+data.CPU)
	fmt.Fprintln(w, " \t Memory: "+data.Memory)

	return
}
