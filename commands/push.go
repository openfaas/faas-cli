// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"log"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(pushCmd)
}

// pushCmd handles pushing function container images to a remote repo
var pushCmd = &cobra.Command{
	Use:   `push -f YAML_FILE [--regex "REGEX"] [--filter "WILDCARD"]`,
	Short: "Push OpenFaaS functions to remote registry (Docker Hub)",
	Long: `Pushes the OpenFaaS function container image(s) defined in the supplied YAML
config to a remote repository.

These container images must already be present in your local image cache.

NOTE - this command currently supports pushing to docker hub only, support for
       additional container repos is planned.`,

	Example: `  faas-cli push -f https://domain/path/myfunctions.yml
  faas-cli push -f ./samples.yml
  faas-cli push -f ./samples.yml --filter "*gif*"
  faas-cli push -f ./samples.yml --regex "fn[0-9]_.*"`,
	Run: runPush,
}

func runPush(cmd *cobra.Command, args []string) {

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if len(services.Functions) > 0 {
		for k, function := range services.Functions {
			function.Name = k
			fmt.Printf("Pushing: %s to remote repository.\n", function.Name)
			pushImage(function.Image)
		}
	} else {
		fmt.Println("You must supply a valid YAML file.")
		return
	}

}

func pushImage(image string) {
	builder.ExecCommand("./", []string{"docker", "push", image})
}
