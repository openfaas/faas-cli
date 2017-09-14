// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexellis/faas-cli/proxy"
	"github.com/alexellis/faas-cli/stack"
	"github.com/spf13/cobra"
)

// Flags that are to be added to commands.

var (
	envvarOpts  []string
	replace     bool
	constraints []string
)

func init() {
	// Setup flags that are used by multiple commands (variables defined in faas.go)
	deployCmd.Flags().StringVar(&fprocess, "fprocess", "", "Fprocess to be run by the watchdog")
	deployCmd.Flags().StringVar(&gateway, "gateway", defaultGateway, "Gateway URI")
	deployCmd.Flags().StringVar(&handler, "handler", "", "Directory with handler for function, e.g. handler.js")
	deployCmd.Flags().StringVar(&image, "image", "", "Docker image name to build")
	deployCmd.Flags().StringVar(&language, "lang", "node", "Programming language template")
	deployCmd.Flags().StringVar(&functionName, "name", "", "Name of the deployed function")

	// Setup flags that are used only by this command (variables defined above)
	deployCmd.Flags().StringArrayVarP(&envvarOpts, "env", "e", []string{}, "Set one or more environment variables (ENVVAR=VALUE)")
	deployCmd.Flags().BoolVar(&replace, "replace", true, "Replace any existing function")

	deployCmd.Flags().StringArrayVar(&constraints, "constraint", []string{}, "Apply a constraint to the function")

	// Set bash-completion.
	_ = deployCmd.Flags().SetAnnotation("handler", cobra.BashCompSubdirsInDir, []string{})

	faasCmd.AddCommand(deployCmd)
}

// deployCmd handles deploying OpenFaaS function containers
var deployCmd = &cobra.Command{
	Use: `deploy -f YAML_FILE [--replace=false]
  faas-cli deploy --image IMAGE_NAME
                  --name FUNCTION_NAME
                  [--lang <ruby|python|node|csharp>]
                  [--gateway GATEWAY_URL]
                  [--handler HANDLER_DIR]
                  [--fprocess PROCESS]
                  [--env ENVVAR=VALUE ...]
                  [--replace=false]
                  [--constraint PLACEMENT_CONSTRAINT ...]
                  [--regex "REGEX"]
                  [--filter "WILDCARD"]`,

	Short: "Deploy OpenFaaS functions",
	Long: `Deploys OpenFaaS function containers either via the supplied YAML config using
the "--yaml" flag (which may contain multiple function definitions), or directly
via flags.`,
	Example: `  faas-cli deploy -f https://domain/path/myfunctions.yml
  faas-cli deploy -f ./samples.yml
  faas-cli deploy -f ./samples.yml --filter "*gif*"
  faas-cli deploy -f ./samples.yml --regex "fn[0-9]_.*"
  faas-cli deploy -f ./samples.yml --replace=false
  faas-cli deploy --image=alexellis/faas-url-ping --name=url-ping
  faas-cli deploy --image=my_image --name=my_fn --handler=/path/to/fn/
                  --gateway=http://remote-site.com:8080 --lang=python
                  --env=MYVAR=myval`,
	Run: runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) {
	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		// Override gateway if passed
		if len(gateway) > 0 && gateway != defaultGateway {
			parsedServices.Provider.GatewayURL = gateway
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if len(services.Functions) > 0 {
		if len(services.Provider.Network) == 0 {
			services.Provider.Network = defaultNetwork
		}

		for k, function := range services.Functions {
			function.Name = k
			fmt.Printf("Deploying: %s.\n", function.Name)
			if function.Constraints != nil {
				constraints = *function.Constraints
			}

			proxy.DeployFunction(function.FProcess, services.Provider.GatewayURL, function.Name, function.Image, function.Language, replace, function.Environment, services.Provider.Network, constraints)
		}
	} else {
		if len(image) == 0 {
			fmt.Println("Please provide a --image to be deployed.")
			return
		}
		if len(functionName) == 0 {
			fmt.Println("Please provide a --name for your function as it will be deployed on FaaS")
			return
		}

		envvars, err := parseEnvvars(envvarOpts)
		if err != nil {
			fmt.Printf("Error parsing envvars: %v\n", err)
			os.Exit(1)
		}

		proxy.DeployFunction(fprocess, gateway, functionName, image, language, replace, envvars, defaultNetwork, constraints)
	}
}

func parseEnvvars(envvars []string) (map[string]string, error) {
	result := map[string]string{}
	for _, envvar := range envvars {
		s := strings.SplitN(strings.TrimSpace(envvar), "=", 2)
		envvarName := s[0]
		envvarValue := s[1]

		if !(len(envvarName) > 0) {
			return nil, fmt.Errorf("Empty envvar name: [%s]", envvar)
		}
		if !(len(envvarValue) > 0) {
			return nil, fmt.Errorf("Empty envvar value: [%s]", envvar)
		}

		result[envvarName] = envvarValue
	}
	return result, nil
}
