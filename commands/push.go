// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package commands

import (
	"fmt"
	"sync"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

func init() {
	faasCmd.AddCommand(pushCmd)

	pushCmd.Flags().IntVar(&parallel, "parallel", 1, "Push images in parallel to depth specified.")
}

// pushCmd handles pushing function container images to a remote repo
var pushCmd = &cobra.Command{
	Use:   `push -f YAML_FILE [--regex "REGEX"] [--filter "WILDCARD"] [--parallel]`,
	Short: "Push OpenFaaS functions to remote registry (Docker Hub)",
	Long: `Pushes the OpenFaaS function container image(s) defined in the supplied YAML
config to a remote repository.

These container images must already be present in your local image cache.`,

	Example: `  faas-cli push -f https://domain/path/myfunctions.yml
  faas-cli push -f ./samples.yml
  faas-cli push -f ./samples.yml --parallel 4
  faas-cli push -f ./samples.yml --filter "*gif*"
  faas-cli push -f ./samples.yml --regex "fn[0-9]_.*"`,
	RunE: runPush,
}

func runPush(cmd *cobra.Command, args []string) error {

	var services stack.Services
	if len(yamlFile) > 0 {
		parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter)
		if err != nil {
			return err
		}

		if parsedServices != nil {
			services = *parsedServices
		}
	}

	if len(services.Functions) > 0 {
		pushStack(&services, parallel)
	} else {
		return fmt.Errorf("you must supply a valid YAML file")
	}
	return nil
}

func pushImage(image string) {
	builder.ExecCommand("./", []string{"docker", "push", image})
}

func pushStack(services *stack.Services, queueDepth int) {
	wg := sync.WaitGroup{}

	workChannel := make(chan stack.Function)

	for i := 0; i < queueDepth; i++ {

		go func(index int) {
			wg.Add(1)
			for function := range workChannel {
				fmt.Printf("[%d] > Pushing: %s.\n", index, function.Name)
				if len(function.Image) == 0 {
					fmt.Println("Please provide a valid Image value in the YAML file.")
				} else {
					pushImage(function.Image)
				}
			}

			fmt.Printf("[%d] < Pushing done.\n", index)
			wg.Done()
		}(i)
	}

	for k, function := range services.Functions {
		function.Name = k
		workChannel <- function
	}

	close(workChannel)

	wg.Wait()

}
