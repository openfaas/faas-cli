// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
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
	describeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

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

	printFunctionDescription(cmd.OutOrStdout(), funcDesc, verbose)

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

func printFunctionDescription(dst io.Writer, funcDesc schema.FunctionDescription, verbose bool) {
	w := tabwriter.NewWriter(dst, 0, 0, 1, ' ', tabwriter.TabIndent)
	defer w.Flush()

	out := printer{
		w:       w,
		verbose: verbose,
	}

	process := "<default>"
	if funcDesc.EnvProcess != "" {
		process = funcDesc.EnvProcess
	}

	out.Printf("Name:\t%s\n", funcDesc.Name)
	out.Printf("Status:\t%s\n", funcDesc.Status)
	out.Printf("Replicas:\t%s\n", strconv.Itoa(int(funcDesc.Replicas)))
	out.Printf("Available Replicas:\t%s\n", strconv.Itoa(int(funcDesc.AvailableReplicas)))
	out.Printf("Invocations:\t%s\n", strconv.Itoa(int(funcDesc.InvocationCount)))
	out.Printf("Image:\t%s\n", funcDesc.Image)
	out.Printf("Function Process:\t%s\n", process)
	out.Printf("URL:\t%s\n", funcDesc.URL)
	out.Printf("Async URL:\t%s\n", funcDesc.AsyncURL)
	out.Printf("Labels", *funcDesc.Labels)
	out.Printf("Annotations", *funcDesc.Annotations)
	out.Printf("Constraints", funcDesc.Constraints)
	out.Printf("Environment", funcDesc.EnvVars)
	out.Printf("Secrets", funcDesc.Secrets)
	out.Printf("Requests", funcDesc.Requests)
	out.Printf("Limits", funcDesc.Limits)
	out.Printf("", funcDesc.Usage)
}

type printer struct {
	verbose bool
	w       io.Writer
}

func (p *printer) Printf(format string, a interface{}) {
	switch v := a.(type) {
	case map[string]string:
		printMap(p.w, format, v, p.verbose)
	case []string:
		printList(p.w, format, v, p.verbose)
	case *types.FunctionResources:
		printResources(p.w, format, v, p.verbose)
	case *types.FunctionUsage:
		printUsage(p.w, v, p.verbose)
	default:
		if !p.verbose && isEmpty(a) {
			return
		}

		if p.verbose && isEmpty(a) {
			a = "<none>"
		}

		fmt.Fprintf(p.w, format, a)
	}

}

func printUsage(w io.Writer, usage *types.FunctionUsage, verbose bool) {
	if !verbose && usage == nil {
		return
	}

	if usage == nil {
		fmt.Fprintln(w, "Usage:\t <none>")
		return
	}

	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  RAM:\t %.2f MB\n", (usage.TotalMemoryBytes / 1024 / 1024))
	cpu := usage.CPU
	if cpu < 0 {
		cpu = 1
	}
	fmt.Fprintf(w, "  CPU:\t %.0f Mi\n", (cpu))
}

func printMap(w io.Writer, name string, m map[string]string, verbose bool) {
	if !verbose && len(m) == 0 {
		return
	}

	if len(m) == 0 {
		fmt.Fprintf(w, "%s:\t <none>\n", name)
		return
	}

	fmt.Fprintf(w, "%s:\n", name)

	if name == "Environment" {
		orderedKeys := generateMapOrder(m)
		for _, keyName := range orderedKeys {
			fmt.Fprintln(w, "\t "+keyName+": "+m[keyName])
		}
		return
	}

	for key, value := range m {
		fmt.Fprintln(w, "\t "+key+": "+value)
	}

	return
}

func printList(w io.Writer, name string, data []string, verbose bool) {
	if !verbose && len(data) == 0 {
		return
	}

	if len(data) == 0 {
		fmt.Fprintf(w, "%s:\t <none>\n", name)
		return
	}

	fmt.Fprintf(w, "%s:\n", name)
	for _, value := range data {
		fmt.Fprintln(w, "\t - "+value)
	}

	return
}

func printResources(w io.Writer, name string, data *types.FunctionResources, verbose bool) {
	if !verbose && data == nil {
		return
	}

	fmt.Fprintf(w, name+":")

	if data == nil {
		fmt.Fprintln(w, "\t <none>")
		return
	}

	fmt.Fprintln(w, "\t CPU: "+data.CPU)
	fmt.Fprintln(w, "\t Memory: "+data.Memory)

	return
}

func isEmpty(a interface{}) bool {
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func generateMapOrder(m map[string]string) []string {

	var keyNames []string

	for keyName := range m {
		keyNames = append(keyNames, keyName)
	}

	sort.Strings(keyNames)

	return keyNames
}
