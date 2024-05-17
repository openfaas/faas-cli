// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/alexellis/hmac"
	"github.com/openfaas/faas-cli/version"
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
	Example: `  faas-cli invoke echo --gateway https://host:port
  faas-cli invoke echo --gateway https://host:port --content-type application/json
  faas-cli invoke env --query repo=faas-cli --query org=openfaas
  faas-cli invoke env --header X-Ping-Url=http://request.bin/etc
  faas-cli invoke resize-img --async -H "X-Callback-Url=http://gateway:8080/function/send2slack" < image.png
  faas-cli invoke env -H X-Ping-Url=http://request.bin/etc
  faas-cli invoke flask --method GET --namespace dev
  faas-cli invoke env --sign X-GitHub-Event --key yoursecret`,
	RunE: runInvoke,
}

func runInvoke(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide a name for the function")
	}
	functionName = args[0]

	if missingSignFlag(sigHeader, key) {
		return fmt.Errorf("signing requires both --sign <header-value> and --key <key-value>")
	}

	err := validateHTTPMethod(httpMethod)
	if err != nil {
		return nil
	}

	httpHeader, err := parseHeaders(headers)
	if err != nil {
		return err
	}

	httpQuery, err := parseQueryValues(query)
	if err != nil {
		return err
	}

	httpHeader.Set("Content-Type", contentType)
	httpHeader.Set("User-Agent", fmt.Sprintf("faas-cli/%s (openfaas; %s; %s)", version.BuildVersion(), runtime.GOOS, runtime.GOARCH))

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintf(os.Stderr, "Reading from STDIN - hit (Control + D) to stop.\n")
	}

	functionInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read standard input: %s", err.Error())
	}

	if len(sigHeader) > 0 {
		sig := generateSignature(functionInput, key)
		httpHeader.Add(sigHeader, sig)
	}

	client, err := GetDefaultSDKClient()
	if err != nil {
		return err
	}

	u, _ := url.Parse("/")
	u.RawQuery = httpQuery.Encode()

	body := bytes.NewReader(functionInput)
	req, err := http.NewRequest(httpMethod, u.String(), body)
	if err != nil {
		return err
	}
	req.Header = httpHeader

	authenticate := false
	res, err := client.InvokeFunction(functionName, functionInvokeNamespace, invokeAsync, authenticate, req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s", client.GatewayURL)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if code := res.StatusCode; code < 200 || code > 299 {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("cannot read result from OpenFaaS on URL: %s %s", gateway, err)
		}

		return fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(resBody))
	}

	if invokeAsync && res.StatusCode == http.StatusAccepted {
		fmt.Fprintf(os.Stderr, "Function submitted asynchronously.\n")
		return nil
	}

	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		return fmt.Errorf("cannot read result from OpenFaaS on URL: %s %s", gateway, err)
	}

	return nil
}

func generateSignature(message []byte, key string) string {
	hash := hmac.Sign(message, []byte(key))
	signature := hex.EncodeToString(hash)

	return fmt.Sprintf(`%s=%s`, "sha1", string(signature[:]))
}

func missingSignFlag(header string, key string) bool {
	return (len(header) > 0 && len(key) == 0) || (len(header) == 0 && len(key) > 0)
}

// parseHeaders parses header values from the header command flag
func parseHeaders(headers []string) (http.Header, error) {
	httpHeader := http.Header{}

	for _, header := range headers {
		headerVal := strings.SplitN(header, "=", 2)
		if len(headerVal) != 2 {
			return httpHeader, fmt.Errorf("the --header or -H flag must take the form of key=value")
		}

		key, value := headerVal[0], headerVal[1]
		if key == "" {
			return httpHeader, fmt.Errorf("the --header or -H flag must take the form of key=value (empty key given)")
		}

		if value == "" {
			return httpHeader, fmt.Errorf("the --header or -H flag must take the form of key=value (empty value given)")
		}

		httpHeader.Add(key, value)
	}

	return httpHeader, nil
}

// parseQueryValues parses query values from the query command flags
func parseQueryValues(query []string) (url.Values, error) {
	v := url.Values{}

	for _, q := range query {
		queryVal := strings.SplitN(q, "=", 2)
		if len(queryVal) != 2 {
			return v, fmt.Errorf("the --query flag must take the form of key=value")
		}

		key, value := queryVal[0], queryVal[1]
		if key == "" {
			return v, fmt.Errorf("the --header or -H flag must take the form of key=value (empty key given)")
		}

		if value == "" {
			return v, fmt.Errorf("the --header or -H flag must take the form of key=value (empty value given)")
		}

		v.Add(key, value)
	}

	return v, nil
}

// validateMethod validates the HTTP request method
func validateHTTPMethod(httpMethod string) error {
	var allowedMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	}
	helpString := strings.Join(allowedMethods, "/")

	if !contains(allowedMethods, httpMethod) {
		return fmt.Errorf("the --method or -m flag must take one of these values (%s)", helpString)
	}
	return nil
}
