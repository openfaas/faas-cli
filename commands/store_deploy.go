// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/openfaas/faas-cli/util"

	"github.com/openfaas/faas-cli/proxy"
	"github.com/spf13/cobra"
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	storeDeployCmd.Flags().StringVarP(&gateway, "gateway", "g", defaultGateway, "Gateway URL starting with http(s)://")
	storeDeployCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function (overriding name from the store)")
	storeDeployCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", "", "Namespace of the function")
	// Setup flags that are used only by deploy command (variables defined above)
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.envvarOpts, "env", "e", []string{}, "Adds one or more environment variables to the defined ones by store (ENVVAR=VALUE)")
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.labelOpts, "label", "l", []string{}, "Set one or more label (LABEL=VALUE)")
	storeDeployCmd.Flags().BoolVar(&storeDeployFlags.replace, "replace", false, "Replace any existing function")
	storeDeployCmd.Flags().BoolVar(&storeDeployFlags.update, "update", true, "Update existing functions")
	storeDeployCmd.Flags().StringArrayVar(&storeDeployFlags.constraints, "constraint", []string{}, "Apply a constraint to the function")
	storeDeployCmd.Flags().StringArrayVar(&storeDeployFlags.secrets, "secret", []string{}, "Give the function access to a secure secret")
	storeDeployCmd.Flags().StringArrayVarP(&storeDeployFlags.annotationOpts, "annotation", "", []string{}, "Set one or more annotation (ANNOTATION=VALUE)")
	storeDeployCmd.Flags().BoolVar(&tlsInsecure, "tls-no-verify", false, "Disable TLS validation")
	storeDeployCmd.Flags().StringVarP(&token, "token", "k", "", "Pass a JWT token to use instead of basic auth")
	storeDeployCmd.Flags().DurationVar(&timeoutOverride, "timeout", commandTimeout, "Timeout for any HTTP calls made to the OpenFaaS API.")

	// Set bash-completion.
	_ = storeDeployCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	storeCmd.AddCommand(storeDeployCmd)
}

var storeDeployCmd = &cobra.Command{
	Use: `deploy (FUNCTION_NAME|FUNCTION_TITLE)
			[--name FUNCTION_NAME]
			[--gateway GATEWAY_URL]
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
	RunE:    runStoreDeploy,
	PreRunE: preRunEStoreDeploy,
}

func preRunEStoreDeploy(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide the function name")
	}

	return nil
}

func runStoreDeploy(cmd *cobra.Command, args []string) error {
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

	flagEnvs, err := util.ParseMap(storeDeployFlags.envvarOpts, "env")
	if err != nil {
		return err
	}

	// Add the store environment variables to the provided ones from cmd
	mergedEnvs := util.MergeMap(item.Environment, flagEnvs)

	envs := []string{}
	for k, v := range mergedEnvs {
		env := fmt.Sprintf("%s=%s", k, v)
		envs = append(envs, env)
	}

	storeDeployFlags.envvarOpts = envs

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

	itemName := item.Name

	if functionName != "" {
		itemName = functionName
	}

	imageName := item.GetImageName(targetPlatform)

	gateway = getGatewayURL(gateway, defaultGateway, "", os.Getenv(openFaaSURLEnvironment))
	cliAuth, err := proxy.NewCLIAuth(token, gateway)
	if err != nil {
		return err
	}
	transport := GetDefaultCLITransport(tlsInsecure, &timeoutOverride)
	proxyClient, err := proxy.NewClient(cliAuth, gateway, transport, &timeoutOverride)
	if err != nil {
		return err
	}

	statusCode, err := deployImage(context.Background(), proxyClient, imageName, item.Fprocess, itemName, "", storeDeployFlags,
		tlsInsecure, item.ReadOnlyRootFilesystem, token, functionNamespace)

	if badStatusCode(statusCode) {
		failedStatusCode := map[string]int{itemName: statusCode}
		err := deployFailed(failedStatusCode)
		return err
	}

	return err
}
