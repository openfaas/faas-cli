// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	readyCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")

	readyCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")

	readyCmd.Flags().Int("attempts", 60, "Number of attempts to check the gateway")
	readyCmd.Flags().Duration("interval", time.Second*1, "Interval between attempts in seconds")

	faasCmd.AddCommand(readyCmd)
}

var readyCmd = &cobra.Command{
	Use:   `ready [--gateway GATEWAY_URL] [--tls-no-verify] [FUNCTION_NAME]`,
	Short: "Block until the gateway or a function is ready for use",
	Example: `  # Block until the gateway is ready
  faas-cli ready --gateway https://127.0.0.1:8080

  # Block until the env function is ready
  faas-cli store deploy env && \
  faas-cli ready env`,
	RunE: runReadyCmd,
}

func runReadyCmd(cmd *cobra.Command, args []string) error {
	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return err
	}

	attempts, err := cmd.Flags().GetInt("attempts")
	if err != nil {
		return err
	}

	if attempts < 1 {
		return fmt.Errorf("attempts must be greater than 0")
	}

	var services stack.Services
	var gatewayAddress string
	var yamlGateway string
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
	gatewayAddress = getGatewayURL(gateway, defaultGateway, yamlGateway, os.Getenv(openFaaSURLEnvironment))
	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)

	if len(args) == 0 {
		ready := false

		c := &http.Client{
			Transport: transport,
		}

		u, err := url.Parse(gatewayAddress)
		if err != nil {
			return err
		}

		u.Path = "/healthz"

		for i := 0; i < attempts; i++ {
			fmt.Printf("[%d/%d] Waiting for gateway\n", i+1, attempts)
			req, err := http.NewRequest(http.MethodGet, u.String(), nil)
			if err != nil {
				return err
			}

			res, err := c.Do(req)
			if err != nil {
				fmt.Printf("[%d/%d] Error reaching OpenFaaS gateway: %s\n", i+1, attempts, err.Error())
			} else if res.StatusCode == http.StatusOK {
				fmt.Printf("OpenFaaS gateway is ready\n")
				ready = true
				break
			}

			time.Sleep(interval)
		}

		if !ready {
			return fmt.Errorf("gateway: %s not ready after: %s", gatewayAddress, interval*time.Duration(attempts).Round(time.Second))
		}

	} else {
		functionName := args[0]
		ready := false
		cliAuth, err := proxy.NewCLIAuth(token, gatewayAddress)
		if err != nil {
			return err
		}

		cliClient, err := proxy.NewClient(cliAuth, gatewayAddress, transport, &commandTimeout)
		if err != nil {
			return err
		}

		ctx := context.Background()

		for i := 0; i < attempts; i++ {
			fmt.Printf("[%d/%d] Waiting for function %s\n", i+1, attempts, functionName)

			function, err := cliClient.GetFunctionInfo(ctx, functionName, functionNamespace)
			if err != nil {
				fmt.Printf("[%d/%d] Error getting function info: %s\n", i+1, attempts, err.Error())
			}

			if function.AvailableReplicas > 0 {
				fmt.Printf("Function %s is ready\n", functionName)
				ready = true
				break
			}
			time.Sleep(interval)
		}

		if !ready {
			return fmt.Errorf("function %s not ready after: %s", functionName, interval*time.Duration(attempts).Round(time.Second))
		}

	}

	return nil
}
