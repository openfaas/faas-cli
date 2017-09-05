// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/alexellis/faas-cli/proxy"
	"github.com/alexellis/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	verboseInvoke bool
	contentType   string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	invokeCmd.Flags().StringVar(&gateway, "gateway", "http://localhost:8080", "Gateway URI")
	invokeCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")

	invokeCmd.Flags().StringVar(&language, "lang", "node", "Programming language template")
	invokeCmd.Flags().StringVar(&contentType, "content-type", "text/plain", "The content-type HTTP header such as application/json")
	invokeCmd.Flags().BoolVar(&verboseInvoke, "verbose", false, "Verbose output for the function list")

	faasCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use: `invoke --gateway GATEWAY_URL
  faas-cli invoke [--gateway GATEWAY_URL] [--content-type CONTENT_TYPE] STDIN`,

	Short: "Invoke an OpenFaaS function",
	Long:  `Invokes an OpenFaaS function and reads from STDIN for the body of the request`,
	Example: `  faas-cli invoke --gateway https://domain:port --name echo
  faas-cli invoke --gateway https://domain:port --name echo --content-type application/json`,
	Run: runInvoke,
}

func runInvoke(cmd *cobra.Command, args []string) {
	var services stack.Services

	if len(functionName) == 0 {
		fmt.Println("Give a function to invoke via --name")
		return
	}

	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAML(yamlFile)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		if parsedServices != nil {
			services = *parsedServices
			gateway = services.Provider.GatewayURL
		}
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
	}

	functionInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Unable to read standard input: %s\n", err.Error())
		return
	}

	response, err := proxy.InvokeFunction(gateway, functionName, &functionInput, contentType)
	if err != nil {
		fmt.Println(err)
		return
	}
	if response != nil {
		os.Stdout.Write(*response)
	}
}
