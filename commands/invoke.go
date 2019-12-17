// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alexellis/hmac"
	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

var (
	contentType             string
	query                   []string
	headers                 []string
	invokeAsync             bool
	httpMethod              string
	sigHeader               string
	key                     string
	functionInvokeNamespace string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	invokeCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")
	invokeCmd.Flags().StringVarP(&functionInvokeNamespace, "namespace", "n", "", "Namespace of the deployed function")

	invokeCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")

	invokeCmd.Flags().StringVar(&contentType, "content-type", "text/plain", "The content-type HTTP header such as application/json")
	invokeCmd.Flags().StringArrayVar(&query, "query", []string{}, "pass query-string options")
	invokeCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "pass HTTP request header")
	invokeCmd.Flags().BoolVarP(&invokeAsync, "async", "a", false, "Invoke the function asynchronously")
	invokeCmd.Flags().StringVarP(&httpMethod, "method", "m", "POST", "pass HTTP request method")
	invokeCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	invokeCmd.Flags().StringVar(&sigHeader, "sign", "", "name of HTTP request header to hold the signature")
	invokeCmd.Flags().StringVar(&key, "key", "", "key to be used to sign the request (must be used with --sign)")

	invokeCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")

	faasCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   `invoke FUNCTION_NAME [--gateway GATEWAY_URL] [--content-type CONTENT_TYPE] [--query PARAM=VALUE] [--header PARAM=VALUE] [--method HTTP_METHOD]`,
	Short: "Invoke an OpenFaaS function",
	Long:  `Invokes an OpenFaaS function and reads from STDIN for the body of the request`,
	Example: `  faas-cli invoke echo --gateway https://domain:port
  faas-cli invoke echo --gateway https://domain:port --content-type application/json
  faas-cli invoke env --query repo=faas-cli --query org=openfaas
  faas-cli invoke env --header X-Ping-Url=http://request.bin/etc
  faas-cli invoke resize-img --async -H "X-Callback-Url=http://gateway:8080/function/send2slack" < image.png
  faas-cli invoke env -H X-Ping-Url=http://request.bin/etc
  faas-cli invoke flask --method GET --namespace dev
  faas-cli invoke env --sign X-GitHub-Event --key yoursecret`,
	RunE: runInvoke,
}

func runInvoke(cmd *cobra.Command, args []string) error {
	var services stack.Services

	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}

	if missingSignFlag(sigHeader, key) {
		return fmt.Errorf("signing requires both --sign <header-value> and --key <key-value>")
	}

	var yamlGateway string
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

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
	}

	functionInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read standard input: %s", err.Error())
	}

	if len(sigHeader) > 0 {
		signedHeader, err := generateSignedHeader(functionInput, key, sigHeader)
		if err != nil {
			return fmt.Errorf("unable to sign message: %s", err.Error())
		}
		headers = append(headers, signedHeader)
	}

	response, err := proxy.InvokeFunction(gatewayAddress, functionName, &functionInput, contentType, query, headers, invokeAsync, httpMethod, tlsInsecure, functionInvokeNamespace)
	if err != nil {
		return err
	}

	if response != nil {
		os.Stdout.Write(*response)
	}

	return nil
}

func generateSignedHeader(message []byte, key string, headerName string) (string, error) {

	if len(headerName) == 0 {
		return "", fmt.Errorf("signed header must have a non-zero length")
	}

	hash := hmac.Sign(message, []byte(key))
	signature := hex.EncodeToString(hash)
	signedHeader := fmt.Sprintf(`%s=%s=%s`, headerName, "sha1", string(signature[:]))

	return signedHeader, nil
}

func missingSignFlag(header string, key string) bool {
	return (len(header) > 0 && len(key) == 0) || (len(header) == 0 && len(key) > 0)
}
