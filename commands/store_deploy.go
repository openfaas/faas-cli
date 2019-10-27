// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	storeDeployCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	storeDeployCmd.Flags().StringVar(&network, "network", "", "Name of the network")
	storeDeployCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function (overriding name from the store)")
	storeDeployCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", "", "Namespace of the function")
	// Setup flags that are used only by deploy command (variables defined above)
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.envvarOpts, "env", "e", []string{}, "Adds one or more environment variables to the defined ones by store (ENVVAR=VALUE)")
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.labelOpts, "label", "l", []string{}, "Set one or more label (LABEL=VALUE)")
	storeDeployCmd.Flags().BoolVar(&storeDeployFlags.replace, "replace", false, "Replace any existing function")
	storeDeployCmd.Flags().BoolVar(&storeDeployFlags.update, "update", true, "Update existing functions")
	storeDeployCmd.Flags().StringArrayVar(&storeDeployFlags.constraints, "constraint", []string{}, "Apply a constraint to the function")
	storeDeployCmd.Flags().StringArrayVar(&storeDeployFlags.secrets, "secret", []string{}, "Give the function access to a secure secret")
	storeDeployCmd.Flags().BoolVarP(&storeDeployFlags.sendRegistryAuth, "send-registry-auth", "a", false, "send registryAuth from Docker credentials manager with the request")
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.annotationOpts, "annotation", "", []string{}, "Set one or more annotation (ANNOTATION=VALUE)")
	storeDeployCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	storeDeployCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")

	// Set bash-completion.
	_ = storeDeployCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	storeCmd.AddCommand(storeDeployCmd)
}

var storeDeployCmd = &cobra.Command{
	Use: `deploy (FUNCTION_NAME|FUNCTION_TITLE)
			[--name FUNCTION_NAME]
			[--gateway GATEWAY_URL]
			[--network NETWORK_NAME]
			[--env ENVVAR=VALUE ...]
			[--label LABEL=VALUE ...]
			[--annotation ANNOTATION=VALUE ...]
			[--replace=false]
			[--update=true]
			[--constraint PLACEMENT_CONSTRAINT ...]
			[--secret "SECRET_NAME"]
			[--url STORE_URL]
			[--tls-no-verify=false]`,

	Short: "Deploy OpenFaaS functions from a store",
	Long:  `Same as faas-cli deploy except that function is pre-loaded with arguments from the store`,
	Example: `  faas-cli store deploy figlet
  faas-cli store deploy figlet \
    --gateway=http://127.0.0.1:8080 \
    --env=MYVAR=myval`,
	RunE: runStoreDeploy,
}

func runStoreDeploy(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	targetPlatform := getTargetPlatform(platformValue)
	storeItems, err := storeList(storeAddress)
	if err != nil {
		return err
	}

	platformFunctions := filterStoreList(storeItems, targetPlatform)

	requestedStoreFn := args[0]
	item := storeFindFunction(requestedStoreFn, platformFunctions)
	if item == nil {
		return fmt.Errorf("function '%s' not found for platform '%s'", requestedStoreFn, targetPlatform)
	}

	// Add the store environment variables to the provided ones from cmd
	if item.Environment != nil {
		for k, v := range item.Environment {
			env := fmt.Sprintf("%s=%s", k, v)
			storeDeployFlags.envvarOpts = append(storeDeployFlags.envvarOpts, env)
		}
	}

	// Add the store labels to the provided ones from cmd
	if item.Labels != nil {
		for k, v := range item.Labels {
			label := fmt.Sprintf("%s=%s", k, v)
			storeDeployFlags.labelOpts = append(storeDeployFlags.labelOpts, label)
		}
	}

	if item.Annotations != nil {
		for k, v := range item.Annotations {
			annotation := fmt.Sprintf("%s=%s", k, v)
			storeDeployFlags.annotationOpts = append(storeDeployFlags.annotationOpts, annotation)
		}
	}

	// Use the network from manifest if not changed by user
	if !cmd.Flag("network").Changed {
		network = item.Network
	}

	itemName := item.Name

	if functionName != "" {
		itemName = functionName
	}

	var registryAuth string
	imageName := item.GetImageName(targetPlatform)

	if storeDeployFlags.sendRegistryAuth {

		dockerConfig := configFile{}
		err := readDockerConfig(&dockerConfig)
		if err != nil {
			log.Printf("Unable to read the docker config - %v\n", err.Error())
		}

		registryAuth = getRegistryAuth(&dockerConfig, imageName)
	}

	gateway = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	cliAuth := NewCLIAuth(token, gateway)
	transport := GetDefaultCLITransport(tlsInsecure, &commandTimeout)
	proxyClient := proxy.NewClient(cliAuth, gateway, transport, &commandTimeout)

	statusCode, err := deployImage(context.Background(), proxyClient, imageName, item.Fprocess, itemName, registryAuth, storeDeployFlags,
		tlsInsecure, item.ReadOnlyRootFilesystem, token, functionNamespace)

	if badStatusCode(statusCode) {
		failedStatusCode := map[string]int{itemName: statusCode}
		err := deployFailed(failedStatusCode)
		return err
	}

	return err
}
